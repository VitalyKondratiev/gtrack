package command

import (
	"github.com/VitalyKondratiev/gtrack/src/config"
)

type configGetCommand struct {
	needAuthorized bool
}

func ConfigGetCommand(needAuthorized bool) *configGetCommand {
	return &configGetCommand{
		needAuthorized: needAuthorized,
	}
}

func (cmd *configGetCommand) Execute() config.GlobalConfig {
	cfg := config.GlobalConfig{}
	return cfg.LoadConfig(cmd.needAuthorized)
}
