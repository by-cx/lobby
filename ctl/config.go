package main

import (
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config keeps info about configuration of this daemon
type Config struct {
	Token string `envconfig:"TOKEN" required:"false"`                    // Authentication token, if empty auth is disabled
	Proto string `envconfig:"PROTOCOL" required:"false" default:"http"`  // selected http or https protocols, default is http
	Host  string `envconfig:"HOST" required:"false" default:"127.0.0.1"` // IP address or hostname where lobbyd is listening
	Port  uint   `envconfig:"PORT" required:"false" default:"1313"`      // Same thing but the port part
}

// GetConfig return configuration created based on environment variables
func GetConfig() *Config {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return &config
}
