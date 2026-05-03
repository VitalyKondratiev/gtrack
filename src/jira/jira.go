package jira

import (
	"fmt"
	"time"

	"github.com/VitalyKondratiev/gtrack/src/helpers"
)

func (jira Jira) IsLoggedIn() bool {
	return jira.isLoggedIn
}

func (jira Jira) SetConfig() Jira {
	jira.isLoggedIn = false
	result, err := helpers.GetString("Enter domain of your Jira instance", false)
	if err != nil {
		return jira
	}
	jira.Config.Domain = result

	result, err = helpers.GetString("Enter your Jira access token", true)
	if err != nil {
		return jira
	}
	jira.Config.Token = result
	jira.Config.Cookies = nil

	currentUser, statusCode := jira.GetCurrentUser()
	jira.isLoggedIn = statusCode == 200

	if jira.isLoggedIn {
		if currentUser != nil {
			fmt.Printf("You sucessfully logged in %s as %s\n", jira.Config.Domain, *currentUser)
		} else {
			fmt.Printf("You sucessfully logged in %s using access token\n", jira.Config.Domain)
		}
	} else {
		helpers.LogFatal(fmt.Errorf("You not authorized, try to reauth"))
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
		helpers.LogFatal(err)
	}
	return issues[index]
}

func (jira Jira) GetIssueByKey(issueKey string) JiraIssue {
	issues := jira.GetIssuesByField([]string{issueKey}, "key")
	if len(issues) == 0 {
		return JiraIssue{}
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
