# AGENTS.md

`git-get` is a `git` extension that clones repositories into a structured directory tree by host and path (e.g. `~/src/github.com/owner/repo`), following the `go get` convention. Entry point is `main.go`; clone, path, and URL logic lives in `get/`.

## Commands

```sh
go test -short ./...                                        # unit only; plain go test ./... also hits the real network
go test -race ./...                                         # required by CI
golangci-lint run
govulncheck ./...
actionlint                                                  # lint .github/workflows (local only; not run in CI)
mandoc -T lint -W warning man/git-get.1                     # lint the man page (local only; not run in CI)
bunx --package renovate renovate-config-validator --strict  # validate .github/renovate.json
```

When running the built binary by hand, set `GETPATH` to a temp directory so clones land there instead of `~/src`.

## CLI flags

- Before `--`, help and version requests exit 0 and take precedence over other arguments.
- `--` ends option parsing.
- Invalid input prints `error: <message>` plus usage to stderr and exits 1.
- Git intercepts `git get --help` to open the man page; use `-h` for in-binary help.
- `--complete` is checked first and consumed only by the `completions/` scripts; keep it out of usage, the man page, README, and skill.
- A lone `-v` prints the version; with a repository, `-v` means `--verbose`.
- stdout carries only the resulting path (scripts do `cd $(git get …)`); verbose output goes to stderr.
- When flags, output, or behavior change, update `usage` in `main.go`, `man/git-get.1`, `README.md`, and `skills/git-get/SKILL.md` together.

## Hard constraints

- **Git operations:** use `github.com/ldez/go-git-cmd-wrapper/v2` exclusively. In non-test code, never use `exec.Command("git", ...)` or build command strings to execute. (`shellQuote` only formats the verbose log line.) The wrapper uses `CombinedOutput`, so git output is buffered, not streamed.
- **Dependencies:** zero new dependencies. The only permitted external module is `github.com/ldez/go-git-cmd-wrapper/v2`.
- **Error handling:** format URLs in errors and verbose logs through `redact()` (drops userinfo except a bare ssh username); raw user input through `get.RedactRemote()`.
- **Clone:** run the `ls-remote` validation before any directory is created.
- **Directory containment:** `Directory` guarantees its result is `filepath.IsLocal` (never escapes upward via `..`), so any GETPATH it's joined to stays contained.

## Testing

- Always call `gitConfigGlobalFixture(t)` to redirect `GIT_CONFIG_GLOBAL` to a temp file — defined once per package (`main_test.go`, `get/get_test.go`); other test files in the same package reuse it. Do not import across packages.
- Guard network tests with `if testing.Short() { t.Skip(...) }`.
- Completion tests (`completions_test.go`) spawn a real binary via the shell scripts, and skip if the shell is missing.
- Offline clone tests: map a fake `https://example.test/` to a local `git init --bare` remote with `url.<file://…>.insteadOf` in `GIT_CONFIG_GLOBAL`.
- The `clone ssh github` case needs GitHub SSH auth locally; CI rewrites SSH to HTTPS.

## Conventions

- Not-exist checks: `errors.Is(err, fs.ErrNotExist)`, per `go doc os.IsNotExist`, new code should use `errors.Is`.
- Tilde expansion: handle `~/` and bare `~` only, so `~otheruser` stays literal.
- Path portability: use `filepath.FromSlash` in tests.
