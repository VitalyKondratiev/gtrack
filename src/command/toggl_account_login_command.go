package command

import (
	"os"

	"github.com/VitalyKondratiev/gtrack/src/toggl"
)

type togglAccountLoginCommand struct {
}

func TogglAccountLoginCommand() *togglAccountLoginCommand {
	return &togglAccountLoginCommand{}
}

func (cmd *togglAccountLoginCommand) Execute() toggl.Toggl {
	togglInstance := toggl.Toggl{}
	togglInstance = togglInstance.SetConfig()

	if !togglInstance.IsLoggedIn() {
		os.Exit(1)
	}

	return togglInstance
}
