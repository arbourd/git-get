package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
)

func TestGetPath(t *testing.T) {
	home, _ := homedir.Dir()
	dir, _ := ioutil.TempDir("", "git-get")
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
		"git": {
			remote: "git://github.com/arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"https": {
			remote: "https://github.com/arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"filepath": {
			remote: "github.com///arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"https filepath": {
			remote: "https://github.com///arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			str := clean(c.remote)
			if str != c.want {
				t.Fatalf("unexpected cleaned string:\n\t(GOT): %#v\n\t(WNT): %#v", str, c.want)
			}
		})
	}
}
