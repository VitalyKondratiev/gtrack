package command

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/VitalyKondratiev/gtrack/src/config"
	"github.com/kirsle/configdir"
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
	configPath := configdir.LocalConfig("gtrack")
	configFile, err := os.ReadFile(configPath + "/" + config.UserConfigName)
	if err != nil {
		if cmd.needAuthorized {
			fmt.Println("Configuration file not found, try to auth")
			os.Exit(1)
		} else {
			return cfg
		}
	}
	_ = json.Unmarshal([]byte(configFile), &cfg)
	return cfg
}
