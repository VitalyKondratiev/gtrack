package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
)

const userConfigPath = "/.config/gtrack"
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
func (config GlobalConfig) LoadConfig(needAuthorized bool) GlobalConfig {
	usr, _ := user.Current()
	dir := usr.HomeDir
	configPath := filepath.Join(dir, userConfigPath)
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
	usr, _ := user.Current()
	dir := usr.HomeDir
	configPath := filepath.Join(dir, userConfigPath)
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
	w_prompt := promptui.Select{
		Label: "You already authorized in gtrack, select action",
		Items: []string{"Change existing Jira account", "Add one more Jira account", "Remove config (after this you need run auth again)"},
		Templates: &promptui.SelectTemplates{
			Active:   "{{ . | green  }}",
			Inactive: "{{ . | white  }}",
		},
	}
	user_input, _, err := w_prompt.Run()
	if err != nil {
		os.Exit(1)
	}
	return user_input
}
