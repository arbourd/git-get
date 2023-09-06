package main

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

func main() {
	path, err := getPath()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Error: must provide a git repository url")
		os.Exit(1)
	}
	remote := args[0]

	url, err := ParseURL(remote)
	if err != nil {
		fmt.Printf("Error: unable to parse repository url: \"%s\"\n", remote)
		os.Exit(1)
	}

	dir := filepath.Join(path, ParseDirectory(url))

	resp, err := Download(url, dir)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Println(resp)
}

// DefaultPath is the path where repositories will be cloned to if GETPATH is unset.
const DefaultPath = "~/src"

// DefaultScheme is the scheme used when a URL is provided without one.
const DefaultScheme = "https"

// getPath gets the GETPATH path and creates the directory if needed.
func getPath() (string, error) {
	path := os.Getenv("GETPATH")
	if len(path) == 0 {
		path = DefaultPath
	}

	path, err := homedir.Expand(path)
	if err != nil {
		return "", err
	}

	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("GETPATH entry is relative; must be absolute path: \"%s\"", path)
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
	if len(u.Scheme) == 0 {
		u.Scheme = DefaultScheme
	}

	return u, err
}

// ParseDirectory parses the directory where the cloned repository will be downloaded from the URL
func ParseDirectory(u *url.URL) string {
	dir, _ := url.JoinPath(u.Host, u.Path)
	dir = strings.TrimSuffix(dir, ".git")
	return filepath.Clean(dir)
}

// Download clones the remote repository to the GETPATH and returns the directory.
func Download(u *url.URL, dir string) (string, error) {
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
