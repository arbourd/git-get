package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arbourd/git-get/get"
)

// Version is set via -ldflags at build time.
var Version = "dev"

const usage = `Usage: git-get [options] <repository>

Clone a git repository to GETPATH (%s).

Arguments:
  repository  The git repository URL to clone

Options:
  -h, --help     Show this help message
  -v, --verbose  Show the git commands being run and their output
  --version      Show version (also -v when no repository is given)`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) > 0 && args[0] == "--complete" {
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

	// Help and version win wherever they appear. A lone -v prints the version;
	// alongside a repository it means --verbose.
	for _, a := range args {
		if a == "--" {
			break
		}
		switch {
		case a == "--help" || a == "-h":
			fmt.Fprintln(stdout, buildUsage())
			return nil
		case a == "--version" || (a == "-v" && len(args) == 1):
			fmt.Fprintf(stdout, "git-get %s\n", Version)
			return nil
		}
	}

	var remote string
	var verbose io.Writer
	positional := false
	for _, a := range args {
		switch {
		case positional || !strings.HasPrefix(a, "-"):
			if remote != "" {
				return fmt.Errorf("too many arguments: %q\n\n%s", get.RedactRemote(a), buildUsage())
			}
			remote = a
		case a == "--":
			positional = true
		case a == "-v" || a == "--verbose":
			if verbose != nil {
				return fmt.Errorf("duplicate option %q\n\n%s", a, buildUsage())
			}
			verbose = stderr
		default:
			return fmt.Errorf("unknown option %q\n\n%s", get.RedactRemote(a), buildUsage())
		}
	}
	if remote == "" {
		return fmt.Errorf("no repository specified\n\n%s", buildUsage())
	}

	return clone(remote, stdout, verbose)
}

func clone(remote string, stdout, verbose io.Writer) error {
	path, err := get.AbsolutePath()
	if err != nil {
		return fmt.Errorf("resolving GETPATH: %w", err)
	}

	url, err := get.ParseURL(remote)
	if err != nil {
		return fmt.Errorf("unable to parse repository url %q: %w", get.RedactRemote(remote), err)
	}

	relDir, err := get.Directory(url)
	if err != nil {
		return fmt.Errorf("unable to determine directory for url %q: %w", get.RedactRemote(remote), err)
	}

	dir := filepath.Join(path, relDir)
	var result string
	cloneFn := func() (err error) {
		result, err = get.Clone(url, dir, verbose)
		return err
	}
	if verbose != nil {
		// Verbose output goes to stderr and would collide with the spinner.
		err = cloneFn()
	} else {
		err = withSpinner("Cloning into "+get.ShortPath(dir)+"...", cloneFn)
	}
	if err != nil {
		return fmt.Errorf("cloning repository: %w", err)
	}

	fmt.Fprintln(stdout, result)
	return nil
}

var spinnerFrames = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}

// withSpinner runs fn behind a spinner on stderr. When stderr is not a terminal, fn runs bare so
// piped and CI output is unchanged.
func withSpinner(title string, fn func() error) error {
	if fi, err := os.Stderr.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		return fn()
	}

	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		t := time.NewTicker(80 * time.Millisecond)
		defer t.Stop()
		for i := 0; ; i++ {
			fmt.Fprintf(os.Stderr, "\r%s %s", spinnerFrames[i%len(spinnerFrames)], title)
			select {
			case <-done:
				fmt.Fprint(os.Stderr, "\r\x1b[K")
				return
			case <-t.C:
			}
		}
	}()
	err := fn()
	close(done)
	<-stopped
	return err
}

func buildUsage() string {
	path, err := get.AbsolutePath()
	if err != nil {
		return fmt.Sprintf(usage, "~/src")
	}

	return fmt.Sprintf(usage, get.ShortPath(path))
}
