package github

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/inconshreveable/go-update"
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
				platformFiles.linux = downloadableFile
			} else if contentType == "application/x-ms-dos-executable" {
				platformFiles.windows = downloadableFile
			} else if contentType == "application/gzip" {
				platformFiles.darwin = downloadableFile
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

func (github Github) DownloadRelease(files GithubFiles) (bool, error) {
	result := false
	executable, _ := os.Executable()
	executablePath := filepath.Clean(executable)
	out, err := os.Create(executablePath + "_latest")
	if err != nil {
		return result, err
	}
	defer out.Close()
	r := reflect.ValueOf(files)
	url := reflect.Indirect(r).FieldByName(runtime.GOOS)
	resp, err := http.Get(url.String())
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return result, err
	}
	return true, err
}

func (github Github) Update() bool {
	executable, _ := os.Executable()
	executablePath := filepath.Clean(executable)
	_bytes, err := ioutil.ReadFile(executablePath + "_latest")
	err = update.Apply(bytes.NewReader(_bytes), update.Options{})
	if err != nil {
		panic(err)
	}
	return true
}
