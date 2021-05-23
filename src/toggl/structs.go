package toggl

import (
	"../config"
)

type Toggl struct {
	Config     config.TogglConfig
	isLoggedIn bool
}

type TogglWorkspace struct {
	Id   int
	Name string
}

type TogglProject struct {
	Id   int
	Name string
}

type TogglWorkspaceTag struct {
	Id   int
	Name string
}

type TogglTimeEntry struct {
	Id          int
	Description string
	Tags        []string
	duration    int
	Start       string
	Stop        string
}
