package server

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDiscovery(t *testing.T) {
	now := time.Now().Unix()
	now90 := now - 90

	discovery := Discovery{
		Hostname: "test.rosti.cz",
		Labels: Labels{
			Label("service:test"),
			Label("test:123"),
			Label("public_ip:1.2.3.4"),
		},
		LastCheck: now,
	}

	assert.True(t, discovery.IsAlive(), "discovery suppose to be alive")
	discovery.LastCheck = now90
	assert.False(t, discovery.IsAlive(), "discovery not suppose to be alive")
	discovery.LastCheck = now

	assert.Equal(t, Labels{Label("service:test")}, discovery.FindLabels("service"))
	assert.Equal(t, nil, discovery.Validate()) // TODO: This needs more love

	content, err := json.Marshal(&discovery)
	assert.Nil(t, err)
	content2, err := discovery.Bytes()
	assert.Nil(t, err)
	assert.Equal(t, content, content2)
}
