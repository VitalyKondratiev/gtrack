package toggl

import (
	"github.com/VitalyKondratiev/gtrack/src/config"
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
	Id          int      `json:"id"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Duration    int      `json:"duration"`
	Start       string   `json:"start"`
	Stop        string   `json:"stop"`
}
