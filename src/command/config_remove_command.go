package command

import (
	"fmt"
	"os"

	"github.com/VitalyKondratiev/gtrack/src/config"
	"github.com/kirsle/configdir"
)

type configRemoveCommand struct {
}

func ConfigRemoveCommand() *configRemoveCommand {
	return &configRemoveCommand{}
}

func (cmd *configRemoveCommand) Execute() {
	configPath := configdir.LocalConfig("gtrack")
	err := os.Remove(configPath + "/" + config.UserConfigName)
	if err != nil {
		fmt.Println("Can't remove configuration file")
		os.Exit(1)
	}
}
