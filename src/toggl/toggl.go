package toggl

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

const tagUncommitedName = "gtrack_uncommited"

func (togglTimeEntry TogglTimeEntry) IsUncommitedEntry() bool {
	for _, tag := range togglTimeEntry.Tags {
		if tag == tagUncommitedName {
			return true
		}
	}
	return false
}

func (toggl Toggl) IsLoggedIn() bool {
	return toggl.isLoggedIn
}

func (toggl Toggl) SetConfig() Toggl {
	prompt := promptui.Prompt{
		Label: "Enter API key of your Toggl.com account",
	}
	result, err := prompt.Run()
	if err != nil {
		toggl.isLoggedIn = false
		return toggl
	}
	toggl.Config.ApiKey = result

	username, statusCode := toggl.GetUser()
	if statusCode == 403 {
		toggl.isLoggedIn = false
		fmt.Printf("API key %s is not correct\n", toggl.Config.ApiKey)
	}
	if statusCode == 200 {
		toggl.isLoggedIn = true
		fmt.Printf("You sucessfully logged in Toggl as %s\n", *username)
	}

	if toggl.isLoggedIn {
		workspaces := toggl.GetWorkspaces()
		w_prompt := promptui.Select{
			Label: "Select workspace",
			Items: workspaces,
			Templates: &promptui.SelectTemplates{
				Active:   "{{ .Name | green }}",
				Inactive: "{{ .Name | white }}",
			},
			HideSelected: true,
		}
		index, _, err := w_prompt.Run()
		if err != nil {
			toggl.isLoggedIn = false
			return toggl
		}
		toggl.Config.WorkspaceId = workspaces[index].Id
	}

	return toggl
}

func (toggl Toggl) renewTagOnWorkspace() {
	tags := toggl.GetWorkspaceTags()
	var workspaceTag TogglWorkspaceTag
	for _, _workspaceTag := range tags {
		if tagUncommitedName == _workspaceTag.Name {
			workspaceTag = _workspaceTag
		}
	}
	if workspaceTag.Id == 0 {
		_ = toggl.CreateTag(tagUncommitedName)
	}
}

func (toggl Toggl) StartIssueTracking(projectKey string, taskName string) {
	toggl.renewTagOnWorkspace()
	var project TogglProject
	for _, _project := range toggl.GetProjects() {
		if projectKey == _project.Name {
			project = _project
		}
	}
	if project.Id == 0 {
		project = toggl.CreateProject(projectKey)
	}
	timeEntry := toggl.StartTimeEntry(project.Id, taskName)
	if timeEntry.Id != 0 {
		fmt.Printf("Time tracking for %s started!\n", timeEntry.Description)
	}
}

func (toggl Toggl) StopIssueTracking() {
	timeEntry := toggl.GetRunningTimeEntry()
	if timeEntry.Id != 0 {
		toggl.StopTimeEntry(timeEntry)
	}
}
