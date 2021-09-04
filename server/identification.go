package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// getIdentification assembles the discovery packet that contains hotname and set of labels describing a single server, in this case the local server.
// Parameter initialLabels usually coming from configuration of the app.
// If hostname is empty it will be discovered automatically.
func GetIdentification(hostname string, initialLabels Labels, labelsPath string) (Discovery, error) {
	discovery := Discovery{}

	localLabels, err := loadLocalLabels(initialLabels, labelsPath)
	if err != nil {
		return discovery, err
	}

	if len(hostname) == 0 {
		info, err := host.Info()
		if err != nil {
			return discovery, err
		}
		discovery.Hostname = info.Hostname
	} else {
		discovery.Hostname = hostname
	}

	discovery.Labels = append(initialLabels, localLabels...)

	return discovery, nil
}

// loadLocalLabels scans local directory where labels are stored and adds them to the labels configured as environment variables.
// Filename in LabelsPath is not importent and each file can contain multiple labels, one per each line.
func loadLocalLabels(skipLabels Labels, labelsPath string) (Labels, error) {
	labels := Labels{}
	var found bool

	if _, err := os.Stat(labelsPath); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(labelsPath)
		if err != nil {
			return labels, err
		}

		for _, filename := range files {
			fullPath := path.Join(labelsPath, filename.Name())

			content, err := os.ReadFile(fullPath)
			if err != nil {
				return labels, fmt.Errorf("read file error: %v", err)

			}
			fmt.Println(string(content))

			for _, line := range strings.Split(string(content), "\n") {
				line = strings.TrimSpace(line)
				if len(line) > 0 {
					found = false
					for _, skipLabel := range skipLabels {
						if skipLabel == Label(line) {
							found = true
							break
						}
					}
					if !found {
						labels = append(labels, Label(line))
					}
				}
			}
		}
	}
	fmt.Println("LABELS", labels)
	return labels, nil
}
