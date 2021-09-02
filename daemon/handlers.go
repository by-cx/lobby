package main

import (
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
