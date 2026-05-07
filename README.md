# git-get

`git get` clones repositories to your `GETPATH` in the same fashion as `go get`.

## Usage

Get a repository to the default `GETPATH` `~/src`

```console
$ git get github.com/arbourd/git-get
~/src/github.com/arbourd/git-get

$ git get https://github.com/arbourd/git-get.git
~/src/github.com/arbourd/git-get

$ git get git@github.com:arbourd/git-get.git
~/src/github.com/arbourd/git-get
```

Set a custom `GETPATH` with `git config`.

```console
$ git config --global get.path "~/dev"

$ git get github.com/arbourd/git-get
~/dev/github.com/arbourd/git-get
```

Set a custom `GETPATH` with the environmental variable `$GETPATH`. 

```console
$ export GETPATH=~/dev

$ git get github.com/arbourd/git-get
~/dev/github.com/arbourd/git-get

$ GETPATH=~/corp-dev git get github.com/arbourd/git-get
~/corp-dev/github.com/arbourd/git-get
```

The environmental variable takes precedence over the `get.path` Git config.

### Using SSH as the default

By default, when getting a repository without specifying a protocol (eg: github.com/arbourd/git-get) HTTPS will be used.

If you would prefer to use SSH or any other protocol, configure your [Git config](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf) to redirect.

```console
$ git config --global url.ssh://git@github.com/.insteadOf https://github.com/
```

## Installation

Install with `brew`.

```console
$ brew install arbourd/tap/git-get
```

Install with `go install`.

```console
$ go install github.com/arbourd/git-get@latest
```

### Autocompletion

Homebrew installs shell completions automatically. For other installs, completion scripts are available in the `completions/` directory of each [release](https://github.com/arbourd/git-get/releases).

**Fish**

```console
$ cp completions/git-get.fish ~/.config/fish/completions/git-get.fish
```

**Bash** — add to `~/.bashrc` to persist.

```console
$ source completions/git-get.bash
```

**Zsh** — copy to a directory in `$fpath`.

```console
$ cp completions/git-get.zsh "${fpath[1]}/_git_get"
```

**PowerShell** — dot-source in `$PROFILE` to persist.

```console
> . completions\git-get.ps1
```
