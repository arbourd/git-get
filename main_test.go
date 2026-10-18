package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

func TestRun(t *testing.T) {
	cases := map[string]struct {
		args               []string
		wantStdout         string
		wantRunErr         bool
		wantErrContains    string
		wantErrNotContains string
		setup              func(t *testing.T)
	}{
		"no args": {
			args:            []string{},
			wantRunErr:      true,
			wantErrContains: "no repository specified",
		},
		"--help": {
			args:       []string{"--help"},
			wantStdout: "Usage: git-get",
		},
		"-h": {
			args:       []string{"-h"},
			wantStdout: "Usage: git-get",
		},
		"--version": {
			args:       []string{"--version"},
			wantStdout: "git-get testversion\n",
		},
		"-v": {
			args:       []string{"-v"},
			wantStdout: "git-get testversion\n",
		},
		"--verbose without repository": {
			args:            []string{"--verbose"},
			wantRunErr:      true,
			wantErrContains: "no repository specified",
		},
		"-v with repository is not version": {
			args:       []string{"-v", "github.com/arbourd/git-get"},
			wantStdout: filepath.FromSlash("github.com/arbourd/git-get\n"),
			setup:      setupGetpath,
		},
		"--verbose after repository": {
			args:       []string{"github.com/arbourd/git-get", "--verbose"},
			wantStdout: filepath.FromSlash("github.com/arbourd/git-get\n"),
			setup:      setupGetpath,
		},
		"--help after -v": {
			args:       []string{"-v", "--help"},
			wantStdout: "Usage: git-get",
		},
		"-h after repository": {
			args:       []string{"github.com/arbourd/git-get", "-h"},
			wantStdout: "Usage: git-get",
		},
		"--version after --verbose": {
			args:       []string{"--verbose", "--version"},
			wantStdout: "testversion\n",
		},
		"unknown option first": {
			args:            []string{"-vv", "github.com/arbourd/git-get"},
			wantRunErr:      true,
			wantErrContains: `unknown option "-vv"`,
		},
		"unknown option last": {
			args:            []string{"github.com/arbourd/git-get", "--verbos"},
			wantRunErr:      true,
			wantErrContains: `unknown option "--verbos"`,
		},
		"duplicate -v": {
			args:            []string{"-v", "-v"},
			wantRunErr:      true,
			wantErrContains: `duplicate option "-v"`,
		},
		"duplicate -v and --verbose": {
			args:            []string{"-v", "--verbose", "github.com/arbourd/git-get"},
			wantRunErr:      true,
			wantErrContains: `duplicate option "--verbose"`,
		},
		"too many arguments": {
			args:            []string{"github.com/arbourd/git-get", "github.com/arbourd/other"},
			wantRunErr:      true,
			wantErrContains: "too many arguments",
		},
		"too many arguments redacts credentials": {
			args:               []string{"github.com/arbourd/git-get", "https://user:s3cret@host/o/r"},
			wantRunErr:         true,
			wantErrContains:    "too many arguments",
			wantErrNotContains: "s3cret",
		},
		"unknown option redacts credentials": {
			args:               []string{"-https://user:s3cret@host/o/r"},
			wantRunErr:         true,
			wantErrContains:    "unknown option",
			wantErrNotContains: "s3cret",
		},
		"unparseable url redacts credentials": {
			args:               []string{"https://user:s3cret@host:badport/o/r"},
			wantRunErr:         true,
			wantErrContains:    "unable to parse repository url",
			wantErrNotContains: "s3cret",
		},
		"-- ends options": {
			args:       []string{"--", "github.com/arbourd/git-get"},
			wantStdout: filepath.FromSlash("github.com/arbourd/git-get\n"),
			setup:      setupGetpath,
		},
		"--complete empty prefix": {
			args:       []string{"--complete"},
			wantStdout: "github.com/\n",
			setup:      setupGetpath,
		},
		"--complete with prefix": {
			args:       []string{"--complete", "github.com/arbourd/git"},
			wantStdout: "github.com/arbourd/git-get\n",
			setup:      setupGetpath,
		},
		"--complete no match": {
			args:  []string{"--complete", "notexist"},
			setup: setupGetpath,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			Version = "testversion"

			if err := gitConfigGlobalFixture(t); err != nil {
				t.Fatalf("setup: %v", err)
			}
			if c.setup != nil {
				c.setup(t)
			}

			var stdout bytes.Buffer
			err := run(c.args, &stdout, io.Discard)

			if c.wantRunErr && err == nil {
				t.Fatal("expected run() to return an error, got nil")
			}
			if !c.wantRunErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), c.wantErrContains)) {
				t.Fatalf("unexpected error:\n\t(GOT): %v\n\t(WNT): contains %q", err, c.wantErrContains)
			}

			if c.wantErrNotContains != "" && err != nil && strings.Contains(err.Error(), c.wantErrNotContains) {
				t.Fatalf("error leaked %q: %v", c.wantErrNotContains, err)
			}

			if c.wantStdout != "" && !strings.Contains(stdout.String(), c.wantStdout) {
				t.Fatalf("unexpected stdout:\n\t(GOT): %q\n\t(WNT): contains %q", stdout.String(), c.wantStdout)
			}
			if c.wantStdout == "" && stdout.Len() > 0 {
				t.Fatalf("expected no stdout, got: %q", stdout.String())
			}
		})
	}
}

