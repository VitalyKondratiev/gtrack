package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"../config"
	"github.com/Jeffail/gabs"
	"github.com/manifoldco/promptui"
)

type Jira struct {
	Credentials config.JiraConfig

	isLoggedIn  bool
	cookieName  string
	cookieValue string
}

type JiraTask struct {
	Id      int
	Key     string
	Summary string
}

func (jira Jira) IsLoggedIn() bool {
	return jira.isLoggedIn
}

func (jira Jira) SetCredentials() Jira {
	jira.isLoggedIn = false
	prompt := promptui.Prompt{
		Label: "Enter domain of your Jira instance",
	}
	result, err := prompt.Run()
	if err != nil {
		return jira
	}
	jira.Credentials.Domain = result

	l_prompt := promptui.Prompt{
		Label: "Enter your username",
	}
	result, err = l_prompt.Run()
	if err != nil {
		return jira
	}
	jira.Credentials.Username = result

	p_prompt := promptui.Prompt{
		Label: "Enter your password",
	}
	result, err = p_prompt.Run()
	if err != nil {
		return jira
	}
	jira.Credentials.Password = result

	jira = jira.authenticate()

	if jira.isLoggedIn {
		fmt.Printf("You sucessfully logged in %s as %s\n", jira.Credentials.Domain, jira.Credentials.Username)
	}

	return jira
}

func (jira Jira) authenticate() Jira {
	postBody, _ := json.Marshal(map[string]string{
		"username": jira.Credentials.Username,
		"password": jira.Credentials.Password,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("https://"+jira.Credentials.Domain+"/rest/auth/1/session", "application/json", responseBody)
	if err != nil {
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
	}
	return jira
}

func (jira Jira) GetAssignedTasks() []JiraTask {
	jira = jira.authenticate()
	var tasks []JiraTask
	client := http.Client{}
	req, err := http.NewRequest("GET", "https://"+jira.Credentials.Domain+"/rest/api/latest/search?jql=assignee="+jira.Credentials.Username+"&fields=summary", nil)
	if err != nil {
		return nil
	}
	authCookie := &http.Cookie{Name: jira.cookieName, Value: jira.cookieValue, HttpOnly: true}
	req.AddCookie(authCookie)
	resp, err := client.Do(req)
	if err != nil {
		return tasks
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		data, _ := ioutil.ReadAll(resp.Body)
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		//total := jsonParsed.Path("total").Data().(float64)
		//fmt.Print(total)
		for _, child := range jsonParsed.S("issues").Children() {
			issue := child.Data().(map[string]interface{})
			id, _ := strconv.Atoi(issue["id"].(string))
			var name string
			for _, value := range issue["fields"].(map[string]interface{}) {
				name = value.(string)
			}
			task := JiraTask{Id: id, Key: issue["key"].(string), Summary: name}
			tasks = append(tasks, task)
		}
	}
	return tasks
}
