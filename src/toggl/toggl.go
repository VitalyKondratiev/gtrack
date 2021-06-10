package toggl

import (
	"fmt"
	"time"

	"../helpers"
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

func (togglTimeEntry TogglTimeEntry) IsCurrent() bool {
	return togglTimeEntry.duration < 0
}

func (togglTimeEntry TogglTimeEntry) GetDuration() int {
	if togglTimeEntry.duration >= 0 {
		return togglTimeEntry.duration
	} else {
		startTime := time.Unix(int64(togglTimeEntry.duration)*-1, 0)
		diff := time.Since(startTime)
		return int(diff.Seconds())
	}
}

func (toggl Toggl) IsLoggedIn() bool {
	return toggl.isLoggedIn
}

func (toggl Toggl) SetConfig() Toggl {
	result, err := helpers.GetString("Enter API key of your Toggl.com account")
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
		index, err := helpers.GetVariant(
			"Select workspace",
			workspaces,
			"{{ .Name }}",
		)
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
		return
	}
	fmt.Println("You shouldn't have seen this text!")
}

func (toggl Toggl) StopIssueTracking() {
	timeEntry := toggl.GetRunningTimeEntry()
	if timeEntry.Id != 0 {
		if toggl.StopTimeEntry(timeEntry) {
			fmt.Printf("Time tracking for %s stopped!\n", timeEntry.Description)
			return
		}
	}
	fmt.Println("There is nothing to stop (or you found an error, but not likely)!")
}

func (toggl Toggl) CommitIssues(timeEntries []TogglTimeEntry, commit bool) (bool, []string, map[int]int, map[int]time.Time) {
	state := true
	timeEntry := toggl.GetRunningTimeEntry()
	if timeEntry.Id != 0 {
		if toggl.StopTimeEntry(timeEntry) {
			fmt.Printf("Time tracking for %s stopped!\n", timeEntry.Description)
		}
	}
	var issueKeys []string
	timeEntriesCounter := 0
	durations := make(map[int]int)
	startTimes := make(map[int]time.Time)
	for _, _timeEntry := range timeEntries {
		if _timeEntry.IsUncommitedEntry() || !commit {
			tags := []string{}
			if !commit {
				tags = append(tags, tagUncommitedName)
			}
			untagState := toggl.UpdateTimeEntryTags(_timeEntry, tags)
			if !untagState {
				state = false
			} else {
				issueKeys = append(issueKeys, _timeEntry.Description)
				durations[timeEntriesCounter] = _timeEntry.GetDuration()
				_time, _ := time.Parse(time.RFC3339, _timeEntry.Start)
				startTimes[timeEntriesCounter] = _time
				timeEntriesCounter++
			}
		}
	}
	return state, issueKeys, durations, startTimes
}
