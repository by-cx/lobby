package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/nats-io/nats.go"
	"github.com/rosti-cz/server_lobby/server"
)

var discoveryStorage server.Discoveries = server.Discoveries{}

var config Config

var shuttingDown bool

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

// sendGoodbyePacket is almost same as sendDiscoveryPacket but it's not running in loop
// and it adds goodbye message so other nodes know this node is gonna die.
func sendGoodbyePacket(nc *nats.Conn) {
	discovery, err := getIdentification()
	if err != nil {
		log.Printf("sending discovery identification error: %v\n", err)
	}

	envelope := discoveryEnvelope{
		Discovery: discovery,
		Message:   "goodbye",
	}

	data, err := envelope.Bytes()
	if err != nil {
		log.Printf("sending discovery formating message error: %v\n", err)
	}
	err = nc.Publish(config.NATSDiscoveryChannel, data)
	if err != nil {
		log.Printf("sending discovery error: %v\n", err)
	}
}

// sendDisoveryPacket sends discovery packet regularly so the network know we exist
func sendDiscoveryPacket(nc *nats.Conn) {
	for {
		discovery, err := getIdentification()
		if err != nil {
			log.Printf("sending discovery identification error: %v\n", err)
		}

		envelope := discoveryEnvelope{
			Discovery: discovery,
			Message:   "hi",
		}

		data, err := envelope.Bytes()
		if err != nil {
			log.Printf("sending discovery formating message error: %v\n", err)
		}
		err = nc.Publish(config.NATSDiscoveryChannel, data)
		if err != nil {
			log.Printf("sending discovery error: %v\n", err)
		}
		time.Sleep(time.Duration(config.KeepAlive) * time.Second)

		if shuttingDown {
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
	go sendDiscoveryPacket(nc)

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

	// ------------------------------
	// Termination signals processing
	// ------------------------------

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func(nc *nats.Conn, e *echo.Echo) {
		sig := <-signals
		shuttingDown = true
		log.Printf("%s signal received, sending goodbye packet\n", sig.String())
		sendGoodbyePacket(nc)
		time.Sleep(5 * time.Second) // we wait for a few seconds to let background jobs to finish their job
		e.Shutdown(context.TODO())
	}(nc, e)

	// Start server
	e.Logger.Error(e.Start(config.Host + ":" + strconv.Itoa(int(config.Port))))
}
