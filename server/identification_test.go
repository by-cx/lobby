package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tmpPath = "./tmp"
const testLabelPath = tmpPath + "/labels"

func TestGetIdentification(t *testing.T) {
	discovery, err := GetIdentification("test.example.com", Labels{Label("service:test"), Label("test:1")}, testLabelPath)
	assert.Nil(t, err)

	assert.Equal(t, "test.example.com", discovery.Hostname)
	assert.Equal(t, "service:test", discovery.Labels[0].String())

	err = os.MkdirAll(testLabelPath, os.ModePerm)
	assert.Nil(t, err)

	err = os.WriteFile(testLabelPath+"/test", []byte("service:test2\npublic_ip:1.2.3.4"), 0644)
	assert.Nil(t, err)

	discovery, err = GetIdentification("test.example.com", Labels{Label("service:test"), Label("test:1")}, testLabelPath)
	assert.Nil(t, err)

	assert.Equal(t, Label("public_ip:1.2.3.4"), discovery.Labels[3])

	os.RemoveAll(tmpPath)
}

func TestLoadLocalLabels(t *testing.T) {
	err := os.MkdirAll(testLabelPath, os.ModePerm)
	assert.Nil(t, err)

	err = os.WriteFile(testLabelPath+"/test", []byte("service:test\npublic_ip:1.2.3.4"), 0644)
	assert.Nil(t, err)

	labels, err := loadLocalLabels(Labels{Label("service:test")}, testLabelPath)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(labels))
	assert.Equal(t, "public_ip:1.2.3.4", labels[0].String())

	os.RemoveAll(tmpPath)
}
