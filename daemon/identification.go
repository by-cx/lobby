package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/rosti-cz/server_lobby/server"
	"github.com/shirou/gopsutil/v3/host"
)

// getIdentification assembles the discovery packet that contains hotname and set of labels describing a single server, in this case the local server.
func getIdentification() (server.Discovery, error) {
	discovery := server.Discovery{}

	localLabels, err := loadLocalLabels(config.Labels)
	if err != nil {
		return discovery, err
	}

	if len(config.HostName) == 0 {
		info, err := host.Info()
		if err != nil {
			return discovery, err
		}
		discovery.Hostname = info.Hostname
	} else {
		discovery.Hostname = config.HostName
	}

	discovery.Labels = append(config.Labels, localLabels...)

	return discovery, nil
}

// loadLocalLabels scans local directory where labels are stored and adds them to the labels configured as environment variables.
// Filename in LabelsPath is not importent and each file can contain multiple labels, one per each line.
func loadLocalLabels(skipLabels []string) ([]string, error) {
	labels := []string{}
	var found bool

	if _, err := os.Stat(config.LabelsPath); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(config.LabelsPath)
		if err != nil {
			return labels, err
		}

		for _, filename := range files {
			fullPath := path.Join(config.LabelsPath, filename.Name())
			fp, err := os.OpenFile(fullPath, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return labels, fmt.Errorf("open file error: %v", err)

			}
			defer fp.Close()

			rd := bufio.NewReader(fp)
			for {
				line, err := rd.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}

					return labels, fmt.Errorf("read file line error: %v", err)
				}
				line = strings.TrimSpace(line)
				if len(line) > 0 {
					found = false
					for _, skipLabel := range skipLabels {
						if skipLabel == line {
							found = true
							break
						}
					}
					if !found {
						labels = append(labels, line)
					}
				}
			}
		}
	}

	return labels, nil
}
