package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arbourd/git-get/get"
)

func main() {
	path, err := get.Path()
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

	url, err := get.ParseURL(remote)
	if err != nil {
		fmt.Printf("Error: unable to parse repository url: \"%s\"\n", remote)
		os.Exit(1)
	}

	dir := filepath.Join(path, get.Directory(url))

	resp, err := get.Clone(url, dir)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(resp)
}
