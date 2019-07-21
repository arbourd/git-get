# git-get

`git-get` clones repositories to your `GETPATH` in the same fashion as `go get`.

## Installation

Install with `brew`.

```console
$ brew install arbourd/tap/git-get
```

Install with `go get`.

```console
$ go get -u github.com/arbourd/git-get
```

## Usage

Set a `GETPATH` or use the default of `~/src`.

```console
$ export GETPATH=~/src
```

Get a repository.

```console
$ git get github.com/arbourd/git-get
~/src/github.com/arbourd/git-get
```
