// +build windows

package main

const (
	// TaskName is the name of the task in the windows task scheduler library.
	TaskName = `WireGuard Toggle`

	// TaskDir is the directory the scheduled task will be placed in.
	TaskDir = `SCJ`

	// TaskPath is the path of the scheduled task.
	TaskPath = TaskDir + `\` + TaskName

	// AbsTaskPath is the rooted path of the scheduled task.
	AbsTaskPath = `\` + TaskPath
)

const (
	// NetworkChangeQuery is a Windows event query that matches network
	// connects and disconnects.
	NetworkChangeQuery = `<QueryList><Query Id='1'><Select Path='Microsoft-Windows-NetworkProfile/Operational'>*[System[(EventID=10000 or EventID=10001)]]</Select></Query></QueryList>`
)
