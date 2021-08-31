package main

import (
	"log"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nats-io/nats.go"
	"github.com/rosti-cz/server_lobby/server"
)

const discoveryChannel = "lobby.discovery"
const cleanEvery = 15 // clean discoveredServers every X seconds
const keepAlive = 15  // sends discovery struct every

var discoveryStorage server.Discoveries = server.Discoveries{}

var config Config

func init() {
	discoveryStorage.LogChannel = make(chan string)
}

// cleanDiscoveryPool clears the local server map and keeps only the alive servers
func cleanDiscoveryPool() {
	for {
		discoveryStorage.Clean()
		time.Sleep(cleanEvery * time.Second)
	}

}

// sendDisoveryPacket sends discovery packet regularly so the network know we exist
func sendDisoveryPacket(nc *nats.Conn) {
	for {
		discovery, err := getIdentification()
		if err != nil {
			log.Printf("sending discovery identification error: %v\n", err)
		}

		data, err := discovery.Bytes()
		if err != nil {
			log.Printf("sending discovery formating message error: %v\n", err)
		}
		err = nc.Publish(discoveryChannel, data)
		if err != nil {
			log.Printf("sending discovery error: %v\n", err)
		}
		time.Sleep(keepAlive * time.Second)
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

	// Load config from environment variables
	config = *GetConfig()

	// ------------------------
	// Server discovering stuff
	// ------------------------

	// Connect to the NATS service
	nc, err := nats.Connect(config.NATSURL)
	if err != nil {
		log.Fatalln(err)
	}
	defer nc.Drain()

	go printDiscoveryLogs()

	// Subscribe
	log.Println("> discovery channel")
	_, err = nc.Subscribe(discoveryChannel, discoveryHandler)
	if err != nil {
		log.Fatalln(err)
	}

	go cleanDiscoveryPool()
	go sendDisoveryPacket(nc)

	// --------
	// REST API
	// --------
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", func(c echo.Context) error {
		discoveries := discoveryStorage.GetAll()
		return c.JSONPretty(200, discoveries, "  ")
	})

	// Start server
	e.Logger.Fatal(e.Start(":1313"))
}
