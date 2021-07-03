package main

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"./src/config"
	"./src/helpers"
	"./src/jira"
	"./src/toggl"
)

type DisplayIssue struct {
	Key            string
	Summary        string
	UncommitedTime int
	TrackingStatus string
}

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
		CommandHelp()
	default:
		CommandHelp()
	}
}

func CommandAuth() {
	gconfig := (config.GlobalConfig{}).LoadConfig(false)
	if len(gconfig.Jira) != 0 {
		switch choice := gconfig.ChangeConfiguration(); choice {
		case 0:
			// Change existing Jira account
			// TODO: Select Jira instance and replace it
			fmt.Println("Not realized yet...")
		case 1:
			// Add one more Jira account
			_jira := (jira.Jira{}).SetConfig()
			if !_jira.IsLoggedIn() {
				os.Exit(1)
			}
			gconfig.Jira = append(gconfig.Jira, _jira.Config)
			gconfig.SaveConfig()
		case 2:
			// Remove config
			config.RemoveConfig()
		}
		os.Exit(0)
	} else {
		_jira := (jira.Jira{}).SetConfig()
		if !_jira.IsLoggedIn() {
			os.Exit(1)
		}
		_toggl := (toggl.Toggl{}).SetConfig()
		if !_toggl.IsLoggedIn() {
			os.Exit(1)
		}
		(config.GlobalConfig{}).SetConfig(_jira.Config, _toggl.Config).SaveConfig()
	}
}

