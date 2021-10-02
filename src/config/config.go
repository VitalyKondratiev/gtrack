package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"../helpers"

	"github.com/Jeffail/gabs"
	"github.com/kirsle/configdir"
)

const userConfigName = "gtrack.json"

type JiraConfig struct {
	Domain   string `json:"domain"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type TogglConfig struct {
	ApiKey      string `json:"api_key"`
	WorkspaceId int    `json:"workspace_id"`
}

type GlobalConfig struct {
	ConfigVersion int          `json:"config_version"`
	Jira          []JiraConfig `json:"jira"`
	Toggl         TogglConfig  `json:"toggl"`
}

func (config GlobalConfig) SetConfig(jiraConfig JiraConfig, togglConfig TogglConfig) GlobalConfig {
	config.ConfigVersion = 2
	config.Jira = []JiraConfig{jiraConfig}
	config.Toggl = togglConfig
	return config
}

// SaveMainConfig : save main application configuration
func (config GlobalConfig) SaveConfig() {
	configPath := configdir.LocalConfig("gtrack")
	_, err := os.Open(configPath + "/" + userConfigName)
	if err != nil {
		os.MkdirAll(configPath, os.ModePerm)
	}
	file, _ := json.MarshalIndent(config, "", "\t")
	_ = ioutil.WriteFile(configPath+"/"+userConfigName, file, 0644)
}

// LoadMainConfig : get main application configuration
func (config GlobalConfig) LoadConfig(needAuthorized bool) GlobalConfig {
	configPath := configdir.LocalConfig("gtrack")
	configFile, err := ioutil.ReadFile(configPath + "/" + userConfigName)
	if err != nil {
		if needAuthorized {
			fmt.Println("Configuration file not found, try to auth")
			os.Exit(1)
		} else {
			return config
		}
	}
	_ = json.Unmarshal([]byte(configFile), &config)
	if config.ConfigVersion == 0 {
		config = updateConfigToV2(configFile)
	}
	return config
}

func RemoveConfig() {
	configPath := configdir.LocalConfig("gtrack")
	err := os.Remove(configPath + "/" + userConfigName)
	if err != nil {
		fmt.Println("Can't remove configuration file")
		os.Exit(1)
	}
}

func updateConfigToV2(oldConfigFile []byte) GlobalConfig {
	jsonOld, _ := gabs.ParseJSON(oldConfigFile)
	jiraJson := jsonOld.Data().(map[string]interface{})["jira"].(map[string]interface{})
	jiraConfig := JiraConfig{
		Domain:   jiraJson["domain"].(string),
		Username: jiraJson["username"].(string),
		Password: jiraJson["password"].(string),
	}
	togglJson := jsonOld.Data().(map[string]interface{})["toggl"].(map[string]interface{})
	togglConfig := TogglConfig{
		ApiKey:      togglJson["apiKey"].(string),
		WorkspaceId: int(togglJson["workspaceId"].(float64)),
	}
	newConfig := GlobalConfig{
		ConfigVersion: 2,
		Jira:          []JiraConfig{jiraConfig},
		Toggl:         togglConfig,
	}
	newConfig.SaveConfig()
	return newConfig
}

func (config GlobalConfig) ChangeConfiguration() int {
	user_input, err := helpers.GetVariant(
		"You already authorized in gtrack, select action",
		[]string{"Change existing Jira account", "Add one more Jira account", "Remove config (after this you need run auth again)"},
		"{{ . }} ",
	)
	if err != nil {
		os.Exit(1)
	}
	return user_input
}

func (config GlobalConfig) SelectJiraInstance() int {
	var jiraDomains []string
	for _, jira := range config.Jira {
		jiraDomains = append(jiraDomains, helpers.GetFormattedDomain(jira.Domain))
	}
	user_input, err := helpers.GetVariant(
		"Select Jira instance",
		jiraDomains,
		"{{ . }} ",
	)
	if err != nil {
		os.Exit(1)
	}
	return user_input
}
