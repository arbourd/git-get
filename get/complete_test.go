package get

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestComplete(t *testing.T) {
	err := gitConfigGlobalFixture(t)
	if err != nil {
		t.Fatalf("unable to setup test fixture: %s", err)
	}

	getpath := t.TempDir()
	repos := []string{
		"github.com/arbourd/git-get",
		"github.com/torvalds/linux",
		"gitlab.com/gitlab-org/dev-subdepartment/ai-dev-promptcollection",
	}
	for _, r := range repos {
		dir := filepath.Join(getpath, filepath.FromSlash(r))
		if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	t.Setenv("GETPATH", getpath)

	cases := map[string]struct {
		prefix string
		want   []string
	}{
		"empty prefix returns hosts": {
			prefix: "",
			want:   []string{"github.com/", "gitlab.com/"},
		},
		"partial host": {
			prefix: "github",
			want:   []string{"github.com/"},
		},
		"full host": {
			prefix: "github.com",
			want:   []string{"github.com/"},
		},
		"host with slash": {
			prefix: "github.com/",
			want:   []string{"github.com/arbourd/", "github.com/torvalds/"},
		},
		"partial user": {
			prefix: "github.com/ar",
			want:   []string{"github.com/arbourd/"},
		},
		"full user": {
			prefix: "github.com/arbourd",
			want:   []string{"github.com/arbourd/"},
		},
		"user with slash": {
			prefix: "github.com/arbourd/",
			want:   []string{"github.com/arbourd/git-get"},
		},
		"partial repo": {
			prefix: "github.com/arbourd/git",
			want:   []string{"github.com/arbourd/git-get"},
		},
		"case insensitive": {
			prefix: "GitHub.com/Arbourd/GIT",
			want:   []string{"github.com/arbourd/git-get"},
		},
		"no match": {
			prefix: "notexist",
			want:   nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := Complete(c.prefix)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, c.want) {
				t.Fatalf("unexpected completions:\n\t(GOT): %v\n\t(WNT): %v", got, c.want)
			}
		})
	}

	t.Run("hidden directories are skipped", func(t *testing.T) {
		hiddenpath := t.TempDir()
		for _, r := range []string{".hidden/user/repo", "visible/user/repo"} {
			dir := filepath.Join(hiddenpath, filepath.FromSlash(r))
			if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
				t.Fatalf("setup: %v", err)
			}
		}
		t.Setenv("GETPATH", hiddenpath)

		got, err := Complete("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !slices.Equal(got, []string{"visible/"}) {
			t.Fatalf("unexpected completions:\n\t(GOT): %v\n\t(WNT): [visible/]", got)
		}
	})

	t.Run("non-existent GETPATH returns empty", func(t *testing.T) {
		t.Setenv("GETPATH", filepath.Join(t.TempDir(), "nonexistent"))

		got, err := Complete("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 0 {
			t.Fatalf("expected empty, got: %v", got)
		}
	})
}
