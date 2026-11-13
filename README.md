# git-get

`git get` clones repositories to your `GETPATH` in the same fashion as `go get`.

## Installation

Install with `brew`.

```console
$ brew install arbourd/tap/git-get
```

Install with `go install`.

```console
$ go install github.com/arbourd/git-get@latest
```

## Usage

```console
$ git get [options] <repository>
```

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

Only the resulting path is printed to stdout, so it works with `cd`.

```console
$ cd "$(git get github.com/arbourd/git-get)"
```

### Options

| Option              | Description                  |
| ------------------- | ---------------------------- |
| `-h`, `--help`      | Show usage                   |
| `-v`, `--verbose`   | Show git commands and output |
| `--version`         | Show the version             |
| `--`                | End of options               |

### SSH

A repository without a protocol (`github.com/arbourd/git-get`) is cloned over HTTPS. To use SSH instead, redirect with [`insteadOf`](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlltbasegtinsteadOf).

```console
$ git config --global url.ssh://git@github.com/.insteadOf https://github.com/
```

### Autocompletion

Homebrew installs shell completions automatically. For other installs, copy or source the script for your shell from `completions/` in each [release](https://github.com/arbourd/git-get/releases).

```console
$ cp completions/git-get.fish ~/.config/fish/completions/   # fish
$ source completions/git-get.bash                           # bash, add to ~/.bashrc
$ cp completions/git-get.zsh "${fpath[1]}/_git_get"         # zsh
```

## Agent skill

An [agent skill](skills/git-get/SKILL.md) is available for AI coding agents.
