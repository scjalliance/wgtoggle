// +build windows

package main

import (
	"net"
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

// AppConfig returns configuration for app.
func AppConfig(app *kingpin.Application) *Config {
	conf := &Config{}
	conf.Bind(app)
	return conf
}

// Config stores wgtunnel configuration.
type Config struct {
	Tunnel  string
	Gateway []net.IP
}

// Bind binds the configuration to the application.
func (conf *Config) Bind(app *kingpin.Application) {
	app.Flag("tunnel", "WireGuard tunnel name to toggle").Short('t').Required().StringVar(&conf.Tunnel)
	app.Flag("gateway", "Gateway IP address match").Short('g').Required().IPListVar(&conf.Gateway)
}
