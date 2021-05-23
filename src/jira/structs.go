package jira

import (
	"../config"
)

type Jira struct {
	Config config.JiraConfig

	isLoggedIn  bool
	cookieName  string
	cookieValue string
}

type JiraIssue struct {
	Id         int
	ProjectKey string
	Key        string
	Summary    string
}
