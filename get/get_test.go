package get

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/mitchellh/go-homedir"
)

func TestPath(t *testing.T) {
	home, _ := homedir.Dir()

	configDir, _ := os.MkdirTemp("", "git-get-gitconfig-")
	envDir, _ := os.MkdirTemp("", "git-get-envvar-")
	defer os.RemoveAll(configDir)
	defer os.RemoveAll(envDir)

	out, _ := git.Config(config.Global, config.Get(GitConfigKey, ""))
	if before := strings.TrimSpace(out); before != "" {
		// Set git-get global gitconfig at the end of tests if previously set
		defer git.Config(config.Global, config.Entry(GitConfigKey, before))
	} else {
		// Unset git-get global gitconfig at the end of tests if previously unset
		defer git.Config(config.Global, config.Unset(GitConfigKey, ""))
	}

	cases := map[string]struct {
		gitConfigGetPath string
		envVarGetPath    string
		expectedPath     string
		wantErr          bool
	}{
		"default": {
			expectedPath: filepath.Join(home, "src"),
		},
		"git config getpath": {
			envVarGetPath: filepath.Join(configDir),
			expectedPath:  filepath.Join(configDir),
		},
		"env var getpath": {
			envVarGetPath: filepath.Join(envDir),
			expectedPath:  filepath.Join(envDir),
		},
		"git config getpath over env var getpath": {
			gitConfigGetPath: filepath.Join(configDir),
			envVarGetPath:    filepath.Join(envDir),
			expectedPath:     filepath.Join(configDir),
		},
		"relative GETPATH": {
			envVarGetPath: "../test",
			wantErr:       true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// Empty GETPATH for tests
			git.Config(config.Global, config.Unset(GitConfigKey, ""))
			os.Setenv("GETPATH", "")

			if c.gitConfigGetPath != "" {
				git.Config(config.Global, config.Entry(GitConfigKey, c.gitConfigGetPath))
			}

			if c.envVarGetPath != "" {
				os.Setenv("GETPATH", c.envVarGetPath)
			}

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
	configDir, _ := os.MkdirTemp("", "git-get-gitconfig-")
	envDir, _ := os.MkdirTemp("", "git-get-envvar-")
	defer os.RemoveAll(configDir)
	defer os.RemoveAll(envDir)

	out, _ := git.Config(config.Global, config.Get(GitConfigKey, ""))
	if before := strings.TrimSpace(out); before != "" {
		// Set git-get global gitconfig at the end of tests if previously set
		defer git.Config(config.Global, config.Entry(GitConfigKey, before))
	} else {
		// Unset git-get global gitconfig at the end of tests if previously unset
		defer git.Config(config.Global, config.Unset(GitConfigKey, ""))
	}

	cases := map[string]struct {
		gitConfigGetPath string
		envVarGetPath    string
		expectedPath     string
	}{
		"default": {
			expectedPath: "~/src",
		},
		"git config getpath": {
			envVarGetPath: filepath.Join(configDir),
			expectedPath:  filepath.Join(configDir),
		},
		"env var getpath": {
			envVarGetPath: filepath.Join(envDir),
			expectedPath:  filepath.Join(envDir),
		},
		"git config getpath over env var getpath": {
			gitConfigGetPath: filepath.Join(configDir),
			envVarGetPath:    filepath.Join(envDir),
			expectedPath:     filepath.Join(configDir),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// Empty GETPATH for tests
			git.Config(config.Global, config.Unset(GitConfigKey, ""))
			os.Setenv("GETPATH", "")

			if c.gitConfigGetPath != "" {
				git.Config(config.Global, config.Entry(GitConfigKey, c.gitConfigGetPath))
			}

			if c.envVarGetPath != "" {
				os.Setenv("GETPATH", c.envVarGetPath)
			}

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
	dir, _ := os.MkdirTemp("", "git-get-")
	defer os.RemoveAll(dir)

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
