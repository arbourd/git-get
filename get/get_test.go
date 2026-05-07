package get

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
)

func TestAbsolutePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unable to detect homedir: %s", err)
	}

	defaultGetpath := filepath.Join(home, "src")
	configGetpath := t.TempDir()
	envGetpath := t.TempDir()
	err = gitConfigGlobalFixture(t)
	if err != nil {
		t.Fatalf("unable to setup test fixture: %s", err)
	}

	homeVar := "$HOME"
	if runtime.GOOS == "windows" {
		homeVar = "$USERPROFILE"
	}
	getpathWithVar := filepath.Join(homeVar, "src")

	cases := map[string]struct {
		gitConfigGetpath string
		envGetpath       string
		expectedPath     string
		wantErr          bool
	}{
		"default": {
			expectedPath: defaultGetpath,
		},
		"git config getpath": {
			gitConfigGetpath: configGetpath,
			expectedPath:     configGetpath,
		},
		"git config getpath with variable": {
			gitConfigGetpath: getpathWithVar,
			expectedPath:     defaultGetpath,
		},
		"env var getpath": {
			envGetpath:   envGetpath,
			expectedPath: envGetpath,
		},
		"env var getpath with variable": {
			gitConfigGetpath: getpathWithVar,
			expectedPath:     defaultGetpath,
		},
		"env var getpath overrides git config getpath": {
			gitConfigGetpath: configGetpath,
			envGetpath:       envGetpath,
			expectedPath:     envGetpath,
		},
		"relative GETPATH": {
			envGetpath: "../test",
			wantErr:    true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			setupEnv(t, c.gitConfigGetpath, c.envGetpath)

			path, err := AbsolutePath()
			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n")
			} else if path != c.expectedPath {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.expectedPath)
			}
		})
	}

	t.Run("tilde introduced by env var expansion", func(t *testing.T) {
		t.Setenv("TEST_GETPATH", "~/src")
		setupEnv(t, "", "$TEST_GETPATH")

		path, err := AbsolutePath()
		if err != nil {
			t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
		}
		if path != defaultGetpath {
			t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, defaultGetpath)
		}
	})
}

func TestShortPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unable to get home dir: %v", err)
	}

	cases := map[string]struct {
		path string
		want string
	}{
		"subdir of home": {
			path: filepath.Join(home, "src"),
			want: filepath.Join(defaultPrefix, "src"),
		},
		"home itself": {
			path: home,
			want: defaultPrefix,
		},
		"homeless": {
			path: "/root/dev",
			want: "/root/dev",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ShortPath(c.path); got != c.want {
				t.Fatalf("unexpected ShortPath(%q):\n\t(GOT): %q\n\t(WNT): %q", c.path, got, c.want)
			}
		})
	}
}

func TestParseURL(t *testing.T) {
	cases := map[string]struct {
		remote  string
		want    string
		wantErr bool
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
		"invalid url": {
			remote:  "github.com/arbourd/git-get%x",
			want:    "https://github.com/arbourd/git-get",
			wantErr: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url, err := ParseURL(c.remote)

			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n\t")
			} else if url != nil && url.String() != c.want {
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
			dir, err := Directory(c.url)
			if err != nil {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			}
			if dir != filepath.Clean(c.want) {
				t.Fatalf("unexpected directory string:\n\t(GOT): %#v\n\t(WNT): %#v", dir, filepath.Clean(c.want))
			}
		})
	}
}

func TestClone(t *testing.T) {
	dir := t.TempDir()

	cases := map[string]struct {
		url          *url.URL
		expectedPath string
		wantErr      bool
	}{
		"clone github": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "arbourd/git-get",
			},
			expectedPath: filepath.Join(dir, "github.com/arbourd/git-get"),
		},
		"clone ssh github": {
			url: &url.URL{
				Scheme: "ssh",
				User:   url.User("git"),
				Host:   "github.com",
				Path:   "arbourd/git-get.git",
			},
			expectedPath: filepath.Join(dir, "github.com/arbourd/git-get"),
		},
		"clone gitlab subgroups": {
			url: &url.URL{
				Scheme: "https",
				Host:   "gitlab.com",
				Path:   "gitlab-org/dev-subdepartment/ai-dev-promptcollection",
			},
			expectedPath: filepath.Join(dir, "gitlab.com/gitlab-org/dev-subdepartment/ai-dev-promptcollection"),
		},
		"invalid url": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "///arbourd/git-get",
			},
			wantErr: true,
		},
		"not found": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "definitely-doesnt-exist",
			},
			wantErr: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path, err := Clone(c.url, c.expectedPath)

			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n\t")
			} else if path != c.expectedPath {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.expectedPath)
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

// setupEnv unsets both global Git config and environmental GETPATHs, before setting them
// again if provided to ensure that the test environement is clean.
func setupEnv(t *testing.T, gitConfigGetpath, envGetpath string) {
	t.Helper()
	t.Setenv("GETPATH", envGetpath)

	_, _ = git.Config(config.Global, config.Unset(GitConfigKey, ""))
	if gitConfigGetpath != "" {
		if _, err := git.Config(config.Global, config.Entry(GitConfigKey, gitConfigGetpath)); err != nil {
			t.Fatalf("unable to set git config: %s", err)
		}
	}
}
