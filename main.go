package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/tools/go/vcs"
)

func main() {
	path, err := getPath()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Error: must provide a git repository url")
		os.Exit(1)
	}
	remote := args[0]

	resp, err := download(path, remote)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(resp)
}

// DefaultPath is the path where repositories will be cloned to if GETPATH is unset.
const DefaultPath = "~/src"

// getPath gets the GETPATH path and creates the directory if needed.
func getPath() (string, error) {
	path := os.Getenv("GETPATH")
	if len(path) == 0 {
		path = DefaultPath
	}

	path, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("GETPATH entry is relative; must be absolute path: \"%s\"", path)
	}

	// Make GETPATH directory if it does not exist.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

// download clones the remote repository to the GETPATH.
func download(path string, remote string) (string, error) {
	cmd := vcs.ByCmd("git")
	repo, err := vcs.RepoRootForImportPath(remote, false)
	if err != nil || repo.VCS != cmd {
		return "", fmt.Errorf("%s is not a valid git repository", remote)
	}

	dir := filepath.Join(path, repo.Root)
	git := filepath.Join(dir, ".git")
	if _, err := os.Stat(git); os.IsNotExist(err) {
		// Check if root folder exists, even though the .git directory does not.
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			return "", fmt.Errorf("%s exists but %s does not", dir, git)
		}

		parent, _ := filepath.Split(dir)
		err := os.MkdirAll(parent, os.ModePerm)
		if err != nil {
			return "", err
		}

		if err = cmd.Create(dir, repo.Repo); err != nil {
			return "", err
		}
	} else {
		if err = cmd.Download(dir); err != nil {
			return "", err
		}
	}

	return dir, nil
}
