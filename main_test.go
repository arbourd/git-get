package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	cases := map[string]struct {
		args            []string
		wantStdout      string
		wantRunErr      bool
		wantErrContains string
		setup           func(t *testing.T)
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
			wantStdout: "testversion\n",
		},
		"-v": {
			args:       []string{"-v"},
			wantStdout: "testversion\n",
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
			err := run(c.args, &stdout)

			if c.wantRunErr && err == nil {
				t.Fatal("expected run() to return an error, got nil")
			}
			if !c.wantRunErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), c.wantErrContains)) {
				t.Fatalf("unexpected error:\n\t(GOT): %v\n\t(WNT): contains %q", err, c.wantErrContains)
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
