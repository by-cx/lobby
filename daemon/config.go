package main

import (
	"log"

	"github.com/by-cx/lobby/server"
	"github.com/kelseyhightower/envconfig"
)

// Config keeps info about configuration of this daemon
type Config struct {
	Token                 string        `envconfig:"TOKEN" required:"false"`                                            // Authentication token, if empty auth is disabled
	Host                  string        `envconfig:"HOST" required:"false" default:"127.0.0.1"`                         // IP address used for the REST server to listen
	Port                  uint16        `envconfig:"PORT" required:"false" default:"1313"`                              // Port related to the address above
	DisableAPI            bool          `envconfig:"DISABLE_API" required:"false" default:"false"`                      // If true API interface won't start
	Driver                string        `envconfig:"DRIVER" required:"false" default:"NATS"`                            // Select driver to use to communicate with the group of nodes. The possible values are NATS and Redis
	NATSURL               string        `envconfig:"NATS_URL" required:"false"`                                         // NATS URL used to connect to the NATS server
	NATSDiscoveryChannel  string        `envconfig:"NATS_DISCOVERY_CHANNEL" required:"false" default:"lobby.discovery"` // Channel where the kepp alive packets are sent
	RedisHost             string        `envconfig:"REDIS_HOST" required:"false" default:"127.0.0.1"`                   // Redis host
	RedisPort             uint16        `envconfig:"REDIS_PORT" required:"false" default:"6379"`                        // Redis port
	RedisDB               uint          `envconfig:"REDIS_DB" required:"false" default:"0"`                             // Redis DB
	RedisChannel          string        `envconfig:"REDIS_CHANNEL" required:"false" default:"lobby:discovery"`          // Redis channel
	RedisPassword         string        `envconfig:"REDIS_PASSWORD" required:"false" default:""`                        // Redis password
	Labels                server.Labels `envconfig:"LABELS" required:"false" default:""`                                // List of labels
	LabelsPath            string        `envconfig:"LABELS_PATH" required:"false" default:"/etc/lobby/labels"`          // Path where filesystem based labels are located
	RuntimeLabelsFilename string        `envconfig:"RUNTIME_LABELS_FILENAME" required:"false" default:"_runtime"`       // Filename for file created in LabelsPath where runtime labels will be added
	HostName              string        `envconfig:"HOSTNAME" required:"false"`                                         // Overrise local machine's hostname
	CleanEvery            uint          `envconfig:"CLEAN_EVERY" required:"false" default:"15"`                         // How often to clean the list of servers to get rid of the not alive ones
	KeepAlive             uint          `envconfig:"KEEP_ALIVE" required:"false" default:"5"`                           // how often to send the keepalive message with all availabel information [secs]
	TTL                   uint          `envconfig:"TTL" required:"false" default:"30"`                                 // After how many secs is discovery record considered as invalid
	NodeExporterPort      uint          `envconfig:"NODE_EXPORTER_PORT" required:"false" default:"9100"`                // Default port where node_exporter listens on all registered servers
	Register              bool          `envconfig:"REGISTER" required:"false" default:"true"`                          // If true (default) then local instance is registered with other instance (discovery packet is sent regularly)
	Callback              string        `envconfig:"CALLBACK" required:"false" default:""`                              // path to a script that runs when the is a change in the labels database
	CallbackCooldown      uint          `envconfig:"CALLBACK_COOLDOWN" required:"false" default:"15"`                   // cooldown that prevents to run the config change script too many times in row
	CallbackFirstRunDelay uint          `envconfig:"CALLBACK_FIRST_RUN_DELAY" required:"false" default:"30"`            // Wait for this amount of seconds before callback is run for first time after fresh start of the daemon
}

// GetConfig return configuration created based on environment variables
func GetConfig() *Config {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	if config.Driver != "Redis" && config.Driver != "NATS" {
		log.Fatal("ERROR: the only supported drivers are Redis and NATS (default)")
	}

	if config.Driver == "NATS" && len(config.NATSURL) == 0 {
		log.Fatal("ERROR: NATS_URL cannot be empty when driver is set to NATS")
	}

	return &config
}
