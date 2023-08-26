package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
)

func TestGetPath(t *testing.T) {
	home, _ := homedir.Dir()
	dir, _ := os.MkdirTemp("", "git-get")
	defer os.RemoveAll(dir)

	cases := map[string]struct {
		pathenv string
		want    string
	}{
		"default": {
			want: filepath.Join(home, "src"),
		},
		"custom": {
			pathenv: filepath.Join(dir, "src"),
			want:    filepath.Join(dir, "src"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// Empty GETPATH if set for tests
			os.Setenv("GETPATH", "")
			if c.pathenv != "" {
				os.Setenv("GETPATH", c.pathenv)
			}

			path, err := getPath()
			if err != nil {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if path != c.want {
				t.Fatalf("unexpected GETPATH:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.want)
			}
		})
	}
}

func TestClean(t *testing.T) {
	cases := map[string]struct {
		remote string
		want   string
	}{
		"git protocol": {
			remote: "git://github.com/arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"https protocol": {
			remote: "https://github.com/arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		".git removal": {
			remote: "github.com/arbourd/git-get.git",
			want:   "github.com/arbourd/git-get",
		},
		"filepath": {
			remote: "github.com///arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"gitlab subgroups": {
			remote: "gitlab.com/gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin",
			want:   "gitlab.com/gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			remote := clean(c.remote)
			if remote != c.want {
				t.Fatalf("unexpected cleaned string:\n\t(GOT): %#v\n\t(WNT): %#v", remote, c.want)
			}
		})
	}
}

func TestDownload(t *testing.T) {
	dir, _ := os.MkdirTemp("", "git-get")
	defer os.RemoveAll(dir)

	cases := map[string]struct {
		remote string
		want   string
		err    bool
	}{
		"github clone": {
			remote: "github.com/arbourd/git-get",
			want:   filepath.Join(dir, "github.com/arbourd/git-get"),
		},
		"gitlab clone": {
			remote: "gitlab.com/gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin",
			want:   filepath.Join(dir, "gitlab.com/gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin"),
		},
		"invalid remote": {
			remote: "https://github.com////arbourd/git-get",
			err:    true,
		},
		"not found remote": {
			remote: "github.com/arbourd/definitely-doesnt-exist",
			err:    true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path, err := download(dir, c.remote)

			if err != nil && !c.err {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.err {
				t.Fatalf("missing error:\n\t(GOT): nil\n\t")
			} else if path != c.want {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.want)
			}
		})
	}
}
