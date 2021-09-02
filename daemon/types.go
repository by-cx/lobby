package main

import (
	"encoding/json"

	"github.com/rosti-cz/server_lobby/server"
)

// discoveryEnvelope adds a message to the standard discovery format. The message
// can be "hi" or "goodbye" where "hi" is used when the node is sending keep alive
// packets and "goodbye" means the node is leaving.
type discoveryEnvelope struct {
	Discovery server.Discovery `json:"discovery"`
	Message   string           `json:"message"` // can be hi or goodbye
}

func (e *discoveryEnvelope) Bytes() ([]byte, error) {
	body, err := json.Marshal(e)
	return body, err
}
