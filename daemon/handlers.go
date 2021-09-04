package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/rosti-cz/server_lobby/server"
)

func listHandler(c echo.Context) error {
	labels := c.QueryParam("labels")

	var discoveries []server.Discovery

	if len(labels) > 0 {
		labelsFilterSlice := strings.Split(labels, ",")
		discoveries = discoveryStorage.Filter(labelsFilterSlice)
	} else {
		discoveries = discoveryStorage.GetAll()
	}

	return c.JSONPretty(200, discoveries, "  ")
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

	return c.String(http.StatusOK, "OK")
}
