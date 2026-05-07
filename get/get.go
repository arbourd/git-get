package get

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ldez/go-git-cmd-wrapper/v2/clone"
	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

const (
	// defaultGetpath is the default GETPATH used when none is specified
	defaultGetpath = defaultPrefix + "/src"
	defaultPrefix  = "~"

	// defaultScheme is the scheme used when a URL is provided without one
	defaultScheme = "https"

	// GitConfigKey is the key that is used to store GETPATH information in the global Git config
	GitConfigKey = "get.path"

	// EnvKey is the name of the environmental variable that is used to store GETPATH information
	EnvKey = "GETPATH"
)

// AbsolutePath returns the absolute GETPATH, resolving env vars and ~ expansion.
// Precedence: GETPATH env var > get.path git config > default.
func AbsolutePath() (string, error) {
	p := os.Getenv(EnvKey)
	if p == "" {
		out, _ := git.Config(config.Global, config.Get(GitConfigKey, ""))
		p = strings.TrimSpace(out)
	}
	if p == "" {
		p = defaultGetpath
	}

	p = os.ExpandEnv(p)
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("detecting home directory: %w", err)
		}
		p = filepath.Join(home, p[1:])
	}

	if !filepath.IsAbs(p) {
		return "", fmt.Errorf("GETPATH is not an absolute path: %q", p)
	}
	return p, nil
}

// ShortPath returns a shortened version of the given path
func ShortPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}

	if rel == "." {
		return defaultPrefix
	}
	return filepath.Join(defaultPrefix, rel)
}

var scpSyntaxRe = regexp.MustCompile(`^(\w+)@([\w.-]+):(.*)$`)

// ParseURL parses and returns a URL from the remote string provided
func ParseURL(remote string) (*url.URL, error) {
	// Parse and return URL if valid SCP
	if m := scpSyntaxRe.FindStringSubmatch(remote); m != nil {
		// Match SCP-like syntax and convert it to a URL.
		// Eg, "git@github.com:user/repo" becomes
		// "ssh://git@github.com/user/repo".
		return &url.URL{
			Scheme: "ssh",
			User:   url.User(m[1]),
			Host:   m[2],
			Path:   m[3],
		}, nil
	}

	u, err := url.Parse(remote)
	if err != nil {
		return nil, fmt.Errorf("parsing url: %w", err)
	}

	if len(u.Scheme) == 0 {
		u.Scheme = defaultScheme
	}
	return u, nil
}

// Directory parses the directory where the cloned repository will be downloaded from the URL
func Directory(u *url.URL) (string, error) {
	dir, err := url.JoinPath(u.Host, u.Path)
	if err != nil {
		return "", fmt.Errorf("joining path: %w", err)
	}
	dir = strings.TrimSuffix(dir, ".git")
	return filepath.Clean(dir), nil
}

// Clone clones the remote repository to the GETPATH and returns the directory
func Clone(u *url.URL, dir string) (string, error) {
	// Check if git remote exists
	_, err := git.Raw("ls-remote", func(g *types.Cmd) {
		g.AddOptions(u.String())
	})
	if err != nil {
		sanitized := *u
		sanitized.User = nil
		return "", fmt.Errorf("git repository not found at %s: %w", sanitized.String(), err)
	}

	parentdir, _ := filepath.Split(dir)

	if !isGitRepository(dir) {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			return "", fmt.Errorf("%s exists but %s does not", dir, filepath.Join(dir, ".git"))
		}

		err := os.MkdirAll(parentdir, 0755)
		if err != nil {
			return "", fmt.Errorf("creating clone directory: %w", err)
		}

		_, err = git.Clone(clone.Repository(u.String()), clone.Directory(dir))
		if err != nil {
			return "", fmt.Errorf("git clone: %w", err)
		}
	}

	return dir, nil
}

func isGitRepository(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
