package main

import (
	"strconv"
	"strings"

	"github.com/rosti-cz/server_lobby/server"
)

// [
//   {
//     "targets": [ "<host>", ... ],
//     "labels": {
//       "<labelname>": "<labelvalue>", ...
//     }
//   },
//   ...
// ]

// PrometheusServices holds multiple PrometheusService structs
type PrometheusServices []PrometheusService

// PrometheusService represents a single set of targets and labels for Prometheus
type PrometheusService struct {
	Targets []string
	Labels  map[string]string
}

// preparePrometheusOutput returns PrometheusServices which is struct compatible to what Prometheus expects
// labels starting "ne:" will be used as NodeExporter labels. Label "ne:port:9123" will be used as port
// used in the targets field. Same for "ne:host:1.2.3.4".
func preparePrometheusOutput(discoveries []server.Discovery) PrometheusServices {
	services := PrometheusServices{}

	for _, discovery := range discoveries {
		port := strconv.Itoa(int(config.NodeExporterPort))
		host := discovery.Hostname

		labels := map[string]string{}

		for _, label := range discovery.FindLabels("ne") {
			trimmed := strings.TrimPrefix(label, "ne:")
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				if parts[0] == "port" {
					port = parts[1]
				} else if parts[0] == "host" {
					host = parts[1]
				} else {
					labels[parts[0]] = parts[1]
				}
			}
		}

		service := PrometheusService{
			Targets: []string{host + ":" + port},
			Labels:  labels,
		}

		services = append(services, service)

	}

	return services
}
