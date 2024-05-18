package command

import (
	"github.com/VitalyKondratiev/gtrack/src/toggl"
)

type trackingStopCommand struct {
}

func TrackingStopCommand() *trackingStopCommand {
	return &trackingStopCommand{}
}

func (cmd *trackingStopCommand) Execute() {
	config := ConfigGetCommand(true).Execute()
	togglInstance := toggl.Toggl{Config: config.Toggl}
	togglInstance.StopIssueTracking()
}