// TestRunVerbose clones from a local bare repository, mapped from a fake host
// with url.insteadOf, to check that git's commands and output reach stderr.
func TestRunVerbose(t *testing.T) {
	cases := map[string]struct {
		args          []string
		preclone      bool
		wantStderr    []string
		notWantStderr []string
	}{
		"-v": {
			args:       []string{"-v", "example.test/o/r"},
			wantStderr: []string{"+ git ls-remote https://example.test/o/r HEAD\n", "+ git clone https://example.test/o/r ", "Cloning into"},
		},
		"--verbose after repository": {
			args:       []string{"example.test/o/r", "--verbose"},
			wantStderr: []string{"+ git ls-remote https://example.test/o/r HEAD\n", "+ git clone https://example.test/o/r "},
		},
		"quiet by default": {
			args: []string{"example.test/o/r"},
		},
		"already cloned": {
			args:          []string{"-v", "example.test/o/r"},
			preclone:      true,
			wantStderr:    []string{"already cloned: "},
			notWantStderr: []string{"+ git"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if err := gitConfigGlobalFixture(t); err != nil {
				t.Fatalf("setup: %v", err)
			}

			remotes := t.TempDir()
			if _, err := git.Raw("init", func(g *types.Cmd) {
				g.AddOptions("--bare")
				g.AddOptions(filepath.Join(remotes, "o", "r"))
			}); err != nil {
				t.Fatalf("setup: git init: %v", err)
			}
			// file:///C:/... on Windows, file:///tmp/... elsewhere.
			remoteURL := "file:///" + strings.TrimPrefix(filepath.ToSlash(remotes), "/") + "/"
			config := fmt.Sprintf("[url %q]\n\tinsteadOf = https://example.test/\n", remoteURL)
			if err := os.WriteFile(os.Getenv("GIT_CONFIG_GLOBAL"), []byte(config), 0o600); err != nil {
				t.Fatalf("setup: %v", err)
			}

			getpath := t.TempDir()
			t.Setenv("GETPATH", getpath)

			if c.preclone {
				if err := run([]string{"example.test/o/r"}, io.Discard, io.Discard); err != nil {
					t.Fatalf("setup: initial clone: %v", err)
				}
			}

			var stdout, stderr bytes.Buffer
			if err := run(c.args, &stdout, &stderr); err != nil {
				t.Fatalf("unexpected error: %v\nstderr: %s", err, stderr.String())
			}

			want := filepath.Join(getpath, "example.test", "o", "r") + "\n"
			if stdout.String() != want {
				t.Fatalf("unexpected stdout:\n\t(GOT): %q\n\t(WNT): %q", stdout.String(), want)
			}
			if len(c.wantStderr) == 0 && stderr.Len() > 0 {
				t.Fatalf("expected no stderr, got: %q", stderr.String())
			}
			assertContains(t, stderr.String(), c.wantStderr, c.notWantStderr)
		})
	}
}

func gitConfigGlobalFixture(t *testing.T) error {
	t.Helper()
	gitconfig := filepath.Join(t.TempDir(), ".gitconfig")
	f, err := os.Create(gitconfig)
	if err != nil {
		return fmt.Errorf("unable to create .gitconfig: %w", err)
	}
	defer f.Close()

	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
	return nil
}

func setupGetpath(t *testing.T) {
	t.Helper()
	getpath := t.TempDir()
	seedRepos(t, getpath, []string{"github.com/arbourd/git-get"})
	t.Setenv("GETPATH", getpath)
}

func seedRepos(t *testing.T, getpath string, repos []string) {
	t.Helper()
	for _, r := range repos {
		dir := filepath.Join(getpath, filepath.FromSlash(r))
		if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
}
