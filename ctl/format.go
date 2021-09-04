package main

import (
	"fmt"
	"strconv"

	"github.com/rosti-cz/server_lobby/server"
)

func printDiscovery(discovery server.Discovery) {
	fmt.Printf("Hostname:\n  %s\n", discovery.Hostname)

	if len(discovery.Labels) > 0 {
		fmt.Printf("Labels:\n")
		for _, label := range discovery.Labels {
			fmt.Printf("  %s\n", label)
		}
	}
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
			fmt.Println(discovery.Hostname)
		} else {
			hostname := fmt.Sprintf("%"+strconv.Itoa(maxHostnameWidth)+"s", discovery.Hostname)
			fmt.Printf("%s    %s\n", hostname, discovery.Labels[0].String())
			if len(discovery.Labels) > 1 {
				for _, label := range discovery.Labels[1:] {
					fmt.Printf("%"+strconv.Itoa(maxHostnameWidth+4)+"s%s\n", " ", label)
				}
			}
		}
		fmt.Println()
	}
}
