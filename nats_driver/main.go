package nats_driver

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/by-cx/lobby/common"
	"github.com/by-cx/lobby/server"
	"github.com/nats-io/nats.go"
)

// NATS drivers is used to send discovery packet to other nodes into the group via NATS messenging protocol.
type Driver struct {
	NATSUrl              string
	NATSDiscoveryChannel string

	LogChannel chan string

	nc                  *nats.Conn
	subscribeListener   common.Listener
	unsubscribeListener common.Listener
}

// handler is called asynchronously so and because it cannot log directly to stderr there
// is channel called LogChannel that can be used to log error messages from this
func (d *Driver) handler(m *nats.Msg) {
	message := discoveryEnvelope{}
	err := json.Unmarshal(m.Data, &message)
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

func (d *Driver) Init() error {
	if d.LogChannel == nil {
		return fmt.Errorf("please initiate LogChannel variable")
	}

	nc, err := nats.Connect(d.NATSUrl)
	if err != nil {
		return err
	}
	d.nc = nc

	_, err = nc.Subscribe(d.NATSDiscoveryChannel, d.handler)
	if err != nil {
		return err
	}

	return nil
}

// Close is called when all is done.
func (d *Driver) Close() error {
	return d.nc.Drain()
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
	err = d.nc.Publish(d.NATSDiscoveryChannel, data)
	// In case the connection is down we will try to reconnect
	if err != nil && strings.Contains(err.Error(), "connection closed") {
		d.nc.Close()
		err = d.Init()
		if err != nil {
			return fmt.Errorf("sending discovery reconnect error: %v", err)
		}
	} else if err != nil {
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
	err = d.nc.Publish(d.NATSDiscoveryChannel, data)
	if err != nil {
		return fmt.Errorf("sending discovery error: %v", err)
	}
	return nil
}
