package get

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

func TestAbsolutePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("unable to detect homedir: %s", err)
	}

	defaultGetpath := filepath.Join(home, "src")
	configGetpath := t.TempDir()
	envGetpath := t.TempDir()
	gitConfigGlobalFixture(t)

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
			envGetpath:   getpathWithVar,
			expectedPath: defaultGetpath,
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
		"tilde-user GETPATH": {
			envGetpath: "~otheruser/src",
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
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

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
			path: filepath.Join(filepath.Dir(home), "other"),
			want: filepath.Join(filepath.Dir(home), "other"),
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
		"ssh hyphenated user": {
			remote: "my-user@git.example.com:owner/repo.git",
			want:   "ssh://my-user@git.example.com/owner/repo.git",
		},
		"ssh dotted user": {
			remote: "my.user@git.example.com:owner/repo.git",
			want:   "ssh://my.user@git.example.com/owner/repo.git",
		},
		"ssh user with shell metacharacters": {
			remote:  "$(id>x)@evil.host:o/r",
			wantErr: true,
		},
		"ssh user with leading hyphen": {
			remote:  "-oProxyCommand=true@example.test:o/r",
			wantErr: true,
		},
		"user with slash is not scp": {
			remote: "foo/bar@host:o/r",
			want:   "https://foo/bar@host:o/r",
		},
		"scp without user unsupported": {
			remote:  "git.example.com:owner/repo.git",
			wantErr: true,
		},
		"no protocol": {
			remote: "github.com/arbourd/git-get",
			want:   "https://github.com/arbourd/git-get",
		},
		"invalid url": {
			remote:  "github.com/arbourd/git-get%x",
			wantErr: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			u, err := ParseURL(c.remote)

			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n\t")
			} else if u != nil && u.String() != c.want {
				t.Fatalf("unexpected parsed url string:\n\t(GOT): %#v\n\t(WNT): %#v", u.String(), c.want)
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
		"ssh port is dropped": {
			url: &url.URL{
				Scheme: "ssh",
				Host:   "git.example.com:2222",
				Path:   "/arbourd/git-get",
			},
			want: "git.example.com/arbourd/git-get",
		},
		"https port is dropped": {
			url: &url.URL{
				Scheme: "https",
				Host:   "git.example.com:8443",
				Path:   "/arbourd/git-get.git",
			},
			want: "git.example.com/arbourd/git-get",
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
		"segment starting with dots is not treated as traversal": {
			url: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "..cache/repo",
			},
			want: "example.com/..cache/repo",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir, err := Directory(c.url)
			if err != nil {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			}
			if want := filepath.Clean(c.want); dir != want {
				t.Fatalf("unexpected directory string:\n\t(GOT): %#v\n\t(WNT): %#v", dir, want)
			}
		})
	}
}

// TestDirectoryNeverEscapes checks that traversal input can't escape GETPATH. Directory
// neutralizes these inputs rather than rejecting them, so each case pins the resulting path.
func TestDirectoryNeverEscapes(t *testing.T) {
	cases := map[string]struct {
		url  *url.URL
		want string
	}{
		"traversal via path": {
			url:  &url.URL{Scheme: "https", Host: "example.com", Path: "../../../etc/passwd"},
			want: filepath.FromSlash("etc/passwd"),
		},
		"traversal via host": {
			url:  &url.URL{Scheme: "ssh", Host: "..", Path: "/../../etc/passwd"},
			want: filepath.FromSlash("etc/passwd"),
		},
		"traversal split across host and path": {
			url:  &url.URL{Scheme: "https", Host: "evil.com/../../etc", Path: "passwd"},
			want: filepath.FromSlash("etc/passwd"),
		},
		"empty host and path": {
			url:  &url.URL{Scheme: "https", Host: "", Path: ""},
			want: ".",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir, err := Directory(c.url)
			if err != nil {
				t.Fatalf("Directory(%+v) unexpected error: %v", c.url, err)
			}
			if !filepath.IsLocal(dir) {
				t.Fatalf("Directory(%+v) returned non-local directory %q", c.url, dir)
			}
			if dir != c.want {
				t.Fatalf("unexpected directory:\n\t(GOT): %q\n\t(WNT): %q", dir, c.want)
			}
		})
	}
}

