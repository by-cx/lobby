package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/by-cx/lobby/client"
	"github.com/by-cx/lobby/server"
)

func Usage() {
	flag.Usage()
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  discovery                        returns discovery packet of the server where the client is connected to")
	fmt.Println("  discoveries                      returns list of all registered discovery packets")
	fmt.Println("  discoveries labels [LABEL] ...   returns list of all registered discovery packets with given labels (OR)")
	fmt.Println("  discoveries search [LABEL] ...   returns list of all registered discovery packets with given label prefixes (OR)")
	fmt.Println("  labels add LABEL [LABEL] ...     adds new runtime labels")
	fmt.Println("  labels del LABEL [LABEL] ...     deletes runtime labels")
}

func main() {
	config := GetConfig()

	// Setup flags
	proto := flag.String("proto", "", "Select HTTP or HTTPS protocol")
	host := flag.String("host", "", "Hostname or IP address of lobby daemon")
	port := flag.Uint("port", 0, "Port of lobby daemon")
	token := flag.String("token", "", "Token needed to communicate lobby daemon, if empty auth is disabled")
	jsonOutput := flag.Bool("json", false, "set output to JSON, error will be still in plain text")

	flag.Parse()

	// Replace empty values from flags by values from environment variables
	if *proto == "" {
		proto = &config.Proto
	}
	if *host == "" {
		host = &config.Host
	}
	if *port == 0 {
		port = &config.Port
	}
	if *token == "" {
		token = &config.Token
	}

	// Validation
	if *proto != "http" && *proto != "https" {
		fmt.Println("Protocol can be only http or https")
	}

	// Setup lobby client library
	client := client.LobbyClient{
		Proto: strings.ToLower(*proto),
		Host:  *host,
		Port:  *port,
		Token: *token,
	}

	// Process rest of the arguments
	if len(flag.Args()) == 0 {
		Usage()
		os.Exit(0)
	}

	switch flag.Args()[0] {
	case "discoveries":
		var discoveries []server.Discovery
		var err error

		if len(flag.Args()) > 2 {
			if flag.Arg(1) == "labels" {
				labels := []server.Label{}
				for _, label := range flag.Args()[2:] {
					labels = append(labels, server.Label(label))
				}

				discoveries, err = client.FindByLabels(labels)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else if flag.Arg(1) == "search" {
				discoveries, err = client.FindByPrefixes(flag.Args()[2:])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				fmt.Println("ERROR: unknown usage of discoveries arguments")
				fmt.Println("")
				Usage()
				os.Exit(0)
			}
		} else {
			discoveries, err = client.GetDiscoveries()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		if *jsonOutput {
			printJSON(discoveries)
		} else {
			printDiscoveries(discoveries)
		}
	case "discovery":
		discovery, err := client.GetDiscovery()
		if err != nil {
			fmt.Println(err)
		}

		if *jsonOutput {
			printJSON(discovery)
		} else {
			printDiscovery(discovery)
		}
	case "labels":
		if len(flag.Args()) < 3 {
			fmt.Println("ERROR: not enough arguments for labels command")
			fmt.Println("")
			Usage()
			os.Exit(0)
		}

		labels := server.Labels{}
		labelsString := flag.Args()[2:]
		for _, labelString := range labelsString {
			labels = append(labels, server.Label(labelString))
		}

		if flag.Args()[1] == "add" {
			err := client.AddLabels(labels)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				os.Exit(2)
			}
		} else if flag.Args()[1] == "del" {
			err := client.DeleteLabels(labels)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
				os.Exit(2)
			}
		} else {
			fmt.Printf("ERROR: wrong labels subcommand\n\n")
			Usage()
			os.Exit(2)
		}

	default:
		Usage()
		os.Exit(0)
	}

}
