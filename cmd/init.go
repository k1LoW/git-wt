package cmd

import (
	"fmt"
	"io"
	"os"
)

// Bash hooks.
const bashGitWrapper = `
# Override git command to cd after 'git wt <branch>'
git() {
    if [[ "$1" == "wt" ]]; then
        shift
        local result
        result=$(command git wt "$@")
        local exit_code=$?
        if [[ $exit_code -eq 0 && -d "$result" ]]; then
            cd "$result"
        else
            echo "$result"
            return $exit_code
        fi
    else
        command git "$@"
    fi
}
`

const bashCompletion = `
# git wt <branch> completion for bash
# Function name follows Git convention: _git_<subcommand>
_git_wt() {
    local cur prev words cword
    _get_comp_words_by_ref -n =: cur prev words cword 2>/dev/null || {
        cur="${COMP_WORDS[COMP_CWORD]}"
    }
    __gitcomp_nl "$(command git-wt __complete "" 2>/dev/null | grep -v '^:')"
}
`

// Zsh hooks.
const zshGitWrapper = `
# Override git command to cd after 'git wt <branch>'
git() {
    if [[ "$1" == "wt" ]]; then
        shift
        local result
        result=$(command git wt "$@")
        local exit_code=$?
        if [[ $exit_code -eq 0 && -d "$result" ]]; then
            cd "$result"
        else
            echo "$result"
            return $exit_code
        fi
    else
        command git "$@"
    fi
}
`

const zshCompletion = `
# git wt <branch> completion for zsh
# Must be compatible with ksh emulation (Git calls: emulate ksh -c $completion_func)
_git_wt() {
    local branches
    branches="$(command git-wt __complete "" 2>/dev/null | grep -v '^:')"
    __gitcomp_nl "$branches"
}
`

// Fish hooks.
const fishGitWrapper = `
# Override git command to cd after 'git wt <branch>'
function git --wraps git
    if test "$argv[1]" = "wt"
        set -l result (command git wt $argv[2..] | string collect)
        set -l exit_code $status
        if test $exit_code -eq 0 -a -d "$result"
            cd "$result"
        else
            printf "%s\n" "$result"
            return $exit_code
        end
    else
        command git $argv
    end
end
`

const fishCompletion = `
# git wt <branch> completion for fish
function __fish_git_wt_branches
    command git-wt __complete "" 2>/dev/null | string match -rv '^:'
end

function __fish_git_wt_needs_branch
    set -l cmd (commandline -opc)
    test (count $cmd) -eq 2 -a "$cmd[2]" = "wt"
end

complete -c git -n '__fish_git_wt_needs_branch' -f -a '(__fish_git_wt_branches)'
`

// PowerShell hooks.
const powershellGitWrapper = `
# Override git command to cd after 'git wt <branch>'
function Invoke-Git {
    if ($args[0] -eq "wt") {
        $wtArgs = $args[1..($args.Length-1)]
        $result = & git.exe wt @wtArgs 2>&1
        if ($LASTEXITCODE -eq 0 -and (Test-Path $result -PathType Container)) {
            Set-Location $result
        } else {
            Write-Output $result
            return $LASTEXITCODE
        }
    } else {
        & git.exe @args
    }
}
Set-Alias -Name git -Value Invoke-Git -Option AllScope
`

const powershellCompletion = `
# git wt <branch> completion for PowerShell
$scriptBlock = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $tokens = $commandAst.ToString() -split '\s+'
    if ($tokens.Count -ge 2 -and $tokens[1] -eq "wt") {
        $branches = & git-wt.exe __complete $wordToComplete 2>$null | Where-Object { $_ -notmatch '^:' }
        $branches | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
    }
}
Register-ArgumentCompleter -Native -CommandName git -ScriptBlock $scriptBlock
`

func runInit(shell string, ignoreSwitchDirectory bool) error {
	switch shell {
	case "bash":
		fmt.Fprint(os.Stdout, "# git-wt shell hook for bash\n")
		if !ignoreSwitchDirectory {
			fmt.Fprint(os.Stdout, bashGitWrapper)
		}
		fmt.Fprint(os.Stdout, bashCompletion)
		return nil
	case "zsh":
		fmt.Fprint(os.Stdout, "# git-wt shell hook for zsh\n")
		if !ignoreSwitchDirectory {
			fmt.Fprint(os.Stdout, zshGitWrapper)
		}
		fmt.Fprint(os.Stdout, zshCompletion)
		return nil
	case "fish":
		if _, err := io.WriteString(os.Stdout, "# git-wt shell hook for fish\n"); err != nil {
			return err
		}
		if !ignoreSwitchDirectory {
			if _, err := io.WriteString(os.Stdout, fishGitWrapper); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(os.Stdout, fishCompletion); err != nil {
			return err
		}
		return nil
	case "powershell":
		fmt.Fprint(os.Stdout, "# git-wt shell hook for PowerShell\n")
		if !ignoreSwitchDirectory {
			fmt.Fprint(os.Stdout, powershellGitWrapper)
		}
		fmt.Fprint(os.Stdout, powershellCompletion)
		return nil
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}
