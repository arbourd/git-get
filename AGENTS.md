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
completions_test.go        Bash/fish/zsh/pwsh completion tests (spawn built binary)
get/get.go                 AbsolutePath, ShortPath, ParseURL, Directory, Clone, RedactRemote
get/complete.go            Shell completion walker
completions/               Shell completion scripts (bash, fish, zsh, ps1)
man/git-get.1              Man page (update when adding flags)
.github/workflows/ci.yml   CI: test on Linux/macOS/Windows, release on tags
.goreleaser.yml            Release archives + Homebrew formula (ships completions/ and man/)
```

## Data flow

```
run() → clone() → AbsolutePath() → ParseURL() → Directory() → Clone()
run() --complete → Complete() → AbsolutePath()   (walks GETPATH for shell completion)
```

- `AbsolutePath` — resolves GETPATH: `$GETPATH` > `get.path` git config > `~/src`
- `ParseURL` — normalizes HTTPS, SSH, and SCP (`git@host:path`) forms
- `Directory` — strips `.git` suffix, returns `host/path`; guarantees the result is `filepath.IsLocal` (never escapes upward via `..`) so any GETPATH it's joined to stays contained
- `Clone` — three cases: `.git` present → return dir; dir exists without `.git` → error; dir absent → ls-remote, mkdir, clone
- `ShortPath` — inverse of `AbsolutePath`'s `~` expansion; used by `main.go` to render GETPATH in usage output
- `run` parses flags itself: `--complete` first, then `-h`/`--help`/`--version` anywhere, then options. A lone `-v` is the version; with a repository it is `--verbose`. Unknown or duplicate options and extra positionals are errors; `--` ends options.
- stdout carries only the resulting path (scripts do `cd $(git get …)`); verbose output goes to stderr via `verboseOption` (`Clone`'s `io.Writer`, nil = quiet). Git output is printed after each command finishes, not streamed: the wrapper's executor uses `CombinedOutput`, and streaming would need `exec.Command`.

## Hard constraints

**Git operations:** use `github.com/ldez/go-git-cmd-wrapper/v2` exclusively. Never use `exec.Command("git", ...)` or build command strings to execute. (`shellQuote` only formats the verbose log line.)

**Dependencies:** zero new dependencies. The only permitted external module is `github.com/ldez/go-git-cmd-wrapper/v2`.

**Error handling:** `fmt.Errorf("context: %w", err)`. Format URLs in errors and verbose logs through `redact()` (drops userinfo except a bare ssh username); raw user input through `get.RedactRemote()`.

**Not-exist checks:** `errors.Is(err, fs.ErrNotExist)` — not `os.IsNotExist` (deprecated).

**Tilde expansion:** restricted to `~/` and bare `~` only. Never expand `~otheruser`.

**Path portability:** `filepath.Clean`, `filepath.Join`, `filepath.FromSlash` (in tests).

## Testing

- Table-driven: `map[string]struct{ ... }` of named cases.
- `t.TempDir()` for filesystem isolation; `t.Setenv()` for env vars.
- Always call `gitConfigGlobalFixture(t)` to redirect `GIT_CONFIG_GLOBAL` to a temp file — defined once per package (`main_test.go`, `get/get_test.go`); other test files in the same package reuse it. Do not import across packages.
- Guard network tests with `if testing.Short() { t.Skip(...) }`.
- Completion tests (`completions_test.go`) build and spawn a real binary via the bash/fish/zsh/pwsh scripts; `completionEnv` sets `GETPATH`, `HOME`, and `PATH` explicitly and inherits `GIT_CONFIG_GLOBAL`. They skip if the shell is not installed.
- Offline clone tests: map a fake `https://example.test/` to a local `git init --bare` remote with `url.<file://…>.insteadOf` in `GIT_CONFIG_GLOBAL` (see `TestRunVerbose`).
- The `clone ssh github` case needs GitHub SSH auth locally; CI rewrites SSH to HTTPS via `git config url.https://github.com/.insteadOf ssh://git@github.com/`.

## Do not

- Skip the `ls-remote` validation in `Clone` — it must fire before any directory is created.
- Add user-facing flags without updating the usage string in `main.go` and `man/git-get.1`.
- Expose `--complete`; it is an internal flag consumed by `completions/` scripts only.
- Hardcode path separators in tests.
