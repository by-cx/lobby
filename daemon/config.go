package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config keeps info about configuration of this daemon
type Config struct {
	Token   string   `envconfig:"TOKEN" required:"false"` // not used yet
	NATSURL string   `envconfig:"NATS_URL" required:"true"`
	Labels  []string `envconfig:"LABELS" required:"false" default:""`
}

// GetConfig return configuration created based on environment variables
func GetConfig() *Config {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return &config
}
