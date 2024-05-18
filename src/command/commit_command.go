package command

import (
	"fmt"
	"time"

	"github.com/VitalyKondratiev/gtrack/src/config"
	"github.com/VitalyKondratiev/gtrack/src/jira"
	"github.com/VitalyKondratiev/gtrack/src/toggl"
)

type commitCommand struct {
}

func CommitCommand() *commitCommand {
	return &commitCommand{}
}

func (cmd *commitCommand) Execute() {

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
