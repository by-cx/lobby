package main

import (
	"github.com/by-cx/lobby/server"
)

// These functions are called when something has changed in the storage

// discoveryChange is called when daemon detects that a newly arrived discovery
// packet is somehow different than the localone. This can be used to trigger
// some action in the local machine.
func discoveryChange(discovery server.Discovery) error {
	return nil
}
