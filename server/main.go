package server

import (
	"encoding/json"
	"fmt"
	"time"
)

// ----------------
// Discovery struct
// ----------------

const TimeToLife = 60 // when server won't occur in the discovery channel longer than this, it should be considered as not-alive

// Discovery contains information about a single server and is used for server discovery
type Discovery struct {
	Hostname string   `json:"hostname"`
	Labels   []string `json:"labels"`

	// For internal use to check if the server is still alive.
	// Contains timestamp of the last check.
	LastCheck int64 `json:"last_check"`
}

// Validate checks all values in the struct if the content is valid
func (d *Discovery) Validate() error {
	// TODO: implement
	return nil
}

// IsAlive return true if the server should be considered as alive
func (d *Discovery) IsAlive() bool {
	return time.Now().Unix()-d.LastCheck < TimeToLife
}

func (d *Discovery) Bytes() ([]byte, error) {
	data, err := json.Marshal(d)
	return data, err
}

// -----------------
// Discovery storage
// -----------------

// Discoveries helps to store instances of Discovery struct and access them in thread safe mode
type Discoveries struct {
	activeServers []Discovery
	LogChannel    chan string
}

// Add appends a new discovery/server to the storage
func (d *Discoveries) Add(discovery Discovery) {
	if d.Exist(discovery.Hostname) {
		d.Refresh(discovery.Hostname)
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
	for idx, discovery := range d.activeServers {
		if discovery.Hostname == hostname {
			d.activeServers[idx].LastCheck = time.Now().Unix()
		}
	}
}

// Delete removes server identified by hostname from the storage
func (d *Discoveries) Delete(hostname string) {
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

// Clean checks loops over last check values for each discovery object and removes it if it's passed
func (d *Discoveries) Clean() {
	newSet := []Discovery{}
	for _, server := range d.activeServers {
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
