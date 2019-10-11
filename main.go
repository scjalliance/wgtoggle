// +build windows

package main

import (
	"os"
	"syscall"

	"github.com/gentlemanautomaton/signaler"
)

func main() {
	var (
		app                     = App()
		installCmd, installConf = InstallCommand(app)
		uninstallCmd            = UninstallCommand(app)
		updateCmd, updateConf   = UpdateCommand(app)
	)

	command, err := app.Parse(os.Args[1:])
	if err != nil {
		app.Usage(os.Args[1:])
		os.Exit(1)
	}

	// Shutdown when we receive a termination signal
	shutdown := signaler.New().Capture(os.Interrupt, syscall.SIGTERM)

	// Ensure that we cleanup even if we panic
	defer shutdown.Trigger()

	switch command {
	case installCmd.FullCommand():
		install(shutdown.Context(), os.Args[0], *installConf)
	case uninstallCmd.FullCommand():
		uninstall(shutdown.Context())
	case updateCmd.FullCommand():
		update(shutdown.Context(), *updateConf)
	}
}
