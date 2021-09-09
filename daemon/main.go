package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/by-cx/lobby/common"
	"github.com/by-cx/lobby/nats_driver"
	"github.com/by-cx/lobby/redis_driver"
	"github.com/by-cx/lobby/server"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var discoveryStorage server.Discoveries = server.Discoveries{}
var driver common.Driver
var localHost server.LocalHost

var config Config

var shuttingDown bool
var sendDiscoveryPacketTrigger chan bool = make(chan bool)

func init() {
	// Load config from environment variables
	config = *GetConfig()

	// Setup discovery storage
	discoveryStorage.LogChannel = make(chan string)
	discoveryStorage.TTL = config.TTL

	// localhost initization
	localHost = server.LocalHost{
		LabelsPath:            config.LabelsPath,
		HostnameOverride:      config.HostName,
		InitialLabels:         config.Labels,
		RuntimeLabelsFilename: config.RuntimeLabelsFilename,
	}

	// Setup driver
	if config.Driver == "NATS" {
		driver = &nats_driver.Driver{
			NATSUrl:              config.NATSURL,
			NATSDiscoveryChannel: config.NATSDiscoveryChannel,

			LogChannel: discoveryStorage.LogChannel,
		}
	} else if config.Driver == "Redis" {
		driver = &redis_driver.Driver{
			Host:     config.RedisHost,
			Port:     uint(config.RedisPort),
			Password: config.RedisPassword,
			Channel:  config.RedisChannel,
			DB:       uint(config.RedisDB),

			LogChannel: discoveryStorage.LogChannel,
		}
	} else {
		log.Fatalf("unsupported driver %s", config.Driver)
	}

}

// cleanDiscoveryPool clears the local server map and keeps only the alive servers
func cleanDiscoveryPool() {
	for {
		discoveryStorage.Clean()
		time.Sleep(time.Duration(config.CleanEvery) * time.Second)
	}

}

// sendGoodbyePacket is almost same as sendDiscoveryPacket but it's not running in loop
// and it adds goodbye message so other nodes know this node is gonna die.
func sendGoodbyePacket() {
	discovery, err := localHost.GetIdentification()
	if err != nil {
		log.Printf("sending discovery identification error: %v\n", err)
	}

	err = driver.SendGoodbyePacket(discovery)
	if err != nil {
		log.Println(err)
	}
}

// sendDiscoveryPacket sends a single discovery packet out
func sendDiscoveryPacket() {
	sendDiscoveryPacketTrigger <- true
}

// sendDisoveryPacket sends discovery packet to the driver which passes it to the
// other nodes. By this it propagates any change that happens in the local discovery struct.
// Every tune trigger is triggered it sends one message.
func sendDiscoveryPacketTask(trigger chan bool) {
	for {
		// We are waiting for the trigger
		<-trigger

		if !shuttingDown {
			discovery, err := localHost.GetIdentification()
			if err != nil {
				log.Printf("sending discovery identification error: %v\n", err)
			}

			err = driver.SendDiscoveryPacket(discovery)
			if err != nil {
				log.Println(err.Error())
			}
		} else {
			break
		}
	}
}

// Print logs acquired from disovery storage
func printDiscoveryLogs() {
	for {
		logMessage := <-discoveryStorage.LogChannel
		log.Println(logMessage)
	}
}

func main() {
	var err error

	// Closing the logging channel
	defer close(discoveryStorage.LogChannel)
	defer driver.Close()

	// ------------------------
	// Server discovering stuff
	// ------------------------

	// Connect to the NATS service
	driver.RegisterSubscribeFunction(func(d server.Discovery) {
		discoveryStorage.Add(d)
	})
	driver.RegisterUnsubscribeFunction(func(d server.Discovery) {
		discoveryStorage.Delete(d.Hostname)
	})

	err = driver.Init()
	if err != nil {
		log.Fatalln(err)
	}

	go printDiscoveryLogs()
	go cleanDiscoveryPool()

	// If config.Register is false this instance won't be registered with other nodes
	if config.Register {
		// This is background process that sends the message
		go sendDiscoveryPacketTask(sendDiscoveryPacketTrigger)

		// This triggers the process
		go func() {
			for {
				sendDiscoveryPacket()

				time.Sleep(time.Duration(config.KeepAlive) * time.Second)

				if shuttingDown {
					break
				}
			}
		}()
	} else {
		log.Println("standalone mode, I won't register myself")
	}

	// --------
	// REST API
	// --------
	e := echo.New()

	// Middleware
	if len(config.Token) > 0 {
		e.Use(TokenMiddleware)
	}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	if !config.DisableAPI {
		e.GET("/", listHandler)
		e.GET("/v1/discovery", getIdentificationHandler)
		e.GET("/v1/discoveries", listHandler)
		e.POST("/v1/labels", addLabelsHandler)
		e.DELETE("/v1/labels", deleteLabelsHandler)
		e.GET("/v1/prometheus/:name", prometheusHandler)
	}

	// ------------------------------
	// Termination signals processing
	// ------------------------------

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func(e *echo.Echo, config Config) {
		sig := <-signals
		shuttingDown = true
		if config.Register {
			log.Printf("%s signal received, sending goodbye packet\n", sig.String())
			sendGoodbyePacket()
			time.Sleep(1 * time.Second) // we wait for a few seconds to let background jobs to finish their job
		} else {
			log.Printf("%s signal received", sig.String())
		}
		e.Shutdown(context.TODO())
	}(e, config)

	// Start server
	// In most cases this will end expectedly so it doesn't make sense to use the echo's approach to treat this message as an error.
	e.Logger.Info(e.Start(config.Host + ":" + strconv.Itoa(int(config.Port))))
}
