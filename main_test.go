package main

import (
	"testing"
)

func TestClean(t *testing.T) {
	cases := map[string]struct {
		remote string
		want   string
	}{
		"git": {
			remote: "git://github.com/arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"https": {
			remote: "https://github.com/arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"filepath": {
			remote: "github.com///arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
		"https filepath": {
			remote: "https://github.com///arbourd/git-get",
			want:   "github.com/arbourd/git-get",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			str := clean(c.remote)
			if str != c.want {
				t.Fatalf("unexpected cleaned string:\n\t(GOT): %#v\n\t(WNT): %#v", str, c.want)
			}
		})
	}
}
