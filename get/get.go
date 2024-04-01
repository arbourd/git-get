package get

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ldez/go-git-cmd-wrapper/v2/clone"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
	"github.com/mitchellh/go-homedir"
)

// DefaultGetPath is the default path where repositories will be cloned if not configured.
const defaultGetPath = "~/src"

// DefaultScheme is the scheme used when a URL is provided without one.
const defaultScheme = "https"

// Path returns the absolute GETPATH.
func Path() (string, error) {
	path := configPath()

	path, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("GETPATH entry is relative; must be an absolute path: \"%s\"", path)
	}

	// Make GETPATH directory if it does not exist.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return path, nil
}

// configPath returns the GETPATH from the config, environment or default.
func configPath() string {
	path := os.Getenv("GITGET_GETPATH")
	if len(path) != 0 {
		return path
	}

	path = os.Getenv("GETPATH")
	if len(path) != 0 {
		fmt.Println("warning: $GETPATH has been deprecated; use $GITGET_GETPATH")
		return path
	}

	return defaultGetPath
}

var scpSyntaxRe = regexp.MustCompile(`^(\w+)@([\w.-]+):(.*)$`)

// ParseURL parses and returns a URL from the remote string provided.
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
		return nil, err
	}

	if len(u.Scheme) == 0 {
		u.Scheme = defaultScheme
	}
	return u, nil
}

// Directory parses the directory where the cloned repository will be downloaded from the URL.
func Directory(u *url.URL) string {
	dir, _ := url.JoinPath(u.Host, u.Path)
	dir = strings.TrimSuffix(dir, ".git")
	return filepath.Clean(dir)
}

// Clone clones the remote repository to the GETPATH and returns the directory.
func Clone(u *url.URL, dir string) (string, error) {
	// Check if git remote exists
	_, err := git.Raw("ls-remote", func(g *types.Cmd) {
		g.AddOptions(u.String())
	})
	if err != nil {
		return "", fmt.Errorf("git repository not found: %s", u.String())
	}

	parentdir, _ := filepath.Split(dir)
	gitdir := filepath.Join(dir, ".git")

	if _, err := os.Stat(gitdir); os.IsNotExist(err) {
		// Check if root folder exists, even though the .git directory does not.
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			return "", fmt.Errorf("%s exists but %s does not", dir, gitdir)
		}

		err := os.MkdirAll(parentdir, os.ModePerm)
		if err != nil {
			return "", err
		}

		_, err = git.Clone(clone.Repository(u.String()), clone.Directory(dir))
		if err != nil {
			return "", err
		}
	}

	return dir, nil
}
