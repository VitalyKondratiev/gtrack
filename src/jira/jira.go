package jira

import (
	"fmt"
	"os"

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
