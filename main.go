package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/arbourd/git-get/get"
)

// Version is set via -ldflags at build time.
var Version = "dev"

const usage = `Usage: git-get <repository>

Clone a git repository to GETPATH (%s).

Arguments:
  repository  The git repository URL to clone

Options:
  -h, --help     Show this help message
  -v, --version  Show version`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("no repository specified\n\n%s", buildUsage())
	}

	switch args[0] {
	case "--help", "-h":
		fmt.Fprintln(stdout, buildUsage())
		return nil
	case "--version", "-v":
		fmt.Fprintln(stdout, Version)
		return nil
	case "--complete":
		// Internal protocol used by shell completion scripts (completions/); not a user-facing flag.
		prefix := ""
		if len(args) > 1 {
			prefix = args[1]
		}
		matches, err := get.Complete(prefix)
		if err != nil {
			return fmt.Errorf("completions: %w", err)
		}
		for _, m := range matches {
			fmt.Fprintln(stdout, m)
		}
		return nil
	}

	return clone(args[0], stdout)
}

func clone(remote string, stdout io.Writer) error {
	path, err := get.AbsolutePath()
	if err != nil {
		return fmt.Errorf("resolving GETPATH: %w", err)
	}

	url, err := get.ParseURL(remote)
	if err != nil {
		return fmt.Errorf("unable to parse repository url %q: %w", remote, err)
	}

	relDir, err := get.Directory(url)
	if err != nil {
		return fmt.Errorf("unable to determine directory for url %q: %w", remote, err)
	}

	dir := filepath.Join(path, relDir)
	result, err := get.Clone(url, dir)
	if err != nil {
		return fmt.Errorf("cloning repository: %w", err)
	}

	fmt.Fprintln(stdout, result)
	return nil
}

func buildUsage() string {
	path, err := get.AbsolutePath()
	if err != nil {
		return fmt.Sprintf(usage, "~/src")
	}

	return fmt.Sprintf(usage, get.ShortPath(path))
}
