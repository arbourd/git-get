package get

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
)

func TestPath(t *testing.T) {
	home, _ := homedir.Dir()
	dir, _ := os.MkdirTemp("", "git-get")
	defer os.RemoveAll(dir)

	cases := map[string]struct {
		pathenv string
		want    string
		err     string
	}{
		"default": {
			want: filepath.Join(home, "src"),
		},
		"custom": {
			pathenv: filepath.Join(dir, "src"),
			want:    filepath.Join(dir, "src"),
		},
		"invalid": {
			pathenv: "../test",
			err:     "GETPATH entry is relative; must be an absolute path",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// Empty GETPATH if set for tests
			os.Setenv("GITGET_GETPATH", "")
			if c.pathenv != "" {
				os.Setenv("GITGET_GETPATH", c.pathenv)
			}

			path, err := Path()
			if err != nil && c.err == "" {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && len(c.err) > 0 {
				t.Fatalf("expected error:\n\t(GOT): nil\n\t(WNT): %s", c.err)
			} else if path != c.want {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.want)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	dir, _ := os.MkdirTemp("", "git-get")
	defer os.RemoveAll(dir)

	cases := map[string]struct {
		GITGET_GETPATH string
		GETPATH        string
		want           string
	}{
		"default": {
			want: "~/src",
		},
		"env var": {
			GITGET_GETPATH: filepath.Join(dir, "new"),
			want:           filepath.Join(dir, "new"),
		},
		"deprecated env var": {
			GETPATH: filepath.Join(dir, "old"),
			want:    filepath.Join(dir, "old"),
		},
		"new env var overrides deprecated var": {
			GITGET_GETPATH: filepath.Join(dir, "new"),
			GETPATH:        filepath.Join(dir, "old"),
			want:           filepath.Join(dir, "new"),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// Empty GETPATH if set for tests
			os.Setenv("GITGET_GETPATH", "")
			os.Setenv("GETPATH", "")

			if c.GITGET_GETPATH != "" {
				os.Setenv("GITGET_GETPATH", c.GITGET_GETPATH)
			}
			if c.GETPATH != "" {
				os.Setenv("GETPATH", c.GETPATH)
			}

			path := configPath()
			if path != c.want {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.want)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	cases := map[string]struct {
		remote string
		want   string
	}{
		"git protocol": {
			remote: "git://github.com/arbourd/git-get.git",
			want:   "git://github.com/arbourd/git-get.git",
		},
		"https protocol": {
			remote: "https://github.com/arbourd/git-get.git",
			want:   "https://github.com/arbourd/git-get.git",
		},
		"ssh protocol": {
			remote: "git@github.com:arbourd/git-get.git",
			want:   "ssh://git@github.com/arbourd/git-get.git",
		},
		"no protocol": {
			remote: "github.com/arbourd/git-get",
			want:   "https://github.com/arbourd/git-get",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url, _ := ParseURL(c.remote)
			if url.String() != c.want {
				t.Fatalf("unexpected parsed url string:\n\t(GOT): %#v\n\t(WNT): %#v", url.String(), c.want)
			}
		})
	}
}

func TestDirectory(t *testing.T) {
	cases := map[string]struct {
		url  *url.URL
		want string
	}{
		"https protocol": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "arbourd/git-get",
			},
			want: "github.com/arbourd/git-get",
		},
		"ssh protocol": {
			url: &url.URL{
				Scheme: "ssh",
				Host:   "github.com",
				Path:   "arbourd/git-get",
			},
			want: "github.com/arbourd/git-get",
		},
		".git removal": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "arbourd/git-get.git",
			},
			want: "github.com/arbourd/git-get",
		},
		"multiple slashes": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "arbourd///git-get",
			},
			want: "github.com/arbourd/git-get",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir := Directory(c.url)
			if dir != filepath.Clean(c.want) {
				t.Fatalf("unexpected directory string:\n\t(GOT): %#v\n\t(WNT): %#v", dir, filepath.Clean(c.want))
			}
		})
	}
}

func TestClone(t *testing.T) {
	dir, _ := os.MkdirTemp("", "git-get")
	defer os.RemoveAll(dir)

	cases := map[string]struct {
		url  *url.URL
		want string
		err  bool
	}{
		"clone github": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "arbourd/git-get",
			},
			want: filepath.Join(dir, "github.com/arbourd/git-get"),
		},
		"clone ssh github": {
			url: &url.URL{
				Scheme: "ssh",
				User:   url.User("git"),
				Host:   "github.com",
				Path:   "arbourd/git-get.git",
			},
			want: filepath.Join(dir, "github.com/arbourd/git-get"),
		},
		"clone gitlab subgroups": {
			url: &url.URL{
				Scheme: "https",
				Host:   "gitlab.com",
				Path:   "gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin",
			},
			want: filepath.Join(dir, "gitlab.com/gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin"),
		},
		"invalid url": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "///arbourd/git-get",
			},
			err: true,
		},
		"not found": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "definitely-doesnt-exist",
			},
			err: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path, err := Clone(c.url, c.want)

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
