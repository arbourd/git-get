#compdef git-get

_git_get() {
    emulate -L zsh
    # Index of git-get's first argument in $words (git get <repo> vs git-get <repo>).
    local start=2 i
    [[ "${words[1]}" == git ]] && start=3
    # Under git's zsh wrapper CURRENT is shifted down by one but $words is not; cword is the real index - 1.
    local pos=${cword:+$((cword + 1))}
    pos=${pos:-$CURRENT}
    local prefix="${cur-${words[$CURRENT]}}"
    if [[ "$prefix" == -* ]]; then
        compadd -- -h --help -v --verbose --version
        _ret=0
        return
    fi
    # Only the repository is completed; stop once one is already given.
    for (( i = start; i < pos; i++ )); do
        [[ "${words[$i]}" != -* ]] && { _ret=0; return }
    done
    local completions
    completions=($(git-get --complete "$prefix" 2>/dev/null))
    compadd -S "" -- ${(M)completions:#*/}
    compadd -- ${completions:#*/}
    _ret=0
}
