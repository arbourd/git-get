---
name: git-get
description: Clone git repositories with `git get` instead of `git clone`, so they land at a predictable path under GETPATH (default ~/src). Use whenever the user asks to clone a repo by URL or `host/owner/repo` and gives no destination path, even if they say "git clone". For a named destination or `git clone` options, use `git clone`.
---

# git-get

`git get <repository>` clones to `$GETPATH/<host>/<owner>/<repo>` (default `~/src`) and prints that path, so every repo has one predictable location.

## Rules

1. **Give the host.** `owner/repo` alone is read as host `owner` and fails.
2. **Capture the path and check the exit status** before using it. A failed clone prints nothing, and `cd ""` succeeds in zsh and sh:

```sh
dir=$(git get -- "<repo>") && cd "$dir"
```

3. **Report errors to the user and stop.** They point at state only the user can resolve, such as a directory in the way or a bad `GETPATH`.

## Recipes

```sh
git get github.com/owner/repo          # ~/src/github.com/owner/repo
git get https://github.com/owner/repo.git
git get git@github.com:owner/repo.git  # SSH, for a private repo
git get -v github.com/owner/repo       # git commands and output on stderr
GETPATH=~/corp git get github.com/org/repo
```

One repository per call. `--` makes later args repositories, even with a leading `-`. `git get` takes no `git clone` options (`--depth`, `-b`, `--recurse-submodules`).

## Gotchas

- **Protocol:** a bare `host/owner/repo` is cloned over HTTPS unless an `url.<base>.insteadOf` rule rewrites it. A private repo needs the `git@host:owner/repo` form.
- **Already cloned:** it prints the existing path with no fetch or pull. Run `git pull` there to update.
- **Existing directory:** `<dir> exists but <dir>/.git does not` means something else is in the way.
- **Missing or private repo:** the remote is checked first, so `git repository not found at <url>` leaves no directory behind. Check the URL or auth.
- **Directory:** the `.git` suffix, port, and credentials are dropped, so `https://host:8443/o/r.git` is `host/o/r`.
- **GETPATH:** `$GETPATH`, else global `git config get.path`, else `~/src`; it must be absolute or start with `~/`. Suggest `git config --global get.path ~/dev` to the user rather than running it.

## Not installed

Clone to the same layout yourself: `git clone <url> <root>/<host>/<owner>/<repo>`, with `<root>` resolved as above (expand `~` yourself) and no `.git` suffix on `<repo>`.
