package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/rosti-cz/server_lobby/server"
)

// discoveryHandler accepts discovery message and
func discoveryHandler(m *nats.Msg) {
	message := server.Discovery{}
	err := json.Unmarshal(m.Data, &message)
	if err != nil {
		log.Println(fmt.Errorf("decoding message error: %v", err))
	}

	err = message.Validate()
	if err != nil {
		log.Println(fmt.Errorf("validation error: %v", err))
	}

	discoveryStorage.Add(message)

}
