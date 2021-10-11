package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"../helpers"
	"github.com/Jeffail/gabs"
)

func (jira Jira) getApiPath() string {
	return helpers.GetFormattedDomain(jira.Config.Domain) + "/rest/api/latest/"
}

func (jira Jira) authenticate() Jira {
	jira.isLoggedIn = false
	postBody, _ := json.Marshal(map[string]string{
		"username": jira.Config.Username,
		"password": jira.Config.Password,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(helpers.GetFormattedDomain(jira.Config.Domain)+"/rest/auth/1/session", "application/json", responseBody)
	if err != nil {
		fmt.Println(err)
		return jira
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		sessionName := jsonParsed.Path("session.name").Data().(string)
		sessionValue := jsonParsed.Path("session.value").Data().(string)
		jira.cookieName = sessionName
		jira.cookieValue = sessionValue
		jira.isLoggedIn = true
		return jira
	} else {
		fmt.Println("HTTP status code is: " + strconv.Itoa(resp.StatusCode))
	}
	return jira
}

func (jira Jira) apiGetData(apiMethod string) ([]byte, int) {
	jira = jira.authenticate()
	client := http.Client{}
	req, err := http.NewRequest("GET", jira.getApiPath()+apiMethod, nil)
	if err != nil {
		panic(err)
	}
	authCookie := &http.Cookie{Name: jira.cookieName, Value: jira.cookieValue, HttpOnly: true}
	req.AddCookie(authCookie)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return data, resp.StatusCode
}
func (jira Jira) apiPostData(apiMethod string, payload []byte) ([]byte, int) {
	jira = jira.authenticate()
	requestBody := bytes.NewBuffer(payload)
	client := http.Client{}
	req, err := http.NewRequest("POST", jira.getApiPath()+apiMethod, requestBody)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	authCookie := &http.Cookie{Name: jira.cookieName, Value: jira.cookieValue, HttpOnly: true}
	req.AddCookie(authCookie)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	return data, resp.StatusCode
}

func (jira Jira) GetAssignedIssues() []JiraIssue {
	data, statusCode := jira.apiGetData("search?jql=assignee=" + jira.Config.Username + "&fields=summary,project")
	var tasks []JiraIssue
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		for _, child := range jsonParsed.S("issues").Children() {
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
	data, statusCode := jira.apiGetData("search?jql=" + strings.Join(statements, "%20or%20") + "&fields=summary,project")
	if statusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		for _, child := range jsonParsed.S("issues").Children() {
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

func (jira Jira) SetWorklogEntry(issueId int, duration int, startTime time.Time) bool {
	jsonObj := gabs.New()
	jsonObj.SetP("", "comment")
	jsonObj.SetP(startTime.Format("2006-01-02T15:04:05.000+0000"), "started")
	jsonObj.SetP(duration, "timeSpentSeconds")
	_, statusCode := jira.apiPostData("issue/"+strconv.Itoa(issueId)+"/worklog", jsonObj.Bytes())
	return statusCode == 201
}
