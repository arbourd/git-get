package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestBashCompletion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash completion not supported on Windows")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}

	cases := map[string]struct {
		words   string // COMP_WORDS contents
		cword   int
		want    []string
		notWant []string
	}{
		"repository": {
			words: `"git-get" "github.com/arbourd/"`,
			cword: 1,
			want:  []string{"github.com/arbourd/git-get"},
		},
		"repository after -v": {
			words: `"git-get" "-v" "github.com/arbourd/"`,
			cword: 2,
			want:  []string{"github.com/arbourd/git-get"},
		},
		"repository after git get -v": {
			words: `"git" "get" "--verbose" "github.com/arbourd/"`,
			cword: 3,
			want:  []string{"github.com/arbourd/git-get"},
		},
		"nothing after repository": {
			words:   `"git-get" "github.com/arbourd/git-get" "github.com/arbourd/"`,
			cword:   2,
			notWant: []string{"github.com/arbourd/git-get"},
		},
		"flags": {
			words: `"git-get" "--v"`,
			cword: 1,
			want:  []string{"--verbose", "--version"},
		},
	}

	env := completionEnv(t)
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command("bash", "-c", fmt.Sprintf(`source completions/git-get.bash
COMP_WORDS=(%s)
COMP_CWORD=%d
_git_get
printf '%%s\n' "${COMPREPLY[@]}"`, c.words, c.cword))
			cmd.Env = env

			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("bash: %v", err)
			}
			assertContains(t, string(out), c.want, c.notWant)
		})
	}
}

func TestFishCompletion(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish not on PATH")
	}

	cases := map[string]struct {
		line    string
		want    []string
		notWant []string
	}{
		"repository": {
			line: "git-get github.com/arbourd/",
			want: []string{"github.com/arbourd/git-get"},
		},
		"repository after -v": {
			line: "git-get -v github.com/arbourd/",
			want: []string{"github.com/arbourd/git-get"},
		},
		"nothing after repository": {
			line:    "git-get github.com/arbourd/git-get github.com/arbourd/",
			notWant: []string{"github.com/arbourd/git-get"},
		},
		"flags": {
			line: "git-get --v",
			want: []string{"--verbose", "--version"},
		},
		"flags via git get": {
			line: "git get --v",
			want: []string{"--verbose", "--version"},
		},
	}

	env := completionEnv(t)
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command("fish", "-c", fmt.Sprintf(`source completions/git-get.fish
complete --do-complete %q`, c.line))
			cmd.Env = env

			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("fish: %v", err)
			}
			assertContains(t, string(out), c.want, c.notWant)
		})
	}
}

func TestZshCompletion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("zsh completion not supported on Windows")
	}
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not on PATH")
	}

	// Under git's zsh wrapper (wrapper=true) CURRENT is one lower than the real index while
	// words is unshifted, and cword (the real index - 1) is set; natively cword is unset.
	cases := map[string]struct {
		words   string
		current int
		wrapper bool
		want    []string
		notWant []string
	}{
		"native repository": {
			words: `git-get github.com/arbourd/`, current: 2,
			want: []string{"github.com/arbourd/git-get"},
		},
		"native nothing after repository": {
			words: `git-get github.com/arbourd/git-get github.com/arbourd/`, current: 3,
			notWant: []string{"github.com/arbourd/git-get"},
		},
		"wrapper repository after -v": {
			words: `git get -v github.com/arbourd/`, current: 3, wrapper: true,
			want: []string{"github.com/arbourd/git-get"},
		},
		"wrapper nothing after repository": {
			words: `git get github.com/arbourd/git-get github.com/arbourd/`, current: 3, wrapper: true,
			notWant: []string{"github.com/arbourd/git-get"},
		},
		"wrapper nothing after -v repository": {
			words: `git get -v github.com/arbourd/git-get github.com/arbourd/`, current: 4, wrapper: true,
			notWant: []string{"github.com/arbourd/git-get"},
		},
		"wrapper flags": {
			words: `git get --v`, current: 2, wrapper: true,
			want: []string{"--verbose", "--version"},
		},
	}

	env := completionEnv(t)
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			wrapper := ""
			if c.wrapper {
				wrapper = fmt.Sprintf("cword=%d", c.current)
			}
			// The current word is the last one; compadd is stubbed to print what would be offered.
			cmd := exec.Command("zsh", "-c", fmt.Sprintf(`source completions/git-get.zsh
compadd() { local a; for a in "$@"; do [[ -n $a && $a != -S && $a != -- ]] && print -r -- $a; done }
words=(%s)
CURRENT=%d
cur=${words[-1]}
%s
_git_get`, c.words, c.current, wrapper))
			cmd.Env = env

			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("zsh: %v", err)
			}
			assertContains(t, string(out), c.want, c.notWant)
		})
	}
}

