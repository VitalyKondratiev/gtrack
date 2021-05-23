package main

import (
	"fmt"
	"os"
	"text/tabwriter"

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
	case "commit":
		CommandCommit()
	case "help":
		gconfig := (config.GlobalConfig{}).LoadConfig()
		fmt.Print(gconfig.Jira.Username)
		CommandHelp()
	default:
		CommandHelp()
	}
}

func CommandAuth() {
	_jira := (jira.Jira{}).SetConfig()
	if !_jira.IsLoggedIn() {
		return
	}
	_toggl := (toggl.Toggl{}).SetConfig()
	if !_toggl.IsLoggedIn() {
		return
	}
	(config.GlobalConfig{}).SetConfig(_jira.Config, _toggl.Config).SaveConfig()
}

func CommandList() {
	gconfig := (config.GlobalConfig{}).LoadConfig()
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	_jira := jira.Jira{Config: gconfig.Jira}

	togglUsername, _ := _toggl.GetUser()
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

	timeEntries := _toggl.GetTimeEntries()
	totalUncommitedTime := 0
	for _, issue := range _jira.GetAssignedIssues() {
		trackingStatus := "-"
		uncommitedTime := 0
		for _, _timeEntry := range timeEntries {
			if _timeEntry.Description == issue.Key && _timeEntry.IsUncommitedEntry() {
				uncommitedTime += _timeEntry.GetDuration()
				if _timeEntry.IsCurrent() {
					trackingStatus = "current"
				} else {
					trackingStatus = "-"
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
	_jira := jira.Jira{Config: gconfig.Jira}
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	issue := _jira.SelectIssue()
	_toggl.StartIssueTracking(issue.ProjectKey, issue.Key)
}

func CommandStop() {
	gconfig := (config.GlobalConfig{}).LoadConfig()
	toggl := toggl.Toggl{Config: gconfig.Toggl}
	toggl.StopIssueTracking()
}

func CommandCommit() {
	gconfig := (config.GlobalConfig{}).LoadConfig()
	_jira := jira.Jira{Config: gconfig.Jira}
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	timeEntries := _toggl.GetTimeEntries()
	var uncommitedTimeEntries []toggl.TogglTimeEntry
	for _, _timeEntry := range timeEntries {
		if _timeEntry.IsUncommitedEntry() {
			uncommitedTimeEntries = append(uncommitedTimeEntries, _timeEntry)
		}
	}
	uncommitedCount := len(uncommitedTimeEntries)
	if uncommitedCount == 0 {
		fmt.Println("Notning to commit!")
	}
	togglState, durations, startTimes := _toggl.CommitIssues(uncommitedTimeEntries, true)
	if togglState {
		issues := make(map[string]jira.JiraIssue)
		for _, _issue := range _jira.GetAssignedIssues() {
			if durations[_issue.Key] != 0 {
				issues[_issue.Key] = _issue
			}
		}
		jiraState, rejectedWorklogs := _jira.CommitIssues(issues, durations, startTimes)
		if !jiraState {
			fmt.Println(rejectedWorklogs)
			var rejectedEntries []toggl.TogglTimeEntry
			for _, _timeEntry := range uncommitedTimeEntries {
				for _, rejectedKey := range rejectedWorklogs {
					if _timeEntry.Description == rejectedKey {
						rejectedEntries = append(uncommitedTimeEntries, _timeEntry)
					}
				}
			}
			_toggl.CommitIssues(rejectedEntries, false)
		} else {
			fmt.Printf("Successfully commited %v issues!\n", uncommitedCount)
		}
	}
}

// CommandHelp : shows all available commands
func CommandHelp() {
	const padding = 3
	commands := make(map[string]string)
	commands["auth"] = "login in your Jira and Toggl account"
	commands["list"] = "list of tasks assigned to you"
	commands["start"] = "start timetracking for selected task"
	commands["stop"] = "stop timetracking for selected task"
	commands["commit"] = "send your worklog to Jira"
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
