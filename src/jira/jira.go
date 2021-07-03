package jira

import (
	"fmt"
	"os"
	"time"

	"../helpers"
)

func (jira Jira) IsLoggedIn() bool {
	return jira.isLoggedIn
}

func (jira Jira) SetConfig() Jira {
	jira.isLoggedIn = false
	result, err := helpers.GetString("Enter domain of your Jira instance")
	if err != nil {
		return jira
	}
	jira.Config.Domain = result

	result, err = helpers.GetString("Enter your username")
	if err != nil {
		return jira
	}
	jira.Config.Username = result

	result, err = helpers.GetString("Enter your password")
	if err != nil {
		return jira
	}
	jira.Config.Password = result

	jira = jira.authenticate()

	if jira.isLoggedIn {
		fmt.Printf("You sucessfully logged in %s as %s\n", jira.Config.Domain, jira.Config.Username)
	}

	return jira
}

func (jira Jira) SelectIssue() JiraIssue {
	issues := jira.GetAssignedIssues()
	index, err := helpers.GetVariant(
		"Select issue",
		issues,
		"{{ .Key }} {{ printf \"-\" }} {{ printf \"%.35s...\" .Summary }}",
	)
	if err != nil {
		os.Exit(1)
	}
	return issues[index]
}

func (jira Jira) GetIssueByKey(issueKey string) JiraIssue {
	issues := jira.GetIssuesByField([]string{issueKey}, "key")
	if len(issues) == 0 {
		fmt.Printf("Issue '%v' not found in %v\n", issueKey, helpers.GetFormattedDomain(jira.Config.Domain))
		os.Exit(1)
	}
	return issues[0]
}

func (jira Jira) CommitIssues(issues []JiraIssue, durationsByIssues map[string][]int, startTimesByIssues map[string][]time.Time) (bool, map[string][]int) {
	rejectedWorklogs := make(map[string][]int)
	if len(durationsByIssues) != len(startTimesByIssues) {
		for _, issue := range issues {
			rejectedWorklogs[issue.Key] = durationsByIssues[issue.Key]
		}
		return false, rejectedWorklogs
	}
	for _, issue := range issues {
		if len(durationsByIssues[issue.Key]) != len(startTimesByIssues[issue.Key]) {
			rejectedWorklogs[issue.Key] = durationsByIssues[issue.Key]
			continue
		}
		issueState := true
		for index, duration := range durationsByIssues[issue.Key] {
			if duration < 60 {
				duration = 60
			}
			issueState = jira.SetWorklogEntry(issue.Id, duration, startTimesByIssues[issue.Key][index])
			if !issueState {
				rejectedWorklogs[issue.Key] = append(rejectedWorklogs[issue.Key], duration)
			}
		}
	}
	return len(rejectedWorklogs) == 0, rejectedWorklogs
}
