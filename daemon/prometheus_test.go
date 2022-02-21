package main

import (
	"testing"

	"github.com/by-cx/lobby/server"
	"github.com/stretchr/testify/assert"
)

func TestPreparePrometheusOutput(t *testing.T) {
	discoveries := []server.Discovery{
		{
			Hostname: "test.server",
			Labels: server.Labels{
				// test1
				"prometheus:test1:label1:l1",
				"prometheus:test1:host:srv1:1234",
				"prometheus:test1:host:srv1:1235",
				"prometheus:test1:host:srv1",
				"prometheus:test1:label2:l2",
				// test2
				"prometheus:test2:host:srv2",
				"prometheus:test2:host:srv2:1235",
				"prometheus:test2:host:srv2:1236",
				"prometheus:test2:host:srv2:1237",
				"prometheus:test2:label1:l3",
			},
		},
	}
	services := preparePrometheusOutput("test1", discoveries)
	assert.Equal(t, 1, len(services))
	assert.Equal(t, 3, len(services[0].Targets))
	assert.Contains(t, services[0].Targets, "srv1:9100")
	assert.Contains(t, services[0].Targets, "srv1:1234")
	assert.Contains(t, services[0].Targets, "srv1:1235")
	assert.Contains(t, services[0].Labels["label1"], "l1")
	assert.Contains(t, services[0].Labels["label2"], "l2")

	services = preparePrometheusOutput("test2", discoveries)
	assert.Equal(t, 1, len(services))
	assert.Equal(t, 4, len(services[0].Targets))
	assert.Contains(t, services[0].Targets, "srv2:9100")
	assert.Contains(t, services[0].Targets, "srv2:1235")
	assert.Contains(t, services[0].Targets, "srv2:1236")
	assert.Contains(t, services[0].Targets, "srv2:1237")
	assert.Contains(t, services[0].Labels["label1"], "l3")
}
