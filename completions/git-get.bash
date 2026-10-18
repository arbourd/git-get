_git_get() {
    local start=1
    [[ "${COMP_WORDS[0]}" == "git" ]] && start=2
    local cur="${COMP_WORDS[COMP_CWORD]}"
    if [[ "$cur" == -* ]]; then
        COMPREPLY=($(compgen -W "-h --help -v --verbose --version" -- "$cur"))
        return
    fi
    # Only the repository is completed; stop once one is already given.
    local i
    for ((i = start; i < COMP_CWORD; i++)); do
        if [[ "${COMP_WORDS[i]}" != -* ]]; then
            compopt +o default +o bashdefault 2>/dev/null
            return
        fi
    done
    COMPREPLY=($(git-get --complete "$cur" 2>/dev/null))
}
complete -F _git_get git-get
