package github

import "time"

type Github struct {
}

type GithubFiles struct {
	darwin  string
	windows string
	linux   string
}

type GithubRelease struct {
	DownloadableFiles GithubFiles
	Version           string
	PublishedAt       time.Time
	ReleasePage       string
}
