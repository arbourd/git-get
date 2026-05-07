#compdef git-get

_git_get() {
    emulate -L zsh
    local c="${cword:-$CURRENT}"
    [[ $c -gt 2 ]] && { _ret=0; return }
    local prefix="${cur-${words[$CURRENT]}}"
    local completions
    completions=($(git-get --complete "$prefix" 2>/dev/null))
    compadd -S "" -- ${(M)completions:#*/}
    compadd -- ${completions:#*/}
    _ret=0
}
