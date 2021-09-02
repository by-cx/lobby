package nats_driver

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/rosti-cz/server_lobby/common"
	"github.com/rosti-cz/server_lobby/server"
)

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

func (d *Driver) Close() error {
	return d.nc.Drain()
}

func (d *Driver) RegisterSubscribeFunction(listener common.Listener) {
	d.subscribeListener = listener
}

func (d *Driver) RegisterUnsubscribeFunction(listener common.Listener) {
	d.unsubscribeListener = listener
}

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
	if err != nil {
		return fmt.Errorf("sending discovery error: %v", err)
	}
	return nil
}

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
