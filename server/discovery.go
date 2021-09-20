package server

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// ----------------
// Discovery struct
// ----------------

const TimeToLife = 60 // when server won't occur in the discovery channel longer than this, it should be considered as not-alive

// Discovery contains information about a single server and is used for server discovery
type Discovery struct {
	Hostname string `json:"hostname"`
	Labels   Labels `json:"labels"`

	// For internal use to check if the server is still alive.
	// Contains timestamp of the last check.
	LastCheck int64 `json:"last_check"`

	TTL uint `json:"-"` // after how many second consider the server to be off, if 0 then 60 secs is used
}

// Validate checks all values in the struct if the content is valid
func (d *Discovery) Validate() error {
	// TODO: implement
	return nil
}

// IsAlive return true if the server should be considered as alive
func (d *Discovery) IsAlive() bool {
	if d.TTL == 0 {
		d.TTL = TimeToLife
	}
	return time.Now().Unix()-d.LastCheck < int64(d.TTL)
}

func (d *Discovery) Bytes() ([]byte, error) {
	data, err := json.Marshal(d)
	return data, err
}

// FindLabelsByPrefix returns list of labels with given prefix. For example "service:ns" has prefix "service" or "service:".
// It doesn't have to be prefix, but for example "service:test" will match "service:test" and also "service:test2".
func (d *Discovery) FindLabelsByPrefix(prefix string) Labels {
	labels := Labels{}
	for _, label := range d.Labels {
		if strings.HasPrefix(label.String(), prefix) {
			labels = append(labels, label)
		}
	}
	return labels
}

func (d *Discovery) SortLabels() {
	labelStrings := d.Labels.StringSlice()
	sort.Strings(labelStrings)

	labels := Labels{}
	for _, label := range labelStrings {
		labels = append(labels, Label(label))
	}

	d.Labels = labels
}

// -----------------
// Discovery storage
// -----------------

// Discoveries helps to store instances of Discovery struct and access them in thread safe mode
type Discoveries struct {
	activeServers []Discovery
	LogChannel    chan string
	TTL           uint
}

func (d *Discoveries) hostnameIndex(hostname string) int {
	for idx, discovery := range d.activeServers {
		if discovery.Hostname == hostname {
			return idx
		}
	}
	return -1
}

// Add appends a new discovery/server to the storage
func (d *Discoveries) Add(discovery Discovery) {
	if d.Exist(discovery.Hostname) {
		d.Refresh(discovery.Hostname)

		idx := d.hostnameIndex(discovery.Hostname)
		if idx >= 0 {
			d.activeServers[idx].Labels = discovery.Labels
		}

		return
	}

	discovery.LastCheck = time.Now().Unix()
	d.activeServers = append(d.activeServers, discovery)
	if d.LogChannel != nil {
		d.LogChannel <- fmt.Sprintf("%s registered", discovery.Hostname)
	}
}

// Refresh updates
func (d *Discoveries) Refresh(hostname string) {
	idx := d.hostnameIndex(hostname)
	if idx >= 0 {
		d.activeServers[idx].LastCheck = time.Now().Unix()
	}
}

// Delete removes server identified by hostname from the storage
func (d *Discoveries) Delete(hostname string) {
	if !d.Exist(hostname) {
		return
	}

	if d.LogChannel != nil {
		d.LogChannel <- fmt.Sprintf("removing %s", hostname)
	}

	newSet := []Discovery{}
	for _, server := range d.activeServers {
		if server.Hostname != hostname {
			newSet = append(newSet, server)
		}
	}
	d.activeServers = newSet
}

// Exist returns true if server with given hostname exists
func (d *Discoveries) Exist(hostname string) bool {
	for _, server := range d.activeServers {
		if server.Hostname == hostname {
			return true
		}
	}
	return false
}

// Get returns Discovery struct with the given hostname but it can be also an empty struct if it's not found. Check if hostname is empty or use Exist first to be sure.
func (d *Discoveries) Get(hostname string) Discovery {
	for _, server := range d.activeServers {
		if server.Hostname == hostname {
			return server
		}
	}
	return Discovery{}
}

// GetAll returns copy of the internal storage
func (d *Discoveries) GetAll() []Discovery {
	return d.activeServers
}

// Filter returns list of discoveries based on given labels
func (d *Discoveries) Filter(labelsFilter []string) []Discovery {
	newSet := []Discovery{}

	var found bool
	if len(labelsFilter) > 0 {
		for _, discovery := range d.activeServers {
			newDiscovery := discovery
			newDiscovery.Labels = Labels{}

			found = false
			for _, label := range discovery.Labels {
				for _, labelFilter := range labelsFilter {
					if label.String() == labelFilter {
						found = true
						newDiscovery.Labels = append(newDiscovery.Labels, label)
						break
					}
				}
			}

			if found {
				newSet = append(newSet, newDiscovery)
			}

		}

	}

	return newSet
}

// Filter returns list of discoveries based on given label prefixes.
func (d *Discoveries) FilterPrefix(prefixes []string) []Discovery {
	newSet := []Discovery{}

	var found bool
	if len(prefixes) > 0 {
		for _, discovery := range d.activeServers {
			newDiscovery := discovery
			newDiscovery.Labels = Labels{}

			found = false

			if found {
				newSet = append(newSet, newDiscovery)
			}

			for _, label := range discovery.Labels {
				for _, prefix := range prefixes {
					if strings.HasPrefix(label.String(), prefix) {
						found = true
						newDiscovery.Labels = append(newDiscovery.Labels, label)
						break
					}
				}
			}

			if found {
				newSet = append(newSet, newDiscovery)
			}
		}

	}

	return newSet
}

// Clean checks loops over last check values for each discovery object and removes it if it's passed
func (d *Discoveries) Clean() {
	newSet := []Discovery{}
	for _, server := range d.activeServers {
		server.TTL = d.TTL
		if server.IsAlive() {
			newSet = append(newSet, server)
		} else {
			if d.LogChannel != nil {
				d.LogChannel <- fmt.Sprintf("%s not alive anymore", server.Hostname)
			}
		}
	}
	d.activeServers = newSet
}
