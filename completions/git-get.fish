function __fish_gitget_completion
    git-get --complete (commandline -ct) 2>/dev/null
end

complete -c git-get -f -n 'test (count (commandline -opc)) -le 1' -a "(__fish_gitget_completion)"
complete -c git-get -f -n 'test (count (commandline -opc)) -gt 1'
complete -c git -n '__fish_git_using_command get; and test (count (commandline -opc)) -le 2' -f -a "(__fish_gitget_completion)"
complete -c git -n '__fish_git_using_command get; and test (count (commandline -opc)) -gt 2' -f
