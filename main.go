package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"./src/config"
	"./src/jira"
	"./src/toggl"
)

func main() {
	if len(os.Args) < 2 {
		CommandHelp()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "auth":
		CommandAuth()
	case "list":
		CommandList()
	case "help":
		gconfig := (config.GlobalConfig{}).LoadConfig()
		fmt.Print(gconfig.Jira.Username)
		CommandHelp()
	default:
		CommandHelp()
	}
}

func CommandAuth() {
	jira := (jira.Jira{}).SetCredentials()
	if !jira.IsLoggedIn() {
		return
	}
	toggl := (toggl.Toggl{}).SetCredentials()
	if !toggl.IsLoggedIn() {
		return
	}
	(config.GlobalConfig{}).SetConfig(jira.Credentials, toggl.Credentials).SaveConfig()
}

func CommandList() {
	gconfig := (config.GlobalConfig{}).LoadConfig()
	toggl := toggl.Toggl{Credentials: gconfig.Toggl}
	jira := jira.Jira{Credentials: gconfig.Jira}

	togglUsername, _ := toggl.GetUser()
	const padding = 3
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.StripEscape)
	fmt.Fprintf(writer,
		"%v:\t%v\t\n",
		gconfig.Jira.Domain,
		gconfig.Jira.Username,
	)
	fmt.Fprintf(writer,
		"toggl.com:\t%v\t\n",
		*togglUsername,
	)
	writer.Flush()
	fmt.Println()
	for _, task := range jira.GetAssignedTasks() {
		fmt.Fprintf(writer,
			"\t%v\t%v\t\n",
			task.Key,
			task.Summary,
		)
	}
	writer.Flush()
}

// CommandHelp : shows all available commands
func CommandHelp() {
	const padding = 3
	commands := make(map[string]string)
	commands["auth"] = "login in your Jira and Toggl account"
	commands["list"] = "list of tasks assigned to you"
	// commands["start"] = "start timetracking for selected task"
	// commands["stop"] = "stop timetracking for selected task"
	// commands["commit"] = "send your worklog to Jira"
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.StripEscape)
	fmt.Println("List of avalaible commands:")
	for command, description := range commands {
		fmt.Fprintf(writer,
			"\t%v\t%v\t\n",
			command,
			description,
		)
	}
	writer.Flush()
}
