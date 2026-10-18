package get

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~"+string(filepath.Separator)) {
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
		_, userinfo, _ := splitUserinfo(remote)
		if userinfo == "" {
			return nil, fmt.Errorf("parsing url: %w", err)
		}
		// err quotes the raw URL, and a password containing "/" can land in ue.Err, so only
		// keep ue.Err when the userinfo has no password.
		var ue *url.Error
		if errors.As(err, &ue) && !strings.Contains(userinfo, ":") {
			return nil, fmt.Errorf("parsing url: %w", ue.Err)
		}
		return nil, errors.New("parsing url: invalid URL")
	}

	// Opaque means no "//" after the scheme (e.g. "user:pw@host:o/r"); Directory can't map it.
	if u.Opaque != "" {
		return nil, errors.New("parsing url: unsupported format")
	}

	if len(u.Scheme) == 0 {
		u.Scheme = defaultScheme
	}
	return u, nil
}

// Directory parses the directory where the cloned repository will be downloaded from the URL.
// The returned directory is always local (never escapes upward via ".."), so any GETPATH it is
// joined to remains safely contained.
func Directory(u *url.URL) (string, error) {
	dir, err := url.JoinPath(u.Host, u.Path)
	if err != nil {
		return "", fmt.Errorf("joining path: %w", err)
	}
	dir = strings.TrimSuffix(dir, ".git")
	dir = filepath.Clean(dir)

	if !filepath.IsLocal(dir) {
		return "", fmt.Errorf("directory %q escapes GETPATH", dir)
	}
	return dir, nil
}

// Clone clones the remote repository to the GETPATH and returns the directory.
// If verbose is non-nil, each git command and its output is written to it.
func Clone(u *url.URL, dir string, verbose io.Writer) (string, error) {
	if isGitRepository(dir) {
		if verbose != nil {
			fmt.Fprintf(verbose, "already cloned: %s\n", dir)
		}
		return dir, nil
	}

	if _, statErr := os.Stat(dir); statErr == nil {
		return "", fmt.Errorf("%s exists but %s does not", dir, filepath.Join(dir, ".git"))
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return "", fmt.Errorf("checking directory: %w", statErr)
	}

	// Check if git remote exists before creating any directories
	_, err := git.Raw("ls-remote", func(g *types.Cmd) {
		g.AddOptions(u.String())
		g.AddOptions("HEAD")
	}, verboseOption(verbose, u))
	if err != nil {
		return "", fmt.Errorf("git repository not found at %s: %w", redact(u), err)
	}

	parentdir, _ := filepath.Split(dir)
	if err := os.MkdirAll(parentdir, 0755); err != nil {
		return "", fmt.Errorf("creating clone directory: %w", err)
	}

	if _, err = git.Clone(clone.Repository(u.String()), clone.Directory(dir), verboseOption(verbose, u)); err != nil {
		return "", fmt.Errorf("git clone: %w", err)
	}

	return dir, nil
}

// verboseOption wraps the command executor to write the git command line and
// its output to w. Credentials in u are redacted from the command line.
func verboseOption(w io.Writer, u *url.URL) types.Option {
	return func(g *types.Cmd) {
		if w == nil {
			return
		}

		raw, safe := u.String(), redact(u).String()
		next := g.Executor
		g.Executor = func(ctx context.Context, name string, debug bool, args ...string) (string, error) {
			line := []string{name}
			for _, a := range args {
				line = append(line, shellQuote(strings.ReplaceAll(a, raw, safe)))
			}
			fmt.Fprintln(w, "+", strings.Join(line, " "))

			out, err := next(ctx, name, debug, args...)
			fmt.Fprint(w, out)
			return out, err
		}
	}
}

// redact returns a copy of u that is safe to print: userinfo is dropped, like git does in
// its own messages, except a bare ssh username (e.g. git@), which isn't a secret and keeps
// the logged command identical to the one that ran.
func redact(u *url.URL) *url.URL {
	r := *u
	if _, hasPassword := u.User.Password(); u.Scheme != "ssh" || hasPassword {
		r.User = nil
	}
	return &r
}

// RedactRemote drops userinfo from a raw, possibly unparseable, remote string so it can be
// put in an error message.
func RedactRemote(s string) string {
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		return redact(u).String()
	}
	// Like git's messages, this drops a scp-like username too.
	scheme, _, rest := splitUserinfo(s)
	return scheme + rest
}

// splitUserinfo splits a raw, possibly unparseable, remote into its scheme (with "://"),
// userinfo and the rest. A password may contain "/" or "@", so userinfo runs to the last "@".
func splitUserinfo(s string) (scheme, userinfo, rest string) {
	i := strings.LastIndex(s, "@")
	userinfo, rest = s[:max(i, 0)], s[i+1:]
	if sch, after, ok := strings.Cut(userinfo, "://"); ok {
		scheme, userinfo = sch+"://", after
	}
	return scheme, userinfo, rest
}

var shellSafe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// shellQuote quotes s so the logged command can be pasted into a POSIX shell.
func shellQuote(s string) string {
	if shellSafe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func isGitRepository(path string) bool {
	_, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil
}
