package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const bashHook = `
# git-wt shell hook for bash
git() {
    if [[ "$1" == "wt" && -n "$2" && "$2" != -* ]]; then
        local result
        result=$(command git wt "$2")
        local exit_code=$?
        if [[ $exit_code -eq 0 && -d "$result" ]]; then
            cd "$result"
        else
            return $exit_code
        fi
    else
        command git "$@"
    fi
}

# git wt completion for bash
_git_wt_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    if [[ "${COMP_WORDS[1]}" == "wt" && $COMP_CWORD -eq 2 ]]; then
        local branches
        branches=$(git-wt __complete "$cur" 2>/dev/null | grep -v '^:')
        COMPREPLY=($(compgen -W "$branches" -- "$cur"))
        return 0
    fi
    return 1
}

# Hook into git completion
if type -t __git_wrap__git_main &>/dev/null; then
    _git_wt_orig_completion=$(complete -p git 2>/dev/null)
    _git_wt_wrapper() {
        if [[ "${COMP_WORDS[1]}" == "wt" ]]; then
            _git_wt_completion && return
        fi
        __git_wrap__git_main
    }
    complete -o bashdefault -o default -o nospace -F _git_wt_wrapper git
elif type -t _git &>/dev/null; then
    _git_wt_wrapper() {
        if [[ "${COMP_WORDS[1]}" == "wt" ]]; then
            _git_wt_completion && return
        fi
        _git
    }
    complete -o bashdefault -o default -o nospace -F _git_wt_wrapper git
fi
`

const zshHook = `
# git-wt shell hook for zsh
git() {
    if [[ "$1" == "wt" && -n "$2" && "$2" != -* ]]; then
        local result
        result=$(command git wt "$2")
        local exit_code=$?
        if [[ $exit_code -eq 0 && -d "$result" ]]; then
            cd "$result"
        else
            return $exit_code
        fi
    else
        command git "$@"
    fi
}

# git wt completion for zsh
_git-wt-completion() {
    if [[ "${words[2]}" == "wt" ]]; then
        local completions
        completions=(${(f)"$(git-wt __complete "${words[3]:-}" 2>/dev/null | grep -v '^:')"})
        _describe 'branch' completions
        return 0
    fi
    return 1
}

# Hook into git completion
if (( $+functions[_git] )); then
    _git-wt-orig-git() { _git "$@" }
    _git() {
        if [[ "${words[2]}" == "wt" ]]; then
            _git-wt-completion && return
        fi
        _git-wt-orig-git "$@"
    }
fi
`

const fishHook = `
# git-wt shell hook for fish
function git --wraps git
    if test "$argv[1]" = "wt" -a -n "$argv[2]" -a (string sub -l 1 -- "$argv[2]") != "-"
        set -l result (command git wt $argv[2])
        set -l exit_code $status
        if test $exit_code -eq 0 -a -d "$result"
            cd "$result"
        else
            return $exit_code
        end
    else
        command git $argv
    end
end

# git wt completion for fish
function __fish_git_wt_branches
    git-wt __complete "" 2>/dev/null | string match -rv '^:'
end

function __fish_git_wt_needs_branch
    set -l cmd (commandline -opc)
    test (count $cmd) -eq 2 -a "$cmd[2]" = "wt"
end

complete -c git -n '__fish_git_wt_needs_branch' -f -a '(__fish_git_wt_branches)'
`

const powershellHook = `
# git-wt shell hook for PowerShell
function git {
    if ($args[0] -eq "wt" -and $args[1] -and $args[1] -notlike "-*") {
        $result = & git-wt $args[1] 2>&1
        if ($LASTEXITCODE -eq 0 -and (Test-Path $result -PathType Container)) {
            Set-Location $result
        } else {
            return $LASTEXITCODE
        }
    } else {
        & git.exe @args
    }
}

# git wt completion for PowerShell
$scriptBlock = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $tokens = $commandAst.ToString() -split '\s+'
    if ($tokens.Count -ge 2 -and $tokens[1] -eq "wt") {
        $branches = git-wt __complete $wordToComplete 2>$null | Where-Object { $_ -notmatch '^:' }
        $branches | ForEach-Object {
            [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
        }
    }
}
Register-ArgumentCompleter -Native -CommandName git -ScriptBlock $scriptBlock
`

func runInit(cmd *cobra.Command, shell string) error {
	switch shell {
	case "bash":
		if err := cmd.Root().GenBashCompletion(os.Stdout); err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, bashHook)
		return nil
	case "zsh":
		if err := cmd.Root().GenZshCompletion(os.Stdout); err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, zshHook)
		return nil
	case "fish":
		if err := cmd.Root().GenFishCompletion(os.Stdout, true); err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, fishHook)
		return nil
	case "powershell":
		if err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout); err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, powershellHook)
		return nil
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}
