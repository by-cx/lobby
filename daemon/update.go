package main

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/by-cx/lobby/server"
)

// These functions are called when something has changed in the storage

var changeDetectedChannel chan bool = make(chan bool)
var changeDetected bool

// changeCatcherLoop waits for a change signal and switches variable that says if the callback should run or not
func changeCatcherLoop() {
	for {
		<-changeDetectedChannel
		changeDetected = true
	}
}

// discoveryChangeLoop is used to process dynamic configuration changes.
// When there is a change detected a shell script or given process is triggered
// which does some operations with the new data. Usually it generates the
// configuration.
// This function has internal loop and won't allow to run the command more
// often than it's the configured amount of time. That prevents
func discoveryChangeLoop() {
	// This other loop tics in strict intervals and prevents the callback script to run more often than it's configured
	var cmd *exec.Cmd

	// Delay first run of the callback script a little so everything can set up
	log.Printf("Delaying start of discovery change loop (%d seconds)\n", config.CallbackFirstRunDelay)
	time.Sleep(time.Duration(config.CallbackFirstRunDelay) * time.Second)
	log.Println("Starting discovery change loop")

	for {
		if changeDetected {
			// We switch this at the beginning so we can detect new changes while the callback script is running
			changeDetected = false

			log.Println("Running callback function")

			// TODO: this is not the best way
			callbackCommandSlice := strings.Split(config.Callback, " ")

			if len(callbackCommandSlice) == 1 {
				cmd = exec.Command(callbackCommandSlice[0])
			} else if len(callbackCommandSlice) > 1 {
				cmd = exec.Command(callbackCommandSlice[0], callbackCommandSlice[1:]...)
			} else {
				log.Println("wrong number of parts of the callback command")
				time.Sleep(time.Duration(config.CallbackCooldown) * time.Second)
				continue
			}

			stdin, err := cmd.StdinPipe()
			if err != nil {
				log.Println("stdin writing error: ", err.Error())
				continue
			}

			discoveriesJSON, err := json.Marshal(discoveryStorage.GetAll())
			if err != nil {
				log.Println("stdin writing error: ", err.Error())
				continue
			}

			_, err = stdin.Write([]byte(discoveriesJSON))
			if err != nil {
				log.Println("stdin writing error: ", err.Error())
				continue
			}
			err = stdin.Close()
			if err != nil {
				log.Println("stdin writing error: ", err.Error())
				continue
			}

			stdout, err := cmd.CombinedOutput()
			if err != nil {
				log.Println("Callback error: ", err.Error())
			}
			log.Println("Callback output: ", string(stdout))
		}
		time.Sleep(time.Duration(config.CallbackCooldown) * time.Second)
	}
}

// discoveryChange is called when daemon detects that a newly arrived discovery
// packet is somehow different than the localone. This can be used to trigger
// some action in the local machine.
func discoveryChange(discovery server.Discovery) error {
	changeDetectedChannel <- true
	return nil
}
