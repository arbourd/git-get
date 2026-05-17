# AGENTS.md

`git-get` is a `git` extension that clones repositories into a structured directory tree by host and path (e.g. `~/src/github.com/owner/repo`), following the `go get` convention. Entry point is `main.go`; all logic lives in `get/`.

## Commands

```sh
go build ./...        # build
go test ./...         # unit + integration (hits real network)
go test -short ./...  # unit only (skips network tests)
go test -race ./...   # race detector (required by CI)
go vet ./...          # static analysis
```

Version is injected at build time: `go build -ldflags "-X main.Version=$(git rev-parse --short HEAD)" .`

## Layout

```
main.go / main_test.go     CLI entry point and tests
get/get.go                 AbsolutePath, ParseURL, Directory, Clone
get/complete.go            Shell completion walker
completions/               Shell completion scripts (bash, fish, zsh, ps1)
man/git-get.1              Man page (update when adding flags)
.github/workflows/ci.yml   CI: test on Linux/macOS/Windows, release on tags
```

## Data flow

```
run() → clone() → AbsolutePath() → ParseURL() → Directory() → Clone()
```

- `AbsolutePath` — resolves GETPATH: `$GETPATH` > `get.path` git config > `~/src`
- `ParseURL` — normalizes HTTPS, SSH, and SCP (`git@host:path`) forms
- `Directory` — strips `.git` suffix, returns `host/path`
- `Clone` — three cases: `.git` present → return dir; dir exists without `.git` → error; dir absent → ls-remote, mkdir, clone

## Hard constraints

**Git operations:** use `github.com/ldez/go-git-cmd-wrapper/v2` exclusively. Never use `exec.Command("git", ...)` or build command strings manually.

**Dependencies:** zero new dependencies. The only permitted external module is `github.com/ldez/go-git-cmd-wrapper/v2`.

**Error handling:** `fmt.Errorf("context: %w", err)`. Set `url.User = nil` before formatting any URL into an error string.

**Not-exist checks:** `errors.Is(err, fs.ErrNotExist)` — not `os.IsNotExist` (deprecated). Existing inconsistency between packages should be fixed toward `errors.Is`.

**Tilde expansion:** restricted to `~/` and bare `~` only. Never expand `~otheruser`.

**Path portability:** `filepath.Clean`, `filepath.Join`, `filepath.FromSlash` (in tests).

## Testing

- Table-driven: `map[string]struct{ ... }` of named cases.
- `t.TempDir()` for filesystem isolation; `t.Setenv()` for env vars.
- Always call `gitConfigGlobalFixture(t)` to redirect `GIT_CONFIG_GLOBAL` to a temp file — defined separately in each `_test.go` (Go package boundaries prevent sharing; do not import across packages).
- Guard network tests with `if testing.Short() { t.Skip(...) }`.
- Completion tests spawn a real binary; it must inherit `GETPATH` and `GIT_CONFIG_GLOBAL` from the test harness.

## Do not

- Use `os.IsNotExist`; use `errors.Is(err, fs.ErrNotExist)`.
- Skip the `ls-remote` validation in `Clone` — it must fire before any directory is created.
- Add user-facing flags without updating the usage string in `main.go` and `man/git-get.1`.
- Expose `--complete`; it is an internal flag consumed by `completions/` scripts only.
- Hardcode path separators in tests.
