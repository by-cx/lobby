package main

import (
	"github.com/rosti-cz/server_lobby/server"
	"github.com/shirou/gopsutil/v3/host"
)

func getIdentification() (server.Discovery, error) {
	discovery := server.Discovery{}

	info, err := host.Info()
	if err != nil {
		return discovery, err
	}
	discovery.Hostname = info.Hostname
	discovery.Labels = config.Labels

	return discovery, nil
}
