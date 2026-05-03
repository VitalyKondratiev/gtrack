package jira

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/VitalyKondratiev/gtrack/src/helpers"
)

func (jira Jira) getApiPath() string {
	return helpers.GetFormattedDomain(jira.Config.Domain) + "/rest/api/latest/"
}

func (jira Jira) applyAuth(req *http.Request) error {
	if jira.Config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+jira.Config.Token)
		return nil
	}

	return fmt.Errorf("You not authorized")
}

func (jira Jira) apiGetData(apiMethod string) ([]byte, int) {
	client := http.Client{}
	req, err := http.NewRequest("GET", jira.getApiPath()+apiMethod, nil)
	if err != nil {
		helpers.LogFatal(err)
	}
	if err = jira.applyAuth(req); err != nil {
		helpers.LogFatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		helpers.LogFatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (jira Jira) apiPostData(apiMethod string, payload []byte) ([]byte, int) {
	requestBody := bytes.NewBuffer(payload)
	client := http.Client{}
	req, err := http.NewRequest("POST", jira.getApiPath()+apiMethod, requestBody)
	if err != nil {
		helpers.LogFatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	if err = jira.applyAuth(req); err != nil {
		helpers.LogFatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		helpers.LogFatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (jira Jira) GetAssignedIssues() []JiraIssue {
	jql := "assignee=currentUser()"
	data, statusCode := jira.apiGetData("search?jql=" + url.QueryEscape(jql) + "&fields=summary,project")
	var tasks []JiraIssue
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		children, _ := jsonParsed.S("issues").Children()
		for _, child := range children {
			_issue := child.Data().(map[string]interface{})
			id, _ := strconv.Atoi(_issue["id"].(string))
			issue := JiraIssue{
				Id:         id,
				ProjectKey: child.Path("fields.project").Data().(map[string]interface{})["key"].(string),
				Key:        _issue["key"].(string),
				Summary:    child.S("fields").Data().(map[string]interface{})["summary"].(string),
			}
			tasks = append(tasks, issue)
		}
	}
	return tasks
}

func (jira Jira) GetIssuesByField(values []string, field string) []JiraIssue {
	var tasks []JiraIssue
	var statements []string
	for _, value := range values {
		statements = append(statements, fmt.Sprintf("%s=%s", field, value))
	}
	data, statusCode := jira.apiGetData("search?jql=" + url.QueryEscape(strings.Join(statements, " or ")) + "&fields=summary,project")
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\n\nresponse:\n%v", err, string(data)),
			)
		}
		children, _ := jsonParsed.S("issues").Children()
		for _, child := range children {
			_issue := child.Data().(map[string]interface{})
			id, _ := strconv.Atoi(_issue["id"].(string))
			issue := JiraIssue{
				Id:         id,
				ProjectKey: child.Path("fields.project").Data().(map[string]interface{})["key"].(string),
				Key:        _issue["key"].(string),
				Summary:    child.S("fields").Data().(map[string]interface{})["summary"].(string),
			}
			tasks = append(tasks, issue)
		}
	}
	return tasks
}

func (jira Jira) GetCurrentUser() (*string, int) {
	data, statusCode := jira.apiGetData("myself")
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err == nil {
			for _, field := range []string{"displayName", "name", "emailAddress"} {
				if value := jsonParsed.Path(field).Data(); value != nil {
					currentUser := value.(string)
					return &currentUser, statusCode
				}
			}
		}
	}

	currentUser := "token auth"
	return &currentUser, statusCode
}

func (jira Jira) SetWorklogEntry(issueId int, duration int, startTime time.Time) bool {
	jsonObj := gabs.New()
	jsonObj.SetP("", "comment")
	jsonObj.SetP(startTime.Format("2006-01-02T15:04:05.000+0000"), "started")
	jsonObj.SetP(duration, "timeSpentSeconds")
	_, statusCode := jira.apiPostData("issue/"+strconv.Itoa(issueId)+"/worklog", jsonObj.Bytes())
	return statusCode == 201
}
