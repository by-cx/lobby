package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

type LocalHost struct {
	LabelsPath            string // Where labels are stored
	RuntimeLabelsFilename string // Filename under which are runtime labels saved in LabelsPath
	InitialLabels         Labels // this usually coming from the config
	HostnameOverride      string // if not empty string hostname in the discovery packet will be replaced by this
}

// saveRuntimeLabels stores labels in the runtime filesname
func (l *LocalHost) saveRuntimeLabels(labels Labels) error {
	stringLabels := []string{}

	for _, label := range labels {
		stringLabels = append(stringLabels, label.String())
	}

	content := strings.Join(stringLabels, "\n")

	err := os.WriteFile(path.Join(l.LabelsPath, l.RuntimeLabelsFilename), []byte(content), 0755)
	return err
}

// getRuntimeLabels returns labels from the runtime filename
func (l *LocalHost) getRuntimeLabels() (Labels, error) {
	labels := Labels{}

	content, err := os.ReadFile(path.Join(l.LabelsPath, l.RuntimeLabelsFilename))
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return labels, nil
		}

		return labels, err
	}

	for _, label := range strings.Split(string(content), "\n") {
		labels = append(labels, Label(strings.TrimSpace(label)))
	}

	return labels, nil
}

// AddLabel adds runtime label into the LabelsPath directory
func (l *LocalHost) AddLabels(labels Labels) error {
	runtimeLabels, err := l.getRuntimeLabels()
	if err != nil {
		return fmt.Errorf("error while loading stored labels: %v", err)
	}

	var found bool

	for _, label := range labels {
		found = false

		for _, runtimeLabel := range runtimeLabels {
			if label == runtimeLabel {
				found = true
				break
			}
		}

		if !found {
			runtimeLabels = append(runtimeLabels, label)
		}
	}

	err = l.saveRuntimeLabels(runtimeLabels)
	if err != nil {
		return fmt.Errorf("error while saving new set of labels: %v", err)
	}

	return nil
}

// DeleteLabels removed labels from LabelsPath directory. Only labels added this way can be deleted.
func (l *LocalHost) DeleteLabels(labels Labels) error {
	runtimeLabels, err := l.getRuntimeLabels()
	if err != nil {
		return fmt.Errorf("error while loading stored labels: %v", err)
	}

	newSet := Labels{}
	var found bool

	for _, runtimeLabel := range runtimeLabels {
		found = false

		for _, label := range labels {
			if label == runtimeLabel {
				found = true
				break
			}
		}

		if !found {
			newSet = append(newSet, runtimeLabel)
		}
	}

	err = l.saveRuntimeLabels(newSet)
	if err != nil {
		return fmt.Errorf("error while saving new set of labels: %v", err)
	}

	return nil
}

// GetIdentification assembles the discovery packet that contains hotname and set of labels describing a single server, in this case the local server.
// Parameter initialLabels usually coming from configuration of the app.
// If hostname is empty it will be discovered automatically.
func (l *LocalHost) GetIdentification() (Discovery, error) {
	discovery := Discovery{}

	localLabels, err := l.loadLocalLabels()
	if err != nil {
		return discovery, err
	}

	if len(l.HostnameOverride) == 0 {
		info, err := host.Info()
		if err != nil {
			return discovery, err
		}
		discovery.Hostname = info.Hostname
	} else {
		discovery.Hostname = l.HostnameOverride
	}

	discovery.Labels = append(l.InitialLabels, localLabels...)
	discovery.SortLabels()

	return discovery, nil
}

// loadLocalLabels scans local directory where labels are stored and adds them to the labels configured as environment variables.
// Filename in LabelsPath is not importent and each file can contain multiple labels, one per each line.
func (l *LocalHost) loadLocalLabels() (Labels, error) {
	labels := Labels{}
	var found bool

	if _, err := os.Stat(l.LabelsPath); !os.IsNotExist(err) {
		files, err := ioutil.ReadDir(l.LabelsPath)
		if err != nil {
			return labels, err
		}

		for _, filename := range files {
			fullPath := path.Join(l.LabelsPath, filename.Name())

			content, err := os.ReadFile(fullPath)
			if err != nil {
				return labels, fmt.Errorf("read file error: %v", err)

			}

			for _, line := range strings.Split(string(content), "\n") {
				line = strings.TrimSpace(line)
				if len(line) > 0 {
					found = false
					for _, skipLabel := range l.InitialLabels {
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
	return labels, nil
}
