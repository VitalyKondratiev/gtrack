package command

import (
	"fmt"
	"os"
	"text/tabwriter"
)

type helpCommand struct {
}

func HelpCommand() *helpCommand {
	return &helpCommand{}
}

func (cmd *helpCommand) Execute() {
	commands := make(map[string]string)
	commands["auth"] = "login in your Jira and Toggl account"
	commands["list"] = "list of tasks assigned to you"
	commands["start"] = "start timetracking for selected task"
	commands["stop"] = "stop timetracking for selected task"
	commands["commit"] = "send your worklog to Jira"
	commands["update"] = "update binary to latest version"

	const padding = 3
	writer := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.StripEscape)
	fmt.Println("List of available commands:")
	for command, description := range commands {
		fmt.Fprintf(writer,
			"\t%v\t%v\t\n",
			command,
			description,
		)
	}
	writer.Flush()
}
