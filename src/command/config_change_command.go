package command

import (
	"fmt"
	"os"

	"github.com/VitalyKondratiev/gtrack/src/helpers"
)

type changeConfigCommand struct {
}

func ChangeConfigCommand() *changeConfigCommand {
	return &changeConfigCommand{}
}

func (cmd *changeConfigCommand) Execute() {
	config := ConfigGetCommand(false).Execute()

	if len(config.Jira) == 0 {
		jiraInstance := JiraAccountLoginCommand().Execute()
		togglInstance := TogglAccountLoginCommand().Execute()
		config.SetConfig(jiraInstance.Config, togglInstance.Config).SaveConfig()
	} else {
		switch choice := cmd.getUserChoice(); choice {
		case 0:
			JiraAccountChangeCommand().Execute()
		case 1:
			jiraInstance := JiraAccountLoginCommand().Execute()
			config.Jira = append(config.Jira, jiraInstance.Config)
			config.SaveConfig()
		case 2:
			ConfigRemoveCommand().Execute()
		}
	}
}

func (cmd *changeConfigCommand) getUserChoice() int {
	user_input, err := helpers.GetVariant(
		"You already authorized in gtrack, select action",
		[]string{
			"Change existing Jira account",
			"Add one more Jira account",
			"Remove config (after this you need run auth again)",
		},
		"{{ . }} ",
	)
	if err != nil {
		fmt.Println("Error getting user input:", err)
		os.Exit(1)
	}
	return user_input
}
