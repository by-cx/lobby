package main

import (
	"strconv"
	"strings"

	"github.com/by-cx/lobby/server"
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
func preparePrometheusOutput(name string, discoveries []server.Discovery) PrometheusServices {
	services := PrometheusServices{}

	for _, discovery := range discoveries {
		port := strconv.Itoa(int(config.NodeExporterPort))
		hosts := []string{}
		var add bool // add to the prometheus output when there is at least one prometheus related label

		labels := map[string]string{} // These are prometheus labels, not Lobby's labels

		for _, label := range discovery.FindLabelsByPrefix("prometheus:" + name + ":") {
			trimmed := strings.TrimPrefix(label.String(), "prometheus:"+name+":")
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				if parts[0] == "port" {
					port = parts[1]
				} else if parts[0] == "host" {
					hosts = append(hosts, parts[1])
				} else {
					labels[parts[0]] = parts[1]
				}
				add = true
			}
		}

		// This has to be checked here again because FindLabels adds : at the end of the label name.
		if !add {
			for _, label := range discovery.Labels {
				if label.String() == "prometheus:"+name {
					add = true
					break
				}
			}
		}

		if add {
			targets := []string{}
			for _, host := range hosts {
				// Omit port part if "-" is set or port is part of the host
				target := host + ":" + port
				if strings.Contains(host, ":") || port == "-" {
					target = host
				}
				targets = append(targets, target)

			}
			service := PrometheusService{
				Targets: targets,
				Labels:  labels,
			}

			services = append(services, service)
		}
	}

	return services
}
