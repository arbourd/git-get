# git-get

`git-get` clones repositories to your `GETPATH` in the same fashion as `go get`.

## Usage

Set a `GETPATH` or use the default of `~/src`.

```shell
export GETPATH=~/src
```

Get a repository.

```console
$ git get github.com/arbourd/git-get
~/src/github.com/arbourd/git-get

$ git get git@github.com:arbourd/git-get.git
~/src/github.com/arbourd/git-get
```

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

Install with `go get`.

```console
$ go get -u github.com/arbourd/git-get
```