func TestClone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test")
	}
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	if os.Getenv("CI") != "" {
		// CI has no SSH auth, and `git clone` ignores the repo-local rewrite the workflow sets.
		t.Setenv("GIT_CONFIG_COUNT", "1")
		t.Setenv("GIT_CONFIG_KEY_0", "url.https://github.com/.insteadOf")
		t.Setenv("GIT_CONFIG_VALUE_0", "ssh://git@github.com/")
	}

	// each case clones into its own temp dir so cases cannot short-circuit one another
	cases := map[string]struct {
		url     *url.URL
		subdir  string
		wantErr bool
	}{
		"clone github": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "arbourd/git-get",
			},
			subdir: "github.com/arbourd/git-get",
		},
		"clone ssh github": {
			url: &url.URL{
				Scheme: "ssh",
				User:   url.User("git"),
				Host:   "github.com",
				Path:   "arbourd/git-get.git",
			},
			subdir: "github.com/arbourd/git-get",
		},
		"clone gitlab subgroups": {
			url: &url.URL{
				Scheme: "https",
				Host:   "gitlab.com",
				Path:   "gitlab-org/dev-subdepartment/ai-dev-promptcollection",
			},
			subdir: "gitlab.com/gitlab-org/dev-subdepartment/ai-dev-promptcollection",
		},
		"invalid url": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "///arbourd/git-get",
			},
			subdir:  "github.com/arbourd/git-get",
			wantErr: true,
		},
		"not found": {
			url: &url.URL{
				Scheme: "https",
				Host:   "github.com",
				Path:   "definitely-doesnt-exist",
			},
			subdir:  "github.com/definitely-doesnt-exist",
			wantErr: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), c.subdir)
			wantPath := target
			if c.wantErr {
				wantPath = ""
			}

			path, err := Clone(c.url, target, nil)

			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n\t")
			} else if path != wantPath {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, wantPath)
			}

			_, statErr := os.Stat(filepath.Join(target, ".git"))
			if c.wantErr && statErr == nil {
				t.Fatalf("failed clone left a repository at %s", target)
			} else if !c.wantErr && statErr != nil {
				t.Fatalf("clone did not produce a repository: %v", statErr)
			}
		})
	}
}

