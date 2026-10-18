# $skip is the number of leading elements before git-get's own arguments (1 for git-get, 2 for git get).
$gitGetCompleter = {
    param($wordToComplete, $commandAst, $cursorPosition, $skip = 1)
    $elements = @($commandAst.CommandElements | ForEach-Object { $_.Extent.Text })

    if ($wordToComplete -like '-*') {
        '-h', '--help', '-v', '--verbose', '--version' | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
        }
        return
    }

    # Skip the command itself and the word being completed;
    # only the repository is completed, so stop once one is already given.
    $given = @($elements | Select-Object -Skip $skip)
    if ($wordToComplete -ne '') { $given = @($given | Select-Object -SkipLast 1) }
    if (@($given | Where-Object { $_ -notlike '-*' }).Count -gt 0) { return }

    git-get --complete "$wordToComplete" 2>$null | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}

Register-ArgumentCompleter -Native -CommandName git-get -ScriptBlock $gitGetCompleter

Register-ArgumentCompleter -Native -CommandName git -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    if ($commandAst.CommandElements.Count -ge 2 -and $commandAst.CommandElements[1].Extent.Text -eq 'get') {
        & $gitGetCompleter $wordToComplete $commandAst $cursorPosition 2
    }
}
