package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

// discoveryHandler accepts discovery message and
func discoveryHandler(m *nats.Msg) {
	message := discoveryEnvelope{}
	err := json.Unmarshal(m.Data, &message)
	if err != nil {
		log.Println(fmt.Errorf("decoding message error: %v", err))
	}

	err = message.Discovery.Validate()
	if err != nil {
		log.Println(fmt.Errorf("validation error: %v", err))
	}

	if message.Message == "hi" {
		discoveryStorage.Add(message.Discovery)
	} else if message.Message == "goodbye" {
		discoveryStorage.Delete(message.Discovery.Hostname)
	}
}