func TestCloneOffline(t *testing.T) {
	t.Setenv("GIT_TERMINAL_PROMPT", "0")

	cases := map[string]struct {
		repo         string
		remote       bool   // a bare remote exists for repo
		existingDir  string // subdir of the target to create beforehand; "." is the target itself
		wantErr      string // substring of the expected error
		wantGit      bool   // target/.git exists afterwards
		wantOrigin   bool   // clone recorded the requested URL as origin
		wantNoParent bool   // the target's parent directory was never created
	}{
		"successful clone": {
			repo:       "o/r",
			remote:     true,
			wantGit:    true,
			wantOrigin: true,
		},
		// no remote: reaching ls-remote would fail, so success proves the early return
		"already cloned returns early": {
			repo:        "o/r",
			existingDir: ".git",
			wantGit:     true,
		},
		"directory exists without .git": {
			repo:        "o/r",
			remote:      true,
			existingDir: ".",
			wantErr:     "exists but",
		},
		"failing ls-remote creates no directories": {
			repo:         "o/missing",
			wantErr:      "git repository not found",
			wantNoParent: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			setupExampleRemote(t, c.remote, c.repo)

			target := filepath.Join(t.TempDir(), "example.test", filepath.FromSlash(c.repo))
			if c.existingDir != "" {
				if err := os.MkdirAll(filepath.Join(target, c.existingDir), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			u := &url.URL{Scheme: "https", Host: "example.test", Path: c.repo}
			path, err := Clone(u, target, nil)

			wantPath := target
			if c.wantErr != "" {
				wantPath = ""
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("unexpected error:\n\t(GOT): %v\n\t(WNT): contains %q", err, c.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if path != wantPath {
				t.Fatalf("unexpected path:\n\t(GOT): %q\n\t(WNT): %q", path, wantPath)
			}

			_, err = os.Stat(filepath.Join(target, ".git"))
			if gotGit := err == nil; gotGit != c.wantGit {
				t.Fatalf(".git present at %s = %t, want %t", target, gotGit, c.wantGit)
			}

			if c.wantOrigin {
				origin, err := git.Raw("config", func(g *types.Cmd) {
					g.AddOptions("--file")
					g.AddOptions(filepath.Join(target, ".git", "config"))
					g.AddOptions("remote.origin.url")
				})
				if err != nil {
					t.Fatalf("reading origin: %v", err)
				}
				if got := strings.TrimSpace(origin); got != u.String() {
					t.Fatalf("unexpected origin:\n\t(GOT): %q\n\t(WNT): %q", got, u.String())
				}
			}

			// ls-remote must fail before any directory is created
			if c.wantNoParent {
				if _, err := os.Stat(filepath.Dir(target)); !errors.Is(err, fs.ErrNotExist) {
					t.Fatalf("parent directory created before ls-remote succeeded: %v", err)
				}
			}
		})
	}
}

// setupExampleRemote mirrors the fixture in TestRunVerbose (main_test.go); keep them in sync.
// It points https://example.test/ at a local directory through the
// global git config, creating a bare repository at repo inside it if bare is set.
func setupExampleRemote(t *testing.T, bare bool, repo string) {
	t.Helper()
	gitConfigGlobalFixture(t)

	remotes := t.TempDir()
	if bare {
		if _, err := git.Raw("init", func(g *types.Cmd) {
			g.AddOptions("--bare")
			g.AddOptions(filepath.Join(remotes, filepath.FromSlash(repo)))
		}); err != nil {
			t.Fatalf("git init: %v", err)
		}
	}

	// file:///C:/... on Windows, file:///tmp/... elsewhere.
	remoteURL := "file:///" + strings.TrimPrefix(filepath.ToSlash(remotes), "/") + "/"
	cfg := fmt.Sprintf("[url %q]\n\tinsteadOf = https://example.test/\n", remoteURL)
	if err := os.WriteFile(os.Getenv("GIT_CONFIG_GLOBAL"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write gitconfig: %v", err)
	}
}

func TestVerboseOption(t *testing.T) {
	plain := &url.URL{Scheme: "https", Host: "github.com", Path: "arbourd/git-get"}

	cases := map[string]struct {
		url     *url.URL
		extra   []string
		wantLog string
	}{
		"logs command and output": {
			url:     plain,
			wantLog: "+ git ls-remote https://github.com/arbourd/git-get\n",
		},
		"redacts password": {
			url:     &url.URL{Scheme: "https", User: url.UserPassword("user", "secret"), Host: "github.com", Path: "arbourd/git-get"},
			wantLog: "+ git ls-remote https://github.com/arbourd/git-get\n",
		},
		"redacts token used as username": {
			url:     &url.URL{Scheme: "https", User: url.User("ghp_token123"), Host: "github.com", Path: "arbourd/git-get"},
			wantLog: "+ git ls-remote https://github.com/arbourd/git-get\n",
		},
		"keeps ssh username so the logged command matches the one that ran": {
			url:     &url.URL{Scheme: "ssh", User: url.User("git"), Host: "github.com", Path: "arbourd/git-get"},
			wantLog: "+ git ls-remote ssh://git@github.com/arbourd/git-get\n",
		},
		"redacts ssh password but keeps username": {
			url:     &url.URL{Scheme: "ssh", User: url.UserPassword("git", "secret"), Host: "github.com", Path: "arbourd/git-get"},
			wantLog: "+ git ls-remote ssh://github.com/arbourd/git-get\n",
		},
		"quotes args with spaces": {
			url:     plain,
			extra:   []string{"/Users/me/my src/repo", "it's"},
			wantLog: "+ git ls-remote https://github.com/arbourd/git-get '/Users/me/my src/repo' 'it'\\''s'\n",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var log bytes.Buffer
			var gotArgs []string

			cmd := types.NewCmd("ls-remote")
			cmd.Executor = func(_ context.Context, _ string, _ bool, args ...string) (string, error) {
				gotArgs = args
				return "fake output\n", nil
			}
			cmd.AddOptions(c.url.String())
			for _, e := range c.extra {
				cmd.AddOptions(e)
			}

			verboseOption(&log, c.url)(cmd)

			if _, err := cmd.Exec(context.Background(), "git", false, cmd.Options...); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// the real command must still receive the unredacted URL
			if len(gotArgs) < 2 || gotArgs[1] != c.url.String() {
				t.Fatalf("executor received unexpected args: %#v", gotArgs)
			}
			if want := c.wantLog + "fake output\n"; log.String() != want {
				t.Fatalf("unexpected log:\n\t(GOT): %q\n\t(WNT): %q", log.String(), want)
			}
		})
	}
}

func gitConfigGlobalFixture(t *testing.T) {
	t.Helper()
	gitconfig := filepath.Join(t.TempDir(), ".gitconfig")
	if err := os.WriteFile(gitconfig, nil, 0o600); err != nil {
		t.Fatalf("unable to create .gitconfig: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
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

// TestCloneVerboseNoCredentialLeak runs the real git against an unreachable
// host to check that neither git's own output nor our log echoes credentials.
func TestCloneVerboseNoCredentialLeak(t *testing.T) {
	cases := map[string]struct {
		url     *url.URL
		secrets []string
	}{
		"https user and password": {
			url:     &url.URL{Scheme: "https", User: url.UserPassword("someuser", "s3cret"), Host: "127.0.0.1:1", Path: "o/r"},
			secrets: []string{"s3cret", "someuser"},
		},
		"https token as username": {
			url:     &url.URL{Scheme: "https", User: url.User("ghp_token123"), Host: "127.0.0.1:1", Path: "o/r"},
			secrets: []string{"ghp_token123"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			gitConfigGlobalFixture(t)
			t.Setenv("GIT_TERMINAL_PROMPT", "0")

			var log bytes.Buffer
			_, err := Clone(c.url, filepath.Join(t.TempDir(), "o", "r"), &log)
			if err == nil {
				t.Fatal("expected error cloning from an unreachable host")
			}

			t.Logf("log:\n%s", log.String())
			if !strings.Contains(log.String(), "+ git ls-remote") {
				t.Fatalf("log missing the logged git command: %q", log.String())
			}
			for _, s := range c.secrets {
				if bytes.Contains(log.Bytes(), []byte(s)) {
					t.Errorf("log leaked %q", s)
				}
				if strings.Contains(err.Error(), s) {
					t.Errorf("error leaked %q: %v", s, err)
				}
			}
		})
	}
}

func TestRedactRemote(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"https user and password": {"https://user:pw@host/o/r", "https://host/o/r"},
		"https token username":    {"https://ghp_tok@host/o/r", "https://host/o/r"},
		"no userinfo":             {"https://host/o/r", "https://host/o/r"},
		"scp without password":    {"git@host:o/r", "host:o/r"},
		"scp-like with password":  {"user:pw@host:o/r", "host:o/r"},
		"scp-like password slash": {"user:pa/ss@host:o/r", "host:o/r"},
		"scp-like password @":     {"user:p@ss@host:o/r", "host:o/r"},
		"https password slash":    {"https://user:pa/ss@host/o/r", "https://host/o/r"},
		"https password @":        {"https://user:p@ss@host/o/r", "https://host/o/r"},
		"@ in path":               {"https://host/o/r@v1.2", "https://host/o/r@v1.2"},
		"option":                  {"--verbos", "--verbos"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if got := RedactRemote(c.in); got != c.want {
				t.Fatalf("RedactRemote(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestParseURLErrorKeepsReasonWithoutPassword(t *testing.T) {
	_, err := ParseURL("ssh://git@host:22x/o/r")
	if err == nil || !strings.Contains(err.Error(), `invalid port ":22x"`) {
		t.Fatalf("want invalid port reason, got %v", err)
	}
}

func TestSecretsNeverPrinted(t *testing.T) {
	tests := map[string]struct {
		input   string
		secrets []string
	}{
		"password with @":        {"https://user:p@ssw0rd@host/o/r", []string{"ssw0rd", "user"}},
		"password with slash":    {"https://user:pa/ssw0rd@host/o/r", []string{"ssw0rd", "user"}},
		"opaque with password":   {"user:pw5ecret@host:o/r", []string{"pw5ecret", "user"}},
		"token as ssh username":  {"ssh://ghp_tok3n:x@host/o/r", []string{"ghp_tok3n"}},
		"bad port with password": {"https://user:s3cret@host:badport/o/r", []string{"s3cret"}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var printed []string
			printed = append(printed, RedactRemote(tt.input))
			if u, err := ParseURL(tt.input); err != nil {
				printed = append(printed, err.Error())
			} else {
				printed = append(printed, redact(u).String())
			}

			for _, out := range printed {
				for _, secret := range tt.secrets {
					if strings.Contains(out, secret) {
						t.Errorf("output %q leaks %q", out, secret)
					}
				}
			}
		})
	}
}
