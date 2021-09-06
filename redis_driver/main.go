package redis_driver

import (
	"encoding/json"

	"github.com/by-cx/lobby/common"
	"github.com/by-cx/lobby/server"
	"github.com/go-redis/redis"

	"fmt"
)

// Redis drivers is used to send discovery packet to other nodes into the group via Redis's pubsub protocol.
type Driver struct {
	Host     string
	Port     uint
	Password string
	Channel  string
	DB       uint

	LogChannel chan string

	subscribeListener   common.Listener
	unsubscribeListener common.Listener

	redis *redis.Client
}

// handler is called asynchronously so and because it cannot log directly to stderr there
// is channel called LogChannel that can be used to log error messages from this
func (d *Driver) handler(payload string) {
	message := discoveryEnvelope{}
	err := json.Unmarshal([]byte(payload), &message)
	if err != nil && d.LogChannel != nil {
		d.LogChannel <- fmt.Errorf("decoding message error: %v", err).Error()
	}

	err = message.Discovery.Validate()
	if err != nil && d.LogChannel != nil {
		d.LogChannel <- fmt.Errorf("validation error: %v", err).Error()
	}

	if message.Message == "hi" {
		d.subscribeListener(message.Discovery)
	} else if message.Message == "goodbye" {
		d.unsubscribeListener(message.Discovery)
	} else {
		if d.LogChannel != nil {
			d.LogChannel <- "incompatible message"
		}
	}
}

// Init connect to Redis, subscribes to the channel and waits for the messages.
// It runs goroutine in background that listens for new messages.
func (d *Driver) Init() error {
	if d.LogChannel == nil {
		return fmt.Errorf("please initiate LogChannel variable")
	}

	if len(d.Host) == 0 {
		return fmt.Errorf("parameter Host cannot be empty")
	}

	if len(d.Channel) == 0 {
		return fmt.Errorf("pattern cannot be empty")
	}

	if d.Port <= 0 || d.Port > 65536 {
		return fmt.Errorf("port has to be in range of 0-65536")
	}

	d.redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", d.Host, d.Port),
		Password: d.Password,
		DB:       int(d.DB),
	})

	pubsub := d.redis.Subscribe(d.Channel)

	go func(pubsub *redis.PubSub, d *Driver) {
		channel := pubsub.Channel()

		for message := range channel {
			d.handler(message.Payload)
		}
	}(pubsub, d)

	return nil
}

// Close is called when all is done.
func (d *Driver) Close() error {
	return d.redis.Close()
}

// RegisterSubscribeFunction sets the function that will process the incoming messages
func (d *Driver) RegisterSubscribeFunction(listener common.Listener) {
	d.subscribeListener = listener
}

// RegisterUnsubscribeFunction sets the function that will process the goodbye incoming messages
func (d *Driver) RegisterUnsubscribeFunction(listener common.Listener) {
	d.unsubscribeListener = listener
}

// SendDiscoveryPacket send discovery packet to the group.
func (d *Driver) SendDiscoveryPacket(discovery server.Discovery) error {
	envelope := discoveryEnvelope{
		Discovery: discovery,
		Message:   "hi",
	}

	data, err := envelope.Bytes()
	if err != nil {
		return fmt.Errorf("sending discovery formating message error: %v", err)
	}

	err = d.redis.Publish(d.Channel, data).Err()
	if err != nil {
		return fmt.Errorf("sending discovery error: %v", err)
	}
	return nil
}

// SendGoodbyePacket deregister node from the group. It tells everybody that it's going to die.
func (d *Driver) SendGoodbyePacket(discovery server.Discovery) error {
	envelope := discoveryEnvelope{
		Discovery: discovery,
		Message:   "goodbye",
	}

	data, err := envelope.Bytes()
	if err != nil {
		return fmt.Errorf("sending discovery formating message error: %v", err)
	}
	err = d.redis.Publish(d.Channel, data).Err()
	if err != nil {
		return fmt.Errorf("sending discovery error: %v", err)
	}
	return nil
}
