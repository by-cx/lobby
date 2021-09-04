package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/rosti-cz/server_lobby/common"
	"github.com/rosti-cz/server_lobby/nats_driver"
	"github.com/rosti-cz/server_lobby/server"
)

var discoveryStorage server.Discoveries = server.Discoveries{}
var driver common.Driver

var config Config

var shuttingDown bool

func init() {
	// Load config from environment variables
	config = *GetConfig()

	// Setup discovery storage
	discoveryStorage.LogChannel = make(chan string)
	discoveryStorage.TTL = config.TTL

	// Setup driver
	driver = &nats_driver.Driver{
		NATSUrl:              config.NATSURL,
		NATSDiscoveryChannel: config.NATSDiscoveryChannel,

		LogChannel: discoveryStorage.LogChannel,
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
	discovery, err := server.GetIdentification(config.HostName, config.Labels, config.LabelsPath)
	if err != nil {
		log.Printf("sending discovery identification error: %v\n", err)
	}

	err = driver.SendGoodbyePacket(discovery)
	if err != nil {
		log.Println(err)
	}
}

// sendDisoveryPacket sends discovery packet regularly so the network know we exist
func sendDiscoveryPacket() {
	for {
		discovery, err := server.GetIdentification(config.HostName, config.Labels, config.LabelsPath)
		if err != nil {
			log.Printf("sending discovery identification error: %v\n", err)
		}

		err = driver.SendDiscoveryPacket(discovery)
		if err != nil {
			log.Println(err)
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
		go sendDiscoveryPacket()
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
	e.GET("/", listHandler)
	e.GET("/v1/", listHandler)
	e.GET("/v1/prometheus/:name", prometheusHandler)

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

	go func(e *echo.Echo, config Config) {
		sig := <-signals
		shuttingDown = true
		if config.Register {
			log.Printf("%s signal received, sending goodbye packet\n", sig.String())
			sendGoodbyePacket()
			time.Sleep(5 * time.Second) // we wait for a few seconds to let background jobs to finish their job
		} else {
			log.Printf("%s signal received", sig.String())
		}
		e.Shutdown(context.TODO())
	}(e, config)

	// Start server
	e.Logger.Error(e.Start(config.Host + ":" + strconv.Itoa(int(config.Port))))
}
