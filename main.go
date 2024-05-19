package main

import (
	"flag"
	"os"

	"github.com/VitalyKondratiev/gtrack/src/command"
)

var taskId string
var g bool

func main() {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startCmd.StringVar(&taskId, "t", "", "Task Id")
	startCmd.BoolVar(&g, "g", false, "Current Git branch as Task Id")

	var cmd command.Command

	if len(os.Args) < 2 {
		cmd = command.HelpCommand()
	} else {
		switch os.Args[1] {
		case "auth":
			cmd = command.ChangeConfigCommand()
		case "list":
			cmd = command.ListCommand()
		case "start":
			startCmd.Parse(os.Args[2:])
			cmd = command.TrackingStartCommand(taskId, g)
		case "stop":
			cmd = command.TrackingStopCommand()
		case "commit":
			cmd = command.CommitCommand()
		case "update":
			cmd = command.GithubUpdateCommand()
		case "help":
			fallthrough
		default:
			cmd = command.HelpCommand()
		}
	}

	if cmd != nil {
		cmd.Execute()
	}
}
