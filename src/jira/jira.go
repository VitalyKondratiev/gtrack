package jira

import (
	"fmt"
	"os"
	"time"

	"github.com/manifoldco/promptui"
)

func (jira Jira) IsLoggedIn() bool {
	return jira.isLoggedIn
}

func (jira Jira) SetConfig() Jira {
	jira.isLoggedIn = false
	prompt := promptui.Prompt{
		Label: "Enter domain of your Jira instance",
	}
	result, err := prompt.Run()
	if err != nil {
		return jira
	}
	jira.Config.Domain = result

	l_prompt := promptui.Prompt{
		Label: "Enter your username",
	}
	result, err = l_prompt.Run()
	if err != nil {
		return jira
	}
	jira.Config.Username = result

	p_prompt := promptui.Prompt{
		Label: "Enter your password",
	}
	result, err = p_prompt.Run()
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
	w_prompt := promptui.Select{
		Label: "Select issue",
		Items: issues,
		Templates: &promptui.SelectTemplates{
			Active:   "{{ .Key | green }} - {{ printf \"%.35s...\" .Summary | green }}",
			Inactive: "{{ .Key | white }} - {{ printf \"%.35s...\" .Summary | white }}",
		},
		HideSelected: true,
	}
	index, _, err := w_prompt.Run()
	if err != nil {
		os.Exit(1)
	}
	return issues[index]
}

func (jira Jira) GetIssueByKey(issueKey string) JiraIssue {
	issues := jira.GetIssuesByField([]string{issueKey}, "key")
	if len(issues) == 0 {
		fmt.Println("Issue not found!")
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
