package github

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Jeffail/gabs"
)

const releasesUri = "https://api.github.com/repos/vitalykondratiev/gtrack/releases/latest"

func (github Github) GetLastRelease() GithubRelease {
	client := http.Client{}
	req, err := http.NewRequest("GET", releasesUri, nil)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	var release GithubRelease
	if resp.StatusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			panic(err)
		}
		platformFiles := GithubFiles{}
		for _, child := range jsonParsed.Path("assets").Data().([]interface{}) {
			_child := child.(map[string]interface{})
			contentType := _child["content_type"].(string)
			downloadableFile := _child["browser_download_url"].(string)
			if contentType == "application/octet-stream" {
				platformFiles.LinuxBunary = downloadableFile
			} else if contentType == "application/x-ms-dos-executable" {
				platformFiles.WindowsBinary = downloadableFile
			} else if contentType == "application/gzip" {
				platformFiles.MacTGZ = downloadableFile
			}
		}
		publishedAt, _ := time.Parse("2006-01-02T15:04:05Z", jsonParsed.Path("published_at").Data().(string))
		release.DownloadableFiles = platformFiles
		release.Version = jsonParsed.Path("tag_name").Data().(string)
		release.PublishedAt = publishedAt.UTC()
		release.ReleasePage = jsonParsed.Path("html_url").Data().(string)
	}
	return release
}