func TestPowerShellCompletion(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not on PATH")
	}

	cases := map[string]struct {
		line    string
		want    []string
		notWant []string
	}{
		"repository": {
			line: "git-get github.com/arbourd/",
			want: []string{"github.com/arbourd/git-get"},
		},
		"repository after -v": {
			line: "git-get -v github.com/arbourd/",
			want: []string{"github.com/arbourd/git-get"},
		},
		"nothing after repository": {
			line:    "git-get github.com/arbourd/git-get github.com/arbourd/",
			notWant: []string{"github.com/arbourd/git-get"},
		},
		"flags": {
			line: "git-get --v",
			want: []string{"--verbose", "--version"},
		},
		"repository via git get": {
			line: "git get -v github.com/arbourd/",
			want: []string{"github.com/arbourd/git-get"},
		},
		"nothing after repository via git get": {
			line:    "git get github.com/arbourd/git-get github.com/arbourd/",
			notWant: []string{"github.com/arbourd/git-get"},
		},
		"nothing after repository via git.exe": {
			line:    "git.exe get github.com/arbourd/git-get github.com/arbourd/",
			notWant: []string{"github.com/arbourd/git-get"},
		},
	}

	env := completionEnv(t)
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command("pwsh", "-NoProfile", "-Command", fmt.Sprintf(`. ./completions/git-get.ps1
$line = %s
(TabExpansion2 -inputScript $line -cursorColumn $line.Length).CompletionMatches | ForEach-Object { $_.CompletionText }`, "'"+c.line+"'"))
			cmd.Env = env

			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("pwsh: %v", err)
			}
			assertContains(t, string(out), c.want, c.notWant)
		})
	}
}

func assertContains(t *testing.T, got string, want, notWant []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("output missing %q: %q", w, got)
		}
	}
	for _, w := range notWant {
		if strings.Contains(got, w) {
			t.Errorf("output unexpectedly contains %q: %q", w, got)
		}
	}
}

// binDir holds the git-get binary shared by all completion tests; TestMain removes it.
var binDir string

var buildCompletionBinary = sync.OnceValues(func() (string, error) {
	var err error
	if binDir, err = os.MkdirTemp("", "git-get-bin"); err != nil {
		return "", err
	}
	bin := filepath.Join(binDir, "git-get")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		return "", fmt.Errorf("go build: %w\n%s", err, out)
	}
	return bin, nil
})

func TestMain(m *testing.M) {
	code := m.Run()
	if binDir != "" {
		os.RemoveAll(binDir)
	}
	os.Exit(code)
}

// completionEnv sets up a git config fixture, the binary on PATH and a seeded GETPATH.
func completionEnv(t *testing.T) []string {
	t.Helper()
	if err := gitConfigGlobalFixture(t); err != nil {
		t.Fatalf("setup: %v", err)
	}
	bin, err := buildCompletionBinary()
	if err != nil {
		t.Fatal(err)
	}
	getpath := t.TempDir()
	seedRepos(t, getpath, []string{"github.com/arbourd/git-get"})

	pathSep := ":"
	if runtime.GOOS == "windows" {
		pathSep = ";"
	}

	env := make([]string, 0, len(os.Environ())+3)
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "PATH=") && !strings.HasPrefix(e, "GETPATH=") && !strings.HasPrefix(e, "HOME=") {
			env = append(env, e)
		}
	}

	return append(env,
		"PATH="+filepath.Dir(bin)+pathSep+os.Getenv("PATH"),
		"GETPATH="+getpath,
		"HOME="+t.TempDir(),
	)
}
