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

func TestPath(t *testing.T) {
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
		"env var getpath": {
			envGetpath:   envGetpath,
			expectedPath: envGetpath,
		},
		"git config getpath overrides env var getpath": {
			gitConfigGetpath: configGetpath,
			envGetpath:       envGetpath,
			expectedPath:     configGetpath,
		},
		"relative GETPATH": {
			envGetpath: "../test",
			wantErr:    true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			setupEnv(c.gitConfigGetpath, c.envGetpath)

			path, err := Path()
			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n")
			} else if path != c.expectedPath {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.expectedPath)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	configGetpath := t.TempDir()
	envGetpath := t.TempDir()
	err := gitConfigGlobalFixture(t)
	if err != nil {
		t.Fatalf("unable to setup test fixture: %s", err)
	}

	cases := map[string]struct {
		gitConfigGetpath string
		envGetpath       string
		expectedPath     string
	}{
		"default": {
			expectedPath: "~/src",
		},
		"git config getpath": {
			envGetpath:   configGetpath,
			expectedPath: configGetpath,
		},
		"env var getpath": {
			envGetpath:   envGetpath,
			expectedPath: envGetpath,
		},
		"git config getpath over env var getpath": {
			gitConfigGetpath: configGetpath,
			envGetpath:       envGetpath,
			expectedPath:     configGetpath,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			setupEnv(c.gitConfigGetpath, c.envGetpath)

			path := configPath()
			if path != c.expectedPath {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.expectedPath)
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
			dir := Directory(c.url)
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
				Path:   "gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin",
			},
			expectedPath: filepath.Join(dir, "gitlab.com/gitlab-org/dev-subdepartment/ai-experimentation-chrome-plugin"),
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

// gitConfigGlobalFixture creates a temporary folder and .gitconfig to be used as the global
// Git config for tests.
func gitConfigGlobalFixture(t *testing.T) error {
	// Skip fixture on Windows in CI
	if os.Getenv("CI") == "true" && runtime.GOOS == "windows" {
		return nil
	}

	gitconfig := filepath.Join(t.TempDir(), ".gitconfig")

	_, err := os.Create(gitconfig)
	if err != nil {
		return fmt.Errorf("unable create .gitconfig: %s", err)
	}

	err = os.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
	if err != nil {
		return fmt.Errorf("unable to set GIT_CONFIG_GLOBAL: %s", err)
	}

	return nil
}

// setupEnv unsets both global Git config and environmental GETPATHs, before setting them
// again if provided to ensure that the test environement is clean.
func setupEnv(gitConfigGetpath, envGetpath string) {
	os.Setenv("GETPATH", "")
	if envGetpath != "" {
		os.Setenv("GETPATH", envGetpath)
	}

	git.Config(config.Global, config.Entry(GitConfigKey, ""))
	if gitConfigGetpath != "" {
		git.Config(config.Global, config.Entry(GitConfigKey, gitConfigGetpath))
	}
}
