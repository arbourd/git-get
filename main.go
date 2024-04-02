package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arbourd/git-get/get"
)

func main() {
	dir, err := run()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(dir)
}

func run() (string, error) {
	path, err := get.Path()
	if err != nil {
		return "", err
	}

	args := os.Args[1:]
	if len(args) == 0 {
		return "", fmt.Errorf("must provide a git repository url")
	}
	remote := args[0]

	url, err := get.ParseURL(remote)
	if err != nil {
		return "", fmt.Errorf("unable to parse repository url: \"%s\"", remote)
	}

	dir := filepath.Join(path, get.Directory(url))
	return get.Clone(url, dir)
}
