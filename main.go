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
	repo := args[0]

	resp, err := download(path, repo)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(resp)
}

// getPath gets the GETPATH path and creates the directory if needed.
func getPath() (string, error) {
	path := os.Getenv("GETPATH")

	if len(path) == 0 {
		path = "~/src"
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

// download clones the repo to the GETPATH.
func download(path string, repo string) (string, error) {
	cmd := vcs.ByCmd("git")
	if err := cmd.Ping("https", repo); err != nil {
		return "", fmt.Errorf("%s is not a valid git repository", repo)
	}

	root := filepath.Join(path, repo)
	gitp := filepath.Join(root, ".git")

	if _, err := os.Stat(gitp); os.IsNotExist(err) {
		// Check if root folder exists, even though the .git directory does not.
		if _, err := os.Stat(root); !os.IsNotExist(err) {
			return "", fmt.Errorf("%s exists but %s does not", root, gitp)
		}

		parent, _ := filepath.Split(root)
		err := os.MkdirAll(parent, os.ModePerm)
		if err != nil {
			return "", err
		}

		if err = cmd.Create(root, "https://"+repo); err != nil {
			return "", err
		}
	} else {
		if err = cmd.Download(root); err != nil {
			return "", err
		}
	}

	return root, nil
}
