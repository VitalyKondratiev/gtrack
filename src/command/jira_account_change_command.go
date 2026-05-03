package command

import (
	"fmt"
	"os"

	"github.com/VitalyKondratiev/gtrack/src/config"
	"github.com/VitalyKondratiev/gtrack/src/helpers"
)

type jiraAccountChangeCommand struct {
}

func JiraAccountChangeCommand() *jiraAccountChangeCommand {
	return &jiraAccountChangeCommand{}
}

func (cmd *jiraAccountChangeCommand) Execute() {
	cfg := ConfigGetCommand(true).Execute()
	jiraIndex := 0
	if len(cfg.Jira) > 1 {
		jiraIndex = cfg.SelectJiraInstance([]int{})
	}
	switch choice := cmd.getUserChoice(); choice {
	case 0:
		cmd.updateJiraDomain(&cfg, jiraIndex)
	case 1:
		cmd.updateJiraToken(&cfg, jiraIndex)
	}

	cfg.SaveConfig()
}

func (cmd *jiraAccountChangeCommand) updateJiraDomain(cfg *config.GlobalConfig, jiraIndex int) {
	optionValue, err := helpers.GetString(
		fmt.Sprintf("Enter domain of your Jira instance (%v)", cfg.Jira[jiraIndex].Domain),
		false,
	)
	if err != nil {
		os.Exit(0)
	}
	cfg.Jira[jiraIndex].Domain = optionValue
}

func (cmd *jiraAccountChangeCommand) updateJiraToken(cfg *config.GlobalConfig, jiraIndex int) {
	optionValue, err := helpers.GetString(
		"Enter Jira access token",
		true,
	)
	if err != nil {
		os.Exit(0)
	}
	cfg.Jira[jiraIndex].Token = optionValue
	cfg.Jira[jiraIndex].Cookies = nil
}

func (cmd *jiraAccountChangeCommand) getUserChoice() int {
	user_input, err := helpers.GetVariant(
		"Select value for changing",
		[]string{
			"Domain",
			"Access token",
		},
		"{{ . }} ",
	)
	if err != nil {
		os.Exit(0)
	}
	return user_input
}
