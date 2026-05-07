package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBashCompletion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("bash completion not supported on Windows")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
	if err := gitConfigGlobalFixture(t); err != nil {
		t.Fatalf("setup: %v", err)
	}

	bin, getpath := buildCompletionFixture(t)
	cmd := exec.Command("bash", "-c", `source completions/git-get.bash
COMP_WORDS=("git-get" "github.com/arbourd/")
COMP_CWORD=1
_git_get
printf '%s\n' "${COMPREPLY[@]}"`)
	cmd.Env = completionEnv(t, bin, getpath)

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("bash: %v", err)
	}
	if got := string(out); !strings.Contains(got, "github.com/arbourd/git-get") {
		t.Errorf("unexpected output: %q", got)
	}
}

func TestFishCompletion(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish not on PATH")
	}
	if err := gitConfigGlobalFixture(t); err != nil {
		t.Fatalf("setup: %v", err)
	}

	bin, getpath := buildCompletionFixture(t)
	cmd := exec.Command("fish", "-c", `source completions/git-get.fish
complete --do-complete "git-get github.com/arbourd/"`)
	cmd.Env = completionEnv(t, bin, getpath)

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("fish: %v", err)
	}
	if got := string(out); !strings.Contains(got, "github.com/arbourd/git-get") {
		t.Errorf("unexpected output: %q", got)
	}
}

func buildCompletionFixture(t *testing.T) (bin, getpath string) {
	t.Helper()

	bin = filepath.Join(t.TempDir(), "git-get")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	getpath = t.TempDir()
	seedRepos(t, getpath, []string{"github.com/arbourd/git-get"})
	return bin, getpath
}

func completionEnv(t *testing.T, bin, getpath string) []string {
	t.Helper()
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
