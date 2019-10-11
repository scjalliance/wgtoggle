// +build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/capnspacehook/taskmaster"
	"github.com/gentlemanautomaton/cmdline/cmdlinewindows"
	"github.com/gentlemanautomaton/filework"
	"github.com/gentlemanautomaton/filework/fwos"
	"github.com/rickb777/date/period"
	"gopkg.in/alecthomas/kingpin.v2"
)

// InstallCommand returns an install command for app.
func InstallCommand(app *kingpin.Application) (*kingpin.CmdClause, *UpdateConfig) {
	cmd := app.Command("install", "Installs wgtoggle on the local machine.")
	conf := &UpdateConfig{}
	conf.Bind(cmd)
	return cmd, conf
}

func install(ctx context.Context, program string, conf UpdateConfig) {
	// Determine the source path
	sourcePath, err := filepath.Abs(program)
	if err != nil {
		fmt.Printf("Failed to determine the absolute path of %s: %v\n", program, err)
		os.Exit(1)
	}

	// Determine the installation directory
	dest, err := installDir(Version)
	if err != nil {
		fmt.Printf("Failed to locate installation directory: %v\n", err)
		os.Exit(1)
	}

	// TODO: Determine the version by using the PE package: https://golang.org/pkg/debug/pe/

	// Determine the source directory
	source, exe := filepath.Split(sourcePath)
	if !strings.HasSuffix(exe, ".exe") {
		exe += ".exe"
	}
	fmt.Printf("Installing %s to: %s\n", exe, dest)

	// Ensure that we can open the source file data
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		fmt.Printf("Failed to install %s: %v\n", exe, err)
		os.Exit(1)
	}
	defer sourceFile.Close()

	// Attempt to read the source file size
	var size int64
	if fi, err := sourceFile.Stat(); err == nil {
		size = fi.Size()
	}

	// Check to see if there's an existing file with the expected content
	diff, err := filework.CompareFileContent(sourceFile, fwos.Dir(dest), exe)
	if err != nil {
		fmt.Printf("Failed to examine existing %s file: %v\n", exe, err)
		os.Exit(1)
	}
	switch diff {
	case filework.Same:
		fmt.Printf("Existing %s file is up to date.\n", exe)
	case filework.Different:
		fmt.Printf("Existing %s file is out of date.\n", exe)
	}

	// Create the installation directory
	if err = os.MkdirAll(dest, os.ModePerm); err != nil {
		fmt.Printf("Failed to create installation directory \"%s\": %v\n", dest, err)
		os.Exit(1)
	}

	// Remove previous installation
	if data, err := getUninstallRegKeys(); err == nil {
		fmt.Printf("Removing %s version %s.\n", data.DisplayName, data.DisplayVersion)
		name, args := cmdlinewindows.SplitCommand(data.UninstallCommand)
		if name == "" {
			fmt.Printf("Failed to locate uninstaller.\n")
		} else {
			fmt.Printf("Executing: %s\n", data.UninstallCommand)
			cmd := exec.Command(name, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Removal failed: %v", err)
			} else {
				fmt.Printf("Removal succeeded.\n")
			}
		}
	}

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
		fmt.Printf("Existing \"%s\" task found.\n", TaskPath)
		if err := tasks.DeleteTask(AbsTaskPath); err != nil {
			fmt.Printf("Removal of existing task failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("The \"%s\" task has been removed.\n", TaskPath)
	}

	// Copy the executable
	result := filework.CopyFile(fwos.Dir(source), exe, sourceFile, fwos.Dir(dest), exe)
	if result.Err != nil {
		fmt.Printf("Failed to copy %s executable: %v\n", exe, err)
		os.Exit(1)
	}
	fmt.Printf("%s copied to %s\n", exe, dest)

	// Add an uninstall entry to the Windows registry
	if err := addUninstallRegKeys(source, dest, exe, size); err != nil {
		fmt.Printf("Failed to write uninstall entry to the Windows registry: %v\n", err)
	}
	fmt.Printf("Wrote uninstall entry to the Windows registry.\n")

	// Prepare the task definition
	def := tasks.NewTaskDefinition()
	def.AddExecAction(filepath.Join(dest, exe), execArgs(conf), dest, fmt.Sprintf("Update %s Tunnel"+conf.Tunnel))
	def.RegistrationInfo.Author = CompanyName
	def.RegistrationInfo.Description = fmt.Sprintf("Disables the %s WireGuard tunnel when in the office", conf.Tunnel)
	def.Principal.ID = "Author"
	def.Principal.UserID = "S-1-5-18"
	def.Settings.StopIfGoingOnBatteries = false
	def.Settings.Compatibility = taskmaster.TASK_COMPATIBILITY_V2_4
	def.Settings.MultipleInstances = taskmaster.TASK_INSTANCES_QUEUE
	def.Settings.TimeLimit = period.NewHMS(0, 5, 0)
	def.AddEventTrigger(period.NewHMS(0, 0, 5), NetworkChangeQuery, nil)

	// Install the task
	//_, updated, err := tasks.CreateTaskEx(AbsTaskPath, def, "Local System", "", taskmaster.TASK_LOGON_SERVICE_ACCOUNT, false)
	task, _, err := tasks.CreateTask(AbsTaskPath, def, true)
	if err != nil {
		fmt.Printf("Failed to create \"%s\" task: %v\n", TaskPath, err)
		os.Exit(1)
	}
	fmt.Printf("\"%s\" task created successfully.\n", TaskPath)

	// Run the task
	fmt.Printf("Running task.\n")
	runner, err := task.Run(nil)
	if err != nil {
		fmt.Printf("Failed to start task: %v\n", err)
		os.Exit(1)
	}
	defer runner.Release()

	fmt.Printf("Task started.\n")
}

func installDir(version string) (dir string, err error) {
	dir = os.Getenv("PROGRAMFILES")
	if dir == "" {
		return "", errors.New("unable to determine ProgramFiles location")
	}

	return filepath.Join(dir, "SCJ", "wgtoggle", version), nil
}

func execArgs(conf UpdateConfig) string {
	args := []string{"update"}
	if conf.Tunnel != "" {
		args = append(args, "-t", conf.Tunnel)
	}
	if len(conf.Gateways) > 0 {
		for _, gw := range conf.Gateways {
			args = append(args, "-g", gw)
		}
	}

	var s string
	for _, v := range args {
		if s != "" {
			s += " "
		}
		s += syscall.EscapeArg(v)
	}
	return s
}
