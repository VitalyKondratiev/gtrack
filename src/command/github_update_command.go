package command

import (
	"fmt"
	"os"

	"github.com/VitalyKondratiev/gtrack/src/github"
	"github.com/VitalyKondratiev/gtrack/src/helpers"
)

type githubUpdateCommand struct {
}

func GithubUpdateCommand() *githubUpdateCommand {
	return &githubUpdateCommand{}
}

func (cmd *githubUpdateCommand) Execute() {
	gh := github.Github{}
	hasUpdate, githubRelease := gh.HasUpdate()
	if hasUpdate {
		variant, err := helpers.GetVariant(
			fmt.Sprintf("Update available to %s (see in browser: %s)", githubRelease.Version, githubRelease.ReleasePage),
			[]string{"Update now", "Cancel"},
			"{{ . }} ",
		)
		if err != nil || variant == 1 {
			os.Exit(1)
		}
		fmt.Printf("Downloading release %s...\n", githubRelease.Version)
		isFileDownloaded, err := gh.DownloadRelease(githubRelease.DownloadableFiles)
		if err != nil || !isFileDownloaded {
			fmt.Printf("Download release error :(")
		}
		isUpdated := gh.Update()
		fmt.Printf("Replacing binary file...\n")
		if isUpdated {
			fmt.Printf("Succesfully updated to %s\n", githubRelease.Version)
		} else {
			helpers.LogFatal(err)
		}
	} else {
		fmt.Println("No available updates")
	}
}
