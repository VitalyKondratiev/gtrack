package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

const userConfigPath = "/.config/gtrack"
const userConfigName = "gtrack.json"

type JiraConfig struct {
	Domain   string `json:"domain"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type TogglConfig struct {
	ApiKey      string `json:"apiKey"`
	WorkspaceId int    `json:"workspaceId"`
}

type GlobalConfig struct {
	Jira  JiraConfig  `json:"jira"`
	Toggl TogglConfig `json:"toggl"`
}

func (config GlobalConfig) SetConfig(jiraConfig JiraConfig, togglConfig TogglConfig) GlobalConfig {
	config.Jira = jiraConfig
	config.Toggl = togglConfig
	return config
}

// SaveMainConfig : save main application configuration
func (config GlobalConfig) SaveConfig() {
	usr, _ := user.Current()
	dir := usr.HomeDir
	configPath := filepath.Join(dir, userConfigPath)
	_, err := os.Open(configPath + "/" + userConfigName)
	if err != nil {
		os.MkdirAll(configPath, os.ModePerm)
	}
	file, _ := json.MarshalIndent(config, "", "\t")
	_ = ioutil.WriteFile(configPath+"/"+userConfigName, file, 0644)
}

// LoadMainConfig : get main application configuration
func (config GlobalConfig) LoadConfig() GlobalConfig {
	usr, _ := user.Current()
	dir := usr.HomeDir
	configPath := filepath.Join(dir, userConfigPath)
	configFile, err := ioutil.ReadFile(configPath + "/" + userConfigName)
	if err != nil {

		return config
	}
	_ = json.Unmarshal([]byte(configFile), &config)
	return config
}
