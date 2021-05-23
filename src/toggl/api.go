package toggl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Jeffail/gabs"
)

const apiPath = "https://api.track.toggl.com/api/v8/"

func (toggl Toggl) apiGetData(apiMethod string) ([]byte, int) {
	client := http.Client{}
	req, err := http.NewRequest("GET", apiPath+apiMethod, nil)
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(toggl.Config.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (toggl Toggl) apiPostData(apiMethod string, payload []byte) ([]byte, int) {
	requestBody := bytes.NewBuffer(payload)
	client := http.Client{}
	req, err := http.NewRequest("POST", apiPath+apiMethod, requestBody)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(toggl.Config.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (toggl Toggl) apiPutData(apiMethod string, payload []byte) ([]byte, int) {
	requestBody := bytes.NewBuffer(payload)
	client := http.Client{}
	req, err := http.NewRequest("PUT", apiPath+apiMethod, requestBody)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(toggl.Config.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (toggl Toggl) GetWorkspaces() []TogglWorkspace {
	data, statusCode := toggl.apiGetData("workspaces")
	var workspaces []TogglWorkspace
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		for _, child := range jsonParsed.Children() {
			_workspace := child.Data().(map[string]interface{})
			if _workspace["admin"].(bool) {
				workspace := TogglWorkspace{
					Id:   int(_workspace["id"].(float64)),
					Name: _workspace["name"].(string),
				}
				workspaces = append(workspaces, workspace)
			}
		}
	}
	return workspaces
}

func (toggl Toggl) GetProjects() []TogglProject {
	data, statusCode := toggl.apiGetData("workspaces/" + strconv.Itoa(toggl.Config.WorkspaceId) + "/projects")
	var projects []TogglProject
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		for _, child := range jsonParsed.Children() {
			_project := child.Data().(map[string]interface{})
			project := TogglProject{
				Id:   int(_project["id"].(float64)),
				Name: _project["name"].(string),
			}
			projects = append(projects, project)
		}
	}
	return projects
}

func (toggl Toggl) GetWorkspaceTags() []TogglWorkspaceTag {
	data, statusCode := toggl.apiGetData("workspaces/" + strconv.Itoa(toggl.Config.WorkspaceId) + "/tags")
	var workspaceTags []TogglWorkspaceTag
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		for _, child := range jsonParsed.Children() {
			_workspaceTag := child.Data().(map[string]interface{})
			workspaceTag := TogglWorkspaceTag{
				Id:   int(_workspaceTag["id"].(float64)),
				Name: _workspaceTag["name"].(string),
			}
			workspaceTags = append(workspaceTags, workspaceTag)
		}
	}
	return workspaceTags
}

func (toggl Toggl) GetTimeEntries() []TogglTimeEntry {
	data, statusCode := toggl.apiGetData("time_entries")
	var timeEntries []TogglTimeEntry
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		for _, child := range jsonParsed.Children() {
			_timeEntry := child.Data().(map[string]interface{})
			var tags []string
			var description string
			var stop string
			for _, tag := range _timeEntry["tags"].([]interface{}) {
				tags = append(tags, tag.(string))
			}
			if _timeEntry["description"] != nil {
				description = _timeEntry["description"].(string)
			}
			if _timeEntry["stop"] != nil {
				stop = _timeEntry["stop"].(string)
			}
			timeEntry := TogglTimeEntry{
				Id:          int(_timeEntry["id"].(float64)),
				Description: description,
				Tags:        tags,
				Duration:    int(_timeEntry["duration"].(float64)),
				Start:       _timeEntry["start"].(string),
				Stop:        stop,
			}
			timeEntries = append(timeEntries, timeEntry)
		}
	}
	return timeEntries
}

func (toggl Toggl) GetRunningTimeEntry() TogglTimeEntry {
	data, statusCode := toggl.apiGetData("time_entries/current")
	var timeEntry TogglTimeEntry
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		for _, child := range jsonParsed.Children() {
			if child.String() == "null" {
				return timeEntry
			}
			_timeEntry := child.Data().(map[string]interface{})
			var tags []string
			var description string
			for _, tag := range _timeEntry["tags"].([]interface{}) {
				tags = append(tags, tag.(string))
			}
			if _timeEntry["description"] != nil {
				description = _timeEntry["description"].(string)
			}
			timeEntry = TogglTimeEntry{
				Id:          int(_timeEntry["id"].(float64)),
				Description: description,
				Tags:        tags,
				Duration:    int(_timeEntry["duration"].(float64)),
				Start:       _timeEntry["start"].(string),
			}
		}
	}
	return timeEntry
}

func (toggl Toggl) GetUser() (*string, int) {
	data, statusCode := toggl.apiGetData("me")
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		fullname := jsonParsed.Path("data.fullname").Data().(string)
		return &fullname, statusCode
	}
	return nil, statusCode
}

