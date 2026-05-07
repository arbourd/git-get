$gitGetCompleter = {
    param($wordToComplete, $commandAst, $cursorPosition)
    if ($commandAst.CommandElements.Count -gt 2) { return }
    git-get --complete "$wordToComplete" 2>$null | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}

Register-ArgumentCompleter -Native -CommandName git-get -ScriptBlock $gitGetCompleter

Register-ArgumentCompleter -Native -CommandName git -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $c = $commandAst.CommandElements.Count
    if ($c -ge 2 -and $c -le 3 -and $commandAst.CommandElements[1].Value -eq 'get') {
        & $gitGetCompleter $wordToComplete $commandAst $cursorPosition
    }
}
