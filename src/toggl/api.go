package toggl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/VitalyKondratiev/gtrack/src/helpers"
)

const apiPath = "https://api.track.toggl.com/api/v9/"

func (toggl Toggl) getIssueKey(description string) string {
	keyRegex := regexp.MustCompile(`(?m)\w*-\d*`)
	return keyRegex.FindString(description)
}

func (toggl Toggl) apiGetData(apiMethod string) ([]byte, int) {
	client := http.Client{}
	req, err := http.NewRequest("GET", apiPath+apiMethod, nil)
	if err != nil {
		helpers.LogFatal(err)
	}
	req.SetBasicAuth(toggl.Config.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		helpers.LogFatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (toggl Toggl) apiPatchData(apiMethod string) ([]byte, int) {
	client := http.Client{}
	req, err := http.NewRequest("PATCH", apiPath+apiMethod, nil)
	if err != nil {
		helpers.LogFatal(err)
	}
	req.SetBasicAuth(toggl.Config.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		helpers.LogFatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (toggl Toggl) apiPostData(apiMethod string, payload []byte) ([]byte, int) {
	requestBody := bytes.NewBuffer(payload)
	client := http.Client{}
	req, err := http.NewRequest("POST", apiPath+apiMethod, requestBody)
	if err != nil {
		helpers.LogFatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(toggl.Config.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (toggl Toggl) apiPutData(apiMethod string, payload []byte) ([]byte, int) {
	requestBody := bytes.NewBuffer(payload)
	client := http.Client{}
	req, err := http.NewRequest("PUT", apiPath+apiMethod, requestBody)
	if err != nil {
		helpers.LogFatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(toggl.Config.ApiKey, "api_token")
	resp, err := client.Do(req)
	if err != nil {
		helpers.LogFatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (toggl Toggl) GetWorkspaces() []TogglWorkspace {
	data, statusCode := toggl.apiGetData("workspaces")
	var workspaces []TogglWorkspace
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		children, _ := jsonParsed.Children()
		for _, child := range children {
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
	method := fmt.Sprintf("workspaces/%d/projects", toggl.Config.WorkspaceId)
	data, statusCode := toggl.apiGetData(method)
	var projects []TogglProject
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		children, _ := jsonParsed.Children()
		for _, child := range children {
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
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		children, _ := jsonParsed.Children()
		for _, child := range children {
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
	data, statusCode := toggl.apiGetData("me/time_entries")
	var timeEntries []TogglTimeEntry
	if statusCode == 200 {
		err := json.Unmarshal(data, &timeEntries)
		if err != nil {
			log.Fatalf("Error parsing JSON: %v\nResponse:\n%v", err, string(data))
		}
		for i, entry := range timeEntries {
			timeEntries[i].Description = toggl.getIssueKey(entry.Description)
		}
	}
	return timeEntries
}

func (toggl Toggl) GetRunningTimeEntry() TogglTimeEntry {
	data, statusCode := toggl.apiGetData("me/time_entries/current")
	var timeEntry TogglTimeEntry
	if statusCode == 200 {
		json.Unmarshal(data, &timeEntry)
	}
	return timeEntry
}

func (toggl Toggl) GetUser() (*string, int) {
	data, statusCode := toggl.apiGetData("me")
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		fullname := jsonParsed.Path("fullname").Data().(string)
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
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
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
	jsonObj.SetP(projectName, "name")
	method := fmt.Sprintf("workspaces/%d/projects", toggl.Config.WorkspaceId)
	data, statusCode := toggl.apiPostData(method, jsonObj.Bytes())
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		project = TogglProject{
			Id:   int(jsonParsed.S("data").Data().(map[string]interface{})["id"].(float64)),
			Name: jsonParsed.S("data").Data().(map[string]interface{})["name"].(string),
		}
	}
	return project
}

func populateTimeEntry(workspaceId int, projectId int, description string, jiraTag string) *gabs.Container {
	t := time.Now()
	tZone, _ := t.Zone()
	time := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d%02s:00",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), tZone)
	jsonObj := gabs.New()
	jsonObj.SetP(projectId, "project_id")
	jsonObj.SetP(workspaceId, "wid")
	jsonObj.SetP(description, "description")
	jsonObj.SetP(time, "start")
	jsonObj.SetP(-1, "duration")
	jsonObj.SetP("gtrack", "created_with")
	jsonObj.ArrayP("tags")
	jsonObj.ArrayAppendP(tagUncommitedName, "tags")
	jsonObj.ArrayAppendP(jiraTag, "tags")
	return jsonObj
}

func (toggl Toggl) StartTimeEntry(projectId int, description string, tagJiraDomain string) TogglTimeEntry {
	var timeEntry TogglTimeEntry
	tagJiraDomain = "gtrack:" + tagJiraDomain
	method := fmt.Sprintf("workspaces/%d/time_entries", toggl.Config.WorkspaceId)
	jsonObj := populateTimeEntry(toggl.Config.WorkspaceId, projectId, description, tagJiraDomain)
	data, statusCode := toggl.apiPostData(method, jsonObj.Bytes())
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		timeEntry = TogglTimeEntry{
			Id:          int(jsonParsed.Data().(map[string]interface{})["id"].(float64)),
			Description: description,
			Tags:        []string{tagUncommitedName, tagJiraDomain},
			Duration:    0,
			Start:       jsonParsed.Data().(map[string]interface{})["start"].(string),
		}
	}
	return timeEntry
}

func (toggl Toggl) StopTimeEntry(timeEntry TogglTimeEntry) bool {
	method := fmt.Sprintf("workspaces/%d/time_entries/%d/stop", toggl.Config.WorkspaceId, timeEntry.Id)
	_, statusCode := toggl.apiPatchData(method)
	return statusCode == 200
}

func (toggl Toggl) UpdateTimeEntryTags(timeEntry TogglTimeEntry, removeUncommited bool) bool {
	jsonObj := gabs.New()
	if removeUncommited {
		jsonObj.SetP("remove", "tag_action")
	} else {
		jsonObj.SetP("add", "tag_action")
	}
	jsonObj.ArrayP("tags")
	jsonObj.ArrayAppendP(tagUncommitedName, "tags")
	method := fmt.Sprintf("workspaces/%d/time_entries/%d", toggl.Config.WorkspaceId, timeEntry.Id)
	_, statusCode := toggl.apiPutData(method, jsonObj.Bytes())
	return statusCode == 200
}
