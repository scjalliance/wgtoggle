// +build windows

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/capnspacehook/taskmaster"
	"gopkg.in/alecthomas/kingpin.v2"
)

// UninstallCommand returns an uninstall command for app.
func UninstallCommand(app *kingpin.Application) *kingpin.CmdClause {
	return app.Command("uninstall", "Uninstalls wgtoggle from the local machine.")
}

func uninstall(ctx context.Context) {
	// Connect to the task service
	tasks, err := taskmaster.Connect("", "", "", "")
	if err != nil {
		fmt.Printf("Failed to connect to task scheduler service: %v\n", err)
		os.Exit(1)
	}
	defer tasks.Disconnect()

	// Stop and remove any existing task
	if task, err := tasks.GetRegisteredTask(AbsTaskPath); err != nil {
		fmt.Printf("Failed to check for existing task: %v\n", err)
		os.Exit(1)
	} else if task != nil {
		if err := tasks.DeleteTask(AbsTaskPath); err != nil {
			fmt.Printf("Removal of existing task failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("The \"%s\" task has been removed.\n", TaskPath)
	}

	// Remove registry keys
	if err := delUninstallRegKeys(); err != nil {
		fmt.Printf("Failed to remove uninstall entry from the Windows registry: %v\n", err)
	}
	fmt.Printf("Removed uninstall entry from from the Windows registry.\n")
}

// uninstallCommand returns a command line string can be run to uninstall
// resourceful. The returned string will be properly quoted.
func uninstallCommand(dir, executable string) string {
	return syscall.EscapeArg(filepath.Join(dir, executable)) + " uninstall"
}
