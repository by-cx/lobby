package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nats-io/nats.go"
	"github.com/rosti-cz/server_lobby/server"
)

var discoveryStorage server.Discoveries = server.Discoveries{}

var config Config

func init() {
	discoveryStorage.LogChannel = make(chan string)
}

// cleanDiscoveryPool clears the local server map and keeps only the alive servers
func cleanDiscoveryPool() {
	for {
		discoveryStorage.Clean()
		time.Sleep(time.Duration(config.CleanEvery) * time.Second)
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
		err = nc.Publish(config.NATSDiscoveryChannel, data)
		if err != nil {
			log.Printf("sending discovery error: %v\n", err)
		}
		time.Sleep(time.Duration(config.KeepAlive) * time.Second)
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

	discoveryStorage.TTL = config.TTL

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
	_, err = nc.Subscribe(config.NATSDiscoveryChannel, discoveryHandler)
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
	if len(config.Token) > 0 {
		e.Use(TokenMiddleware)
	}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", func(c echo.Context) error {
		label := c.QueryParam("label")

		var discoveries []server.Discovery

		if len(label) > 0 {
			discoveries = discoveryStorage.Filter(label)
		} else {
			discoveries = discoveryStorage.GetAll()
		}

		return c.JSONPretty(200, discoveries, "  ")
	})

	e.GET("/prometheus", func(c echo.Context) error {
		services := preparePrometheusOutput(discoveryStorage.GetAll())

		return c.JSONPretty(http.StatusOK, services, "  ")
	})

	// e.GET("/template/:template", func(c echo.Context) error {
	// 	templateName := c.Param("template")
	// 	discoveries := discoveryStorage.GetAll()
	// 	var body bytes.Buffer

	// 	tmpl, err := template.New("main").ParseFiles(path.Join(config.TemplatesPath, templateName))
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, err.Error())
	// 	}
	// 	err = tmpl.Execute(&body, &discoveries)
	// 	if err != nil {
	// 		return c.String(http.StatusInternalServerError, err.Error())
	// 	}

	// 	return c.String(http.StatusOK, body.String())
	// })

	// Start server
	e.Logger.Fatal(e.Start(config.Host + ":" + strconv.Itoa(int(config.Port))))
}
