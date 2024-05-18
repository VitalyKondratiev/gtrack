package command

import (
	"fmt"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/VitalyKondratiev/gtrack/src/config"
	"github.com/VitalyKondratiev/gtrack/src/github"
	"github.com/VitalyKondratiev/gtrack/src/helpers"
	"github.com/VitalyKondratiev/gtrack/src/jira"
	"github.com/VitalyKondratiev/gtrack/src/toggl"
	"github.com/fatih/color"
)

type displayIssue struct {
	Key            string
	Summary        string
	UncommitedTime int
	TrackingStatus string
}

type listCommand struct {
}

func ListCommand() *listCommand {
	return &listCommand{}
}

func (cmd *listCommand) Execute() {
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
		displayIssues := make(map[string]displayIssue)
		for _, issue := range _jira.GetAssignedIssues() {
			displayIssues[issue.Key] = displayIssue{
				Key:            issue.Key,
				Summary:        issue.Summary,
				TrackingStatus: "",
				UncommitedTime: 0,
			}
		}
		fmt.Fprintf(writer,
			"\t%v: %v",
			cyan(helpers.GetFormattedDomain(_jira.Config.Domain)),
			_jira.Config.Username,
		)
		writer.Flush()
		fmt.Println()
		totalUncommitedTime := 0
		uncommitedTasksCount := 0
		var notAssignedKeys []string
		for _, _timeEntry := range _toggl.GetTimeEntries() {
			if !_timeEntry.IsUncommitedEntry() || !_timeEntry.IsJiraDomainEntry(_jira.Config.Domain) {
				continue
			}
			uncommitedTasksCount += 1
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
				displayIssues[_timeEntry.Description] = displayIssue{
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
				_displayIssue := displayIssues[unassignedIssue.Key]
				_displayIssue.Summary = unassignedIssue.Summary
				displayIssues[unassignedIssue.Key] = _displayIssue
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
				_displayIssue := displayIssues[key]
				fmt.Fprintf(writer,
					"\t%v\t%.35v...\t%v\t%v",
					_displayIssue.Key,
					_displayIssue.Summary,
					helpers.FormatTimeFromUnix(_displayIssue.UncommitedTime),
					_displayIssue.TrackingStatus,
				)
				if keyIndex < len(keys)-1 {
					fmt.Fprintf(writer, "\t\n")
				}
			}
			writer.Flush()
			fmt.Println()
			fmt.Fprintf(writer,
				"\t%s %s: %v (%d tasks)\n",
				yellow("Uncommited time on"),
				yellow(_jira.Config.Domain),
				bold(helpers.FormatTimeFromUnix(totalUncommitedTime)),
				uncommitedTasksCount,
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
