package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/by-cx/lobby/server"
	"github.com/labstack/echo"
)

func listHandler(c echo.Context) error {
	labels := c.QueryParam("labels")
	prefixes := c.QueryParam("prefixes")

	var discoveries []server.Discovery

	if len(labels) > 0 {
		labelsFilterSlice := strings.Split(labels, ",")
		discoveries = discoveryStorage.Filter(labelsFilterSlice)
	} else if len(prefixes) > 0 {
		discoveries = discoveryStorage.FilterPrefix(strings.Split(prefixes, ","))
	} else {
		discoveries = discoveryStorage.GetAll()
	}

	return c.JSONPretty(200, discoveries, "  ")
}

// resolveHandler returns hostname(s) based on another label
func resolveHandler(c echo.Context) error {
	label := c.QueryParam("label") // This is label we will use to filter discovery packets

	output := []string{}

	discoveries := discoveryStorage.Filter([]string{label})
	for _, discovery := range discoveries {
		output = append(output, discovery.Hostname)
	}

	return c.JSONPretty(http.StatusOK, output, "  ")
}

func prometheusHandler(c echo.Context) error {
	name := c.Param("name")

	services := preparePrometheusOutput(name, discoveryStorage.GetAll())

	return c.JSONPretty(http.StatusOK, services, "  ")
}

func getIdentificationHandler(c echo.Context) error {
	discovery, err := localHost.GetIdentification()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("gathering identification info error: %v\n", err))
	}

	return c.JSONPretty(http.StatusOK, discovery, "  ")
}

func addLabelsHandler(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("reading request body error: %v\n", err))
	}

	labels := server.Labels{}

	for _, label := range strings.Split(string(body), "\n") {
		labels = append(labels, server.Label(label))
	}

	err = localHost.AddLabels(labels)

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Update the other nodes with this new change
	sendDiscoveryPacket()

	return c.String(http.StatusOK, "OK")
}

func deleteLabelsHandler(c echo.Context) error {
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("reading request body error: %v\n", err))
	}

	labels := server.Labels{}

	for _, label := range strings.Split(string(body), "\n") {
		labels = append(labels, server.Label(label))
	}

	err = localHost.DeleteLabels(labels)

	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Update the other nodes with this new change
	sendDiscoveryPacket()

	return c.String(http.StatusOK, "OK")
}
