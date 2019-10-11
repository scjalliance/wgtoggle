// +build windows

package main

import (
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"
)

// App returns a new wgtoggle kingpin app without any commands.
func App() *kingpin.Application {
	app := kingpin.New(filepath.Base(os.Args[0]), "Toggles WireGuard tunnels off and on based on network topology.")
	app.Interspersed(false)
	return app
}
