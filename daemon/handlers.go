package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/rosti-cz/server_lobby/server"
)

func listHandler(c echo.Context) error {
	label := c.QueryParam("label")

	var discoveries []server.Discovery

	if len(label) > 0 {
		discoveries = discoveryStorage.Filter(label)
	} else {
		discoveries = discoveryStorage.GetAll()
	}

	return c.JSONPretty(200, discoveries, "  ")
}

func prometheusHandler(c echo.Context) error {
	services := preparePrometheusOutput(discoveryStorage.GetAll())

	return c.JSONPretty(http.StatusOK, services, "  ")
}
