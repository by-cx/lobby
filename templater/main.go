package main

import (
	"flag"
	"fmt"
	"os"
)

// Templater is used to generate config files from predefined
// templates based on content of gathered discovery packets.
// It can for example configure Nginx's backend or database
// replication.
//
// It reads templates from /var/lib/lobby/templates (default)
// which are YAML files cotaining the template itself and command(s)
// that needs to be run when the template changes.

const defaultTemplatesPath = "/var/lib/lobby/templates"

var templatesPath *string

func init() {
	templatesPath = flag.String("templates-path", defaultTemplatesPath, "path of where templates are stored")

	flag.Parse()

	err := os.MkdirAll(*templatesPath, 0750)
	if err != nil {
		fmt.Println(err)
		flag.Usage()
		os.Exit(1)
	}

}

func main() {
	fmt.Println(*templatesPath)
}
