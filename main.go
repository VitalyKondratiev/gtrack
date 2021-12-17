package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/VitalyKondratiev/gtrack/src/config"
	"github.com/VitalyKondratiev/gtrack/src/github"
	"github.com/VitalyKondratiev/gtrack/src/helpers"
	"github.com/VitalyKondratiev/gtrack/src/jira"
	"github.com/VitalyKondratiev/gtrack/src/toggl"

	"github.com/fatih/color"
	"gopkg.in/src-d/go-git.v4"
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
	case "update":
		CommandUpdate()
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
	togglUsername, _ := _toggl.GetUser()
	const padding = 3
	writer := tabwriter.NewWriter(color.Output, 0, 0, padding, ' ', tabwriter.StripEscape)
	blue := color.New(color.FgHiBlue).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	red := color.New(color.FgHiRed).SprintFunc()
	if duration, _ := time.ParseDuration("6h"); gconfig.UpdateNotify.Add(duration).Sub(time.Now()) < 0 {
		hasUpdate, githubRelease := github.Github{}.HasUpdate()
		if hasUpdate {
			fmt.Fprintf(writer, "\t%v\n\n", red("Update to ", githubRelease.Version, " available, run 'gtrack update' for new version "))
			gconfig.UpdateNotify = time.Now()
			gconfig.SaveConfig()
		}
	}
	fmt.Fprintf(writer,
		"\t%v:\t%v\t\n",
		blue(helpers.GetFormattedDomain("toggl.com")),
		*togglUsername,
	)
	for jiraConfigIndex, jiraConfig := range gconfig.Jira {
		var _jira = jira.Jira{Config: jiraConfig}
		fmt.Fprintf(writer,
			"\t%v: %v",
			cyan(helpers.GetFormattedDomain(_jira.Config.Domain)),
			_jira.Config.Username,
		)
		writer.Flush()
		fmt.Println()
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
		if len(displayIssues) > 0 {
			keys := make([]string, 0, len(displayIssues))
			for key := range displayIssues {
				keys = append(keys, key)
			}
			sort.Strings(keys)

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
			for keyIndex, key := range keys {
				displayIssue := displayIssues[key]
				fmt.Fprintf(writer,
					"\t%v\t%.35v...\t%v\t%v",
					displayIssue.Key,
					displayIssue.Summary,
					helpers.FormatTimeFromUnix(displayIssue.UncommitedTime),
					displayIssue.TrackingStatus,
				)
				if keyIndex < len(keys)-1 {
					fmt.Fprintf(writer, "\t\n")
				}
			}
			writer.Flush()
			fmt.Println()
			fmt.Fprintf(writer,
				"\t%s %s: %v\n",
				yellow("Uncommited time on"),
				yellow(_jira.Config.Domain),
				bold(helpers.FormatTimeFromUnix(totalUncommitedTime)),
			)
		} else {
			fmt.Fprintf(writer,
				"\tYou don't have assigned or uncommited task on %s\n",
				_jira.Config.Domain,
			)
		}
		if jiraConfigIndex < len(gconfig.Jira)-1 {
			fmt.Fprintf(writer, "\t\n")
		}
	}
	writer.Flush()
}

func CommandStart() {
	gconfig := (config.GlobalConfig{}).LoadConfig(true)
	var _jira jira.Jira
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	var issue jira.JiraIssue
	if len(os.Args) < 3 {
		if len(gconfig.Jira) > 1 {
			jiraIndex := gconfig.SelectJiraInstance([]int{})
			_jira = jira.Jira{Config: gconfig.Jira[jiraIndex]}
		} else {
			_jira = jira.Jira{Config: gconfig.Jira[0]}
		}
		issue = _jira.SelectIssue()
	} else {
		var issueKey string
		if os.Args[2] == "-g" {
			dir, _ := os.Getwd()
			dir, isGitRoot := helpers.TryGetGitDirectory(dir)
			if !isGitRoot {
				fmt.Println("Can't find .git folder in this path directories")
				os.Exit(1)
			}
			repo, err := git.PlainOpen(dir)
			if err != nil {
				os.Exit(1)
			}
			h, err := repo.Head()
			if err != nil {
				os.Exit(1)
			}
			issueKey = strings.Replace(h.Name().String(), "refs/heads/", "", -1)
		} else {
			issueKey = os.Args[2]
		}
		var lastJira jira.Jira
		var lastIssue jira.JiraIssue
		issuesByInstances := make(map[int]jira.JiraIssue)
		for jiraIndex, jiraConfig := range gconfig.Jira {
			lastJira = jira.Jira{Config: jiraConfig}
			lastIssue = lastJira.GetIssueByKey(issueKey)
			if lastIssue.Id != 0 {
				issuesByInstances[jiraIndex] = lastIssue
			}
		}
		if len(issuesByInstances) == 0 {
			fmt.Printf("Issue '%v' not found in available Jira instances\n", issueKey)
			os.Exit(1)
		} else {
			var jiraIndexes []int
			var issues []jira.JiraIssue
			for jiraIndex, _issue := range issuesByInstances {
				jiraIndexes = append(jiraIndexes, jiraIndex)
				issues = append(issues, _issue)
			}
			var jiraIndex int
			if len(issuesByInstances) > 1 {
				jiraIndex = gconfig.SelectJiraInstance(jiraIndexes)
				issue = issues[jiraIndex]
			} else {
				jiraIndex = jiraIndexes[0]
				issue = issues[0]
			}
			_jira = jira.Jira{Config: gconfig.Jira[jiraIndex]}
		}
	}
	_toggl.StartIssueTracking(issue.ProjectKey, issue.Key, issue.Summary, _jira.Config.Domain)
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
		jiraIndex := gconfig.SelectJiraInstance([]int{})
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

// CommandHelp : try update executable
func CommandUpdate() {
	github := github.Github{}
	hasUpdate, githubRelease := github.HasUpdate()
	if hasUpdate {
		variant, err := helpers.GetVariant(
			fmt.Sprintf("Update available to %s (see in browser: %s)", githubRelease.Version, githubRelease.ReleasePage),
			[]string{"Update now", "Cancel"},
			"{{ . }} ",
		)
		if err != nil || variant == 1 {
			os.Exit(1)
		}
		fmt.Printf("Downloading release %s...\n", githubRelease.Version)
		isFileDownloaded, err := github.DownloadRelease(githubRelease.DownloadableFiles)
		if err != nil || !isFileDownloaded {
			fmt.Printf("Download release error :(")
		}
		isUpdated := github.Update()
		fmt.Printf("Replacing binary file...\n")
		if isUpdated {
			fmt.Printf("Succesfully updated to %s\n", githubRelease.Version)
		} else {
			panic(err)
		}
	} else {
		fmt.Println("No available updates")
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
	commands["update"] = "update binary to latest version"
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
