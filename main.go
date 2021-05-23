package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"./src/config"
	"./src/helpers"
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
	case "start":
		CommandStart()
	case "stop":
		CommandStop()
	case "help":
		gconfig := (config.GlobalConfig{}).LoadConfig()
		fmt.Print(gconfig.Jira.Username)
		CommandHelp()
	default:
		CommandHelp()
	}
}

func CommandAuth() {
	jira := (jira.Jira{}).SetConfig()
	if !jira.IsLoggedIn() {
		return
	}
	toggl := (toggl.Toggl{}).SetConfig()
	if !toggl.IsLoggedIn() {
		return
	}
	(config.GlobalConfig{}).SetConfig(jira.Config, toggl.Config).SaveConfig()
}

func CommandList() {
	gconfig := (config.GlobalConfig{}).LoadConfig()
	toggl := toggl.Toggl{Config: gconfig.Toggl}
	jira := jira.Jira{Config: gconfig.Jira}

	togglUsername, _ := toggl.GetUser()
	const padding = 3
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.StripEscape)
	fmt.Fprintf(writer,
		"\t%v:\t%v\t\n",
		gconfig.Jira.Domain,
		gconfig.Jira.Username,
	)
	fmt.Fprintf(writer,
		"\ttoggl.com:\t%v\t\n",
		*togglUsername,
	)
	writer.Flush()
	fmt.Println()
	fmt.Fprintf(writer,
		"\t%v\t%v\t%v\t%v\t\n",
		"Issue key",
		"Issue summary",
		"Uncommited time",
		"Tracking status",
	)
	fmt.Fprintf(writer,
		"\t%v\t%v\t%v\t%v\t\n",
		"--------",
		"-------------",
		"---------------",
		"---------------",
	)

	timeEntries := toggl.GetTimeEntries()
	totalUncommitedTime := 0
	for _, issue := range jira.GetAssignedIssues() {
		trackingStatus := "-"
		uncommitedTime := 0
		for _, _timeEntry := range timeEntries {
			if _timeEntry.Description == issue.Key && _timeEntry.IsUncommitedEntry() {
				if _timeEntry.Duration > 0 {
					uncommitedTime += _timeEntry.Duration
				} else {
					startTime := time.Unix(int64(_timeEntry.Duration)*-1, 0)
					diff := time.Since(startTime)
					uncommitedTime += int(diff.Seconds())
					trackingStatus = "current"
				}
			}
		}
		totalUncommitedTime += uncommitedTime
		fmt.Fprintf(writer,
			"\t%v\t%.30v...\t%v\t%v\t\n",
			issue.Key,
			issue.Summary,
			helpers.FormatTimeFromUnix(uncommitedTime),
			trackingStatus,
		)
	}
	writer.Flush()
	fmt.Println()
	fmt.Fprintf(writer,
		"\tUncommited time:\t%v\t\n",
		helpers.FormatTimeFromUnix(totalUncommitedTime),
	)
	writer.Flush()
}

func CommandStart() {
	gconfig := (config.GlobalConfig{}).LoadConfig()
	jira := jira.Jira{Config: gconfig.Jira}
	toggl := toggl.Toggl{Config: gconfig.Toggl}
	issue := jira.SelectIssue()
	toggl.StartIssueTracking(issue.ProjectKey, issue.Key)
}

func CommandStop() {
	gconfig := (config.GlobalConfig{}).LoadConfig()
	toggl := toggl.Toggl{Config: gconfig.Toggl}
	toggl.StopIssueTracking()
}

// CommandHelp : shows all available commands
func CommandHelp() {
	const padding = 3
	commands := make(map[string]string)
	commands["auth"] = "login in your Jira and Toggl account"
	commands["list"] = "list of tasks assigned to you"
	commands["start"] = "start timetracking for selected task"
	commands["stop"] = "stop timetracking for selected task"
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
