package common

import "github.com/by-cx/lobby/server"

// Listener is a function that returns received discovery
type Listener func(server.Discovery)

// Driver interface describes exported methods that have to be implemented in each driver
type Driver interface {
	Init() error
	Close() error
	RegisterSubscribeFunction(listener Listener)
	RegisterUnsubscribeFunction(listener Listener)
	SendDiscoveryPacket(discovery server.Discovery) error
	SendGoodbyePacket(discovery server.Discovery) error
}
