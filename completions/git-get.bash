_git_get() {
    local limit=1
    [[ "${COMP_WORDS[0]}" == "git" ]] && limit=2
    if [[ $COMP_CWORD -gt $limit ]]; then
        compopt +o default +o bashdefault 2>/dev/null
        return
    fi
    COMPREPLY=($(git-get --complete "${COMP_WORDS[COMP_CWORD]}" 2>/dev/null))
}
complete -F _git_get git-get
