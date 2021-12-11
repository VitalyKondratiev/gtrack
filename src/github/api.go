package github

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"syscall"
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

func (github Github) DownloadRelease(files GithubFiles) bool {
	executable, _ := os.Executable()
	executablePath := filepath.Clean(executable)
	out, err := os.Create(executablePath + "_latest")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	r := reflect.ValueOf(files)
	url := reflect.Indirect(r).FieldByName(runtime.GOOS)
	resp, err := http.Get(url.String())
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
	return true
}

func (github Github) ReplaceCurrent() bool {
	executable, _ := os.Executable()
	executablePath := filepath.Clean(executable)
	bytes, err := ioutil.ReadFile(executablePath + "_latest")
	if err != nil {
		panic(err)
	}
	syscall.Unlink(executablePath)
	fd_current, err := syscall.Open(executablePath, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_APPEND, 0)
	if err != nil {
		panic(err)
	}
	_, err = syscall.Write(fd_current, bytes)
	if err != nil {
		panic(err)
	}
	syscall.Unlink(executablePath + "_latest")
	return true
}