func (toggl Toggl) CreateTag(tagName string) TogglWorkspaceTag {
	var workspaceTag TogglWorkspaceTag
	jsonObj := gabs.New()
	jsonObj.SetP(tagName, "tag.name")
	jsonObj.SetP(toggl.Config.WorkspaceId, "tag.wid")
	data, statusCode := toggl.apiPostData("tags", jsonObj.Bytes())
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		workspaceTag = TogglWorkspaceTag{
			Id:   int(jsonParsed.S("data").Data().(map[string]interface{})["id"].(float64)),
			Name: jsonParsed.S("data").Data().(map[string]interface{})["name"].(string),
		}
	}
	return workspaceTag
}

func (toggl Toggl) CreateProject(projectName string) TogglProject {
	var project TogglProject
	jsonObj := gabs.New()
	jsonObj.SetP(projectName, "project.name")
	jsonObj.SetP(toggl.Config.WorkspaceId, "project.wid")
	data, statusCode := toggl.apiPostData("projects", jsonObj.Bytes())
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		project = TogglProject{
			Id:   int(jsonParsed.S("data").Data().(map[string]interface{})["id"].(float64)),
			Name: jsonParsed.S("data").Data().(map[string]interface{})["name"].(string),
		}
	}
	return project
}

func populateTimeEntry(workspaceId int, projectId int, description string) *gabs.Container {
	t := time.Now()
	tZone, _ := t.Zone()
	time := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d%02s:00",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), tZone)
	jsonObj := gabs.New()
	jsonObj.SetP(projectId, "time_entry.pid")
	jsonObj.SetP(workspaceId, "time_entry.wid")
	jsonObj.SetP(description, "time_entry.description")
	jsonObj.SetP(time, "time_entry.start")
	jsonObj.SetP(0, "time_entry.duration")
	jsonObj.SetP("gtrack", "time_entry.created_with")
	jsonObj.ArrayP("time_entry.tags")
	jsonObj.ArrayAppendP(tagUncommitedName, "time_entry.tags")
	return jsonObj
}

func (toggl Toggl) CreateTimeEntry(projectId int, description string) TogglTimeEntry {
	var timeEntry TogglTimeEntry
	jsonObj := populateTimeEntry(toggl.Config.WorkspaceId, projectId, description)
	data, statusCode := toggl.apiPostData("time_entries", jsonObj.Bytes())
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		timeEntry = TogglTimeEntry{
			Id:          int(jsonParsed.S("data").Data().(map[string]interface{})["id"].(float64)),
			Description: description,
			Tags:        []string{tagUncommitedName},
			Duration:    0,
			Start:       jsonParsed.S("data").Data().(map[string]interface{})["start"].(string),
		}
	}
	return timeEntry
}

func (toggl Toggl) StartTimeEntry(projectId int, description string) TogglTimeEntry {
	var timeEntry TogglTimeEntry
	jsonObj := populateTimeEntry(toggl.Config.WorkspaceId, projectId, description)
	data, statusCode := toggl.apiPostData("time_entries/start", jsonObj.Bytes())
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		timeEntry = TogglTimeEntry{
			Id:          int(jsonParsed.S("data").Data().(map[string]interface{})["id"].(float64)),
			Description: description,
			Tags:        []string{tagUncommitedName},
			Duration:    0,
			Start:       jsonParsed.S("data").Data().(map[string]interface{})["start"].(string),
		}
	}
	return timeEntry
}

func (toggl Toggl) StopTimeEntry(timeEntry TogglTimeEntry) bool {
	jsonObj := gabs.New()
	_, statusCode := toggl.apiPutData("time_entries/"+strconv.Itoa(timeEntry.Id)+"/stop", jsonObj.Bytes())
	return statusCode == 200
}
