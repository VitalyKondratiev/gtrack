package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

type getCurrentBranchCommand struct {
}

func GetCurrentBranchCommand() *getCurrentBranchCommand {
	return &getCurrentBranchCommand{}
}

func (cmd *getCurrentBranchCommand) Execute() string {
	dir, _ := os.Getwd()
	dir, isGitRoot := cmd.tryGetGitDirectory(dir)
	if !isGitRoot {
		fmt.Println("Can't find .git folder in this path directories")
		os.Exit(1)
	}
	repo, err := git.PlainOpen(dir)
	if err != nil {
		os.Exit(1)
	}
	h, err := repo.Head()
	if err != nil {
		os.Exit(1)
	}
	return strings.Replace(h.Name().String(), "refs/heads/", "", -1)
}

func (cmd *getCurrentBranchCommand) tryGetGitDirectory(dir string) (string, bool) {
	isGitRoot := false
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if f.Name() == ".git" {
			isGitRoot = true
			break
		}
	}
	if !isGitRoot && dir != filepath.Clean(filepath.Join(dir, "..")) {
		dir = filepath.Clean(filepath.Join(dir, ".."))
		dir, isGitRoot = cmd.tryGetGitDirectory(dir)
	}
	return dir, isGitRoot
}
