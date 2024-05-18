package command

import (
	"fmt"
	"os"

	"github.com/VitalyKondratiev/gtrack/src/config"
	"github.com/VitalyKondratiev/gtrack/src/jira"
	"github.com/VitalyKondratiev/gtrack/src/toggl"
)

type trackingStartCommand struct {
	taskId    string
	useBranch bool
}

func TrackingStartCommand(taskId string, useBranch bool) *trackingStartCommand {
	return &trackingStartCommand{
		taskId:    taskId,
		useBranch: useBranch,
	}
}

func (cmd *trackingStartCommand) Execute() {
	gconfig := (config.GlobalConfig{}).LoadConfig(true)
	var _jira jira.Jira
	_toggl := toggl.Toggl{Config: gconfig.Toggl}
	var issue jira.JiraIssue
	if cmd.taskId == "" && !cmd.useBranch {
		if len(gconfig.Jira) > 1 {
			jiraIndex := gconfig.SelectJiraInstance([]int{})
			_jira = jira.Jira{Config: gconfig.Jira[jiraIndex]}
		} else {
			_jira = jira.Jira{Config: gconfig.Jira[0]}
		}
		issue = _jira.SelectIssue()
	} else {
		var issueKey string
		if cmd.useBranch {
			issueKey = GetCurrentBranchCommand().Execute()
		} else {
			issueKey = cmd.taskId
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
