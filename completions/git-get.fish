function __fish_gitget_completion
    git-get --complete (commandline -ct) 2>/dev/null
end

# True while no repository has been given yet. $argv[1] is the number of
# leading tokens that precede git-get's own arguments (1 for git-get, 2 for git get).
function __fish_gitget_needs_repo
    set -l i 0
    for t in (commandline -opc)
        set i (math $i + 1)
        test $i -le $argv[1]; and continue
        string match -q -- '-*' $t; or return 1
    end
    return 0
end

complete -c git-get -s h -l help -f -d 'Show help'
complete -c git-get -s v -l verbose -f -d 'Show git commands and output'
complete -c git-get -l version -f -d 'Show version'
complete -c git-get -f -n '__fish_gitget_needs_repo 1' -a "(__fish_gitget_completion)"
complete -c git -n '__fish_git_using_command get; and __fish_gitget_needs_repo 2' -f -a "(__fish_gitget_completion)"
complete -c git -n '__fish_git_using_command get' -s h -l help -f -d 'Show help'
complete -c git -n '__fish_git_using_command get' -s v -l verbose -f -d 'Show git commands and output'
complete -c git -n '__fish_git_using_command get' -l version -f -d 'Show version'
