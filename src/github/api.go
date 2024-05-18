package github

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/VitalyKondratiev/gtrack/src/helpers"
	"github.com/inconshreveable/go-update"
)

const releasesUri = "https://api.github.com/repos/vitalykondratiev/gtrack/releases/latest"

func (github Github) HasUpdate() (bool, GithubRelease) {
	executable, _ := os.Executable()
	executablePath := filepath.Clean(executable)
	executableStat, _ := os.Stat(executablePath)
	executableModifiedAt := executableStat.ModTime().UTC()
	githubLastRelease := github.GetLastRelease()
	difference := executableModifiedAt.Sub(githubLastRelease.PublishedAt).Seconds()
	return difference < 0, githubLastRelease
}

func (github Github) GetLastRelease() GithubRelease {
	client := http.Client{}
	req, err := http.NewRequest("GET", releasesUri, nil)
	if err != nil {
		helpers.LogFatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		helpers.LogFatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var release GithubRelease
	if resp.StatusCode == 200 {
		jsonParsed, err := gabs.ParseJSON(data)
		if err != nil {
			helpers.LogFatal(
				fmt.Errorf("message: unable to parse json (%v)\nurl: %v\n\nresponse:\n%v", err, resp.Request.URL, string(data)),
			)
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
	updateFile := os.TempDir() + "/gtrack_update_file"
	out, err := os.Create(updateFile)
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
	updateFile := os.TempDir() + "/gtrack_update_file"
	var _bytes []byte
	var err error
	if runtime.GOOS != "darwin" {
		_bytes, err = os.ReadFile(updateFile)
		if err != nil {
			helpers.LogFatal(err)
		}
	} else {
		fmt.Printf("Archive unpacking...\n")
		tarFile, err := os.Open(updateFile)
		if err != nil {
			helpers.LogFatal(err)
		}
		uncompressedStream, err := gzip.NewReader(tarFile)
		if err != nil {
			helpers.LogFatal(err)
		}
		tarReader := tar.NewReader(uncompressedStream)
		tarHeader, err := tarReader.Next()
		if err != io.EOF && tarHeader.Typeflag == tar.TypeReg {
			_bytes, err = io.ReadAll(tarReader)
			if err != nil {
				helpers.LogFatal(err)
			}
		}
	}
	err = update.Apply(bytes.NewReader(_bytes), update.Options{})
	if err != nil {
		helpers.LogFatal(err)
	} else {
		os.Remove(updateFile)
	}
	return true
}
