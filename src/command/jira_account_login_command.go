package command

import (
	"os"

	"github.com/VitalyKondratiev/gtrack/src/jira"
)

type jiraAccountLoginCommand struct {
}

func JiraAccountLoginCommand() *jiraAccountLoginCommand {
	return &jiraAccountLoginCommand{}
}

func (cmd *jiraAccountLoginCommand) Execute() jira.Jira {
	jiraInstance := jira.Jira{}
	jiraInstance = jiraInstance.SetConfig()

	if !jiraInstance.IsLoggedIn() {
		os.Exit(1)
	}

	return jiraInstance
}
