package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/rosti-cz/server_lobby/server"
)

func printDiscovery(discovery server.Discovery) {
	color.Yellow("Hostname:\n  %s\n", discovery.Hostname)

	if len(discovery.Labels) > 0 {
		fmt.Printf("Labels:\n")
		for _, label := range discovery.Labels {
			fmt.Printf("  %s\n", label)
		}
	}
}

func colorLabel(label server.Label) string {
	parts := strings.Split(label.String(), ":")
	if len(parts) == 1 {
		return color.GreenString(parts[0])
	}

	return color.GreenString(parts[0]) + ":" + color.MagentaString((strings.Join(parts[1:], ":")))
}

func printDiscoveries(discoveries []server.Discovery) {
	maxHostnameWidth := 0
	for _, discovery := range discoveries {
		if len(discovery.Hostname) > maxHostnameWidth {
			maxHostnameWidth = len(discovery.Hostname)
		}
	}

	for _, discovery := range discoveries {
		if len(discovery.Labels) == 0 {
			// fmt.Println(discovery.Hostname)
			color.Yellow(discovery.Hostname)
		} else {
			hostname := fmt.Sprintf("%"+strconv.Itoa(maxHostnameWidth)+"s", discovery.Hostname)

			fmt.Printf("%s    %s\n", color.YellowString(hostname), colorLabel(discovery.Labels[0]))

			if len(discovery.Labels) > 1 {
				for _, label := range discovery.Labels[1:] {
					fmt.Printf("%"+strconv.Itoa(maxHostnameWidth+4)+"s%s\n", " ", colorLabel(label))
				}
			}
		}
		fmt.Println()
	}
}

func printJSON(data interface{}) {
	body, err := json.Marshal(data)
	if err != nil {
		fmt.Println("error occurred while formating the output into JSON:", err.Error())
		os.Exit(3)
	}

	fmt.Println(string(body))
}