func CommandList() {
	gconfig := (config.GlobalConfig{}).LoadConfig(true)
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	var _jira jira.Jira
	if len(gconfig.Jira) > 1 {
		jiraIndex := gconfig.SelectJiraInstance()
		_jira = jira.Jira{Config: gconfig.Jira[jiraIndex]}
	} else {
		_jira = jira.Jira{Config: gconfig.Jira[0]}
	}
	togglUsername, _ := _toggl.GetUser()
	const padding = 3
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.StripEscape)
	fmt.Fprintf(writer,
		"\t%v:\t%v\t\n",
		helpers.GetFormattedDomain("toggl.com"),
		*togglUsername,
	)
	fmt.Fprintf(writer,
		"\t%v:\t%v\t\n",
		helpers.GetFormattedDomain(_jira.Config.Domain),
		_jira.Config.Username,
	)
	writer.Flush()
	fmt.Println()
	fmt.Fprintf(writer,
		"\t%v\t%v\t%v\t%v\t\n",
		"Issue key",
		"Issue summary",
		"Time",
		"Tracking status",
	)
	fmt.Fprintf(writer,
		"\t%v\t%v\t%v\t%v\t\n",
		"--------",
		"-------------",
		"----",
		"---------------",
	)
	displayIssues := make(map[string]DisplayIssue)
	for _, issue := range _jira.GetAssignedIssues() {
		displayIssues[issue.Key] = DisplayIssue{
			Key:            issue.Key,
			Summary:        issue.Summary,
			TrackingStatus: "",
			UncommitedTime: 0,
		}
	}
	totalUncommitedTime := 0
	var notAssignedKeys []string
	for _, _timeEntry := range _toggl.GetTimeEntries() {
		if !_timeEntry.IsUncommitedEntry() || !_timeEntry.IsJiraDomainEntry(_jira.Config.Domain) {
			continue
		}
		uncommitedTime := _timeEntry.GetDuration()
		totalUncommitedTime += uncommitedTime
		if val, ok := displayIssues[_timeEntry.Description]; ok {
			val.UncommitedTime += uncommitedTime
			if len(val.TrackingStatus) == 0 {
				val.TrackingStatus = "uncommited"
			}
			if _timeEntry.IsCurrent() {
				val.TrackingStatus = "current"
			}
			displayIssues[_timeEntry.Description] = val
		} else {
			trackingStatus := "uncommited"
			if _timeEntry.IsCurrent() {
				trackingStatus = "current"
			}
			displayIssues[_timeEntry.Description] = DisplayIssue{
				Key:            _timeEntry.Description,
				Summary:        "",
				TrackingStatus: fmt.Sprintf("%s (not assigned)", trackingStatus),
				UncommitedTime: uncommitedTime,
			}
			notAssignedKeys = append(notAssignedKeys, _timeEntry.Description)
		}
	}
	if len(notAssignedKeys) > 0 {
		for _, unassignedIssue := range _jira.GetIssuesByField(notAssignedKeys, "key") {
			displayIssue := displayIssues[unassignedIssue.Key]
			displayIssue.Summary = unassignedIssue.Summary
			displayIssues[unassignedIssue.Key] = displayIssue
		}
	}
	keys := make([]string, 0, len(displayIssues))
	for key := range displayIssues {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		displayIssue := displayIssues[key]
		fmt.Fprintf(writer,
			"\t%v\t%.30v...\t%v\t%v\t\n",
			displayIssue.Key,
			displayIssue.Summary,
			helpers.FormatTimeFromUnix(displayIssue.UncommitedTime),
			displayIssue.TrackingStatus,
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
	gconfig := (config.GlobalConfig{}).LoadConfig(true)
	var _jira jira.Jira
	if len(gconfig.Jira) > 1 {
		jiraIndex := gconfig.SelectJiraInstance()
		_jira = jira.Jira{Config: gconfig.Jira[jiraIndex]}
	} else {
		_jira = jira.Jira{Config: gconfig.Jira[0]}
	}
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	var issue jira.JiraIssue
	if len(os.Args) < 3 {
		issue = _jira.SelectIssue()
	} else {
		issue = _jira.GetIssueByKey(os.Args[2])
	}
	_toggl.StartIssueTracking(issue.ProjectKey, issue.Key, _jira.Config.Domain)
}

func CommandStop() {
	gconfig := (config.GlobalConfig{}).LoadConfig(true)
	toggl := toggl.Toggl{Config: gconfig.Toggl}
	toggl.StopIssueTracking()
}

func CommandCommit() {
	gconfig := (config.GlobalConfig{}).LoadConfig(true)
	var _jira jira.Jira
	if len(gconfig.Jira) > 1 {
		jiraIndex := gconfig.SelectJiraInstance()
		_jira = jira.Jira{Config: gconfig.Jira[jiraIndex]}
	} else {
		_jira = jira.Jira{Config: gconfig.Jira[0]}
	}
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	timeEntries := _toggl.GetTimeEntries()
	var uncommitedTimeEntries []toggl.TogglTimeEntry
	for _, _timeEntry := range timeEntries {
		if _timeEntry.IsUncommitedEntry() && _timeEntry.IsJiraDomainEntry(_jira.Config.Domain) {
			uncommitedTimeEntries = append(uncommitedTimeEntries, _timeEntry)
		}
	}
	uncommitedCount := len(uncommitedTimeEntries)
	if uncommitedCount == 0 {
		fmt.Println("Notning to commit!")
		return
	}
	togglState, issueKeys, durations, startTimes := _toggl.CommitIssues(uncommitedTimeEntries, true)
	if togglState {
		issues := _jira.GetIssuesByField(issueKeys, "key")
		durationsByIssue := make(map[string][]int)
		startTimesByIssue := make(map[string][]time.Time)
		for index, duration := range durations {
			issueKey := issueKeys[index]
			durationsByIssue[issueKey] = append(durationsByIssue[issueKey], duration)
		}
		for index, startTime := range startTimes {
			issueKey := issueKeys[index]
			startTimesByIssue[issueKey] = append(startTimesByIssue[issueKey], startTime)
		}
		jiraState, rejected := _jira.CommitIssues(issues, durationsByIssue, startTimesByIssue)
		if !jiraState {
			var rejectedEntries []toggl.TogglTimeEntry
			for _, timeEntry := range uncommitedTimeEntries {
				for _, duration := range rejected[timeEntry.Description] {
					if timeEntry.GetDuration() == duration {
						rejectedEntries = append(rejectedEntries, timeEntry)
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
