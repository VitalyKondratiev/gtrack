package github

import "time"

type Github struct {
}

type GithubFiles struct {
	MacTGZ        string
	WindowsBinary string
	LinuxBunary   string
}

type GithubRelease struct {
	DownloadableFiles GithubFiles
	Version           string
	PublishedAt       time.Time
	ReleasePage       string
}
