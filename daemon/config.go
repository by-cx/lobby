package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

// Config keeps info about configuration of this daemon
type Config struct {
	Token                string   `envconfig:"TOKEN" required:"false"`                                            // Authentication token, if empty auth is disabled
	Host                 string   `envconfig:"HOST" required:"false" default:"127.0.0.1"`                         // IP address used for the REST server to listen
	Port                 uint16   `envconfig:"PORT" required:"false" default:"1313"`                              // Port related to the address above
	NATSURL              string   `envconfig:"NATS_URL" required:"true"`                                          // NATS URL used to connect to the NATS server
	NATSDiscoveryChannel string   `envconfig:"NATS_DISCOVERY_CHANNEL" required:"false" default:"lobby.discovery"` // Channel where the kepp alive packets are sent
	Labels               []string `envconfig:"LABELS" required:"false" default:""`                                // List of labels
	LabelsPath           string   `envconfig:"LABELS_PATH" required:"false" default:"/etc/lobby/labels"`          // Path where filesystem based labels are located
	HostName             string   `envconfig:"HOSTNAME" required:"false"`                                         // Overrise local machine's hostname
	CleanEvery           uint     `envconfig:"CLEAN_EVERY" required:"false" default:"15"`                         // How often to clean the list of servers to get rid of the not alive ones
	KeepAlive            uint     `envconfig:"KEEP_ALIVE" required:"false" default:"5"`                           // how often to send the keepalive message with all availabel information [secs]
	TTL                  uint     `envconfig:"TTL" required:"false" default:"30"`                                 // After how many secs is discovery record considered as invalid
	NodeExporterPort     uint     `envconfig:"NODE_EXPORTER_PORT" required:"false" default:"9100"`                // Default port where node_exporter listens on all registered servers
	Register             bool     `envconfig:"REGISTER" required:"false" default:"true"`                          // If true (default) then local instance is registered with other instance (discovery packet is sent regularly)
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
