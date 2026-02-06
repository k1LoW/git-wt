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
        local nocd_mode=""
        local nocd_flag=false
        # Check wt.nocd config (supports: true, all, create, false)
        nocd_mode="$(command git config --get wt.nocd 2>/dev/null || true)"
        local existing_worktrees=""
        if [[ "$nocd_mode" == "create" ]]; then
            # Get existing worktree paths before running git wt
            existing_worktrees=$(command git worktree list --porcelain 2>/dev/null | grep '^worktree ' | cut -d' ' -f2- || true)
        fi
        local args=()
        for arg in "$@"; do
            if [[ "$arg" == "--nocd" || "$arg" == "--no-switch-directory" ]]; then
                nocd_flag=true
            fi
            args+=("$arg")
        done
        local result
        result=$(GIT_WT_SHELL_INTEGRATION=1 command git wt "${args[@]}")
        local exit_code=$?
        # Get the last line for cd target
        local last_line
        last_line=$(echo "$result" | tail -n 1)
        if [[ $exit_code -eq 0 && -d "$last_line" ]]; then
            # Print all lines except the last (intermediate paths)
            echo "$result" | sed '$d' | while IFS= read -r line; do
                [[ -n "$line" ]] && echo "$line"
            done
            # Determine whether to cd
            local should_cd=true
            if [[ "$nocd_flag" == "true" ]]; then
                # --nocd flag always prevents cd
                should_cd=false
            elif [[ "$nocd_mode" == "true" || "$nocd_mode" == "all" ]]; then
                # wt.nocd=true/all prevents cd for all operations
                should_cd=false
            elif [[ "$nocd_mode" == "create" ]]; then
                # wt.nocd=create only prevents cd for new worktrees
                if echo "$existing_worktrees" | grep -qxF "$last_line"; then
                    should_cd=true  # existing worktree, allow cd
                else
                    should_cd=false  # new worktree, prevent cd
                fi
            fi
            if [[ "$should_cd" == "true" ]]; then
                cd "$last_line"
            else
                echo "$last_line"
            fi
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
        words=("${COMP_WORDS[@]}")
        cword=$COMP_CWORD
    }
    # Pass all arguments after 'git wt' to __complete
    local args=("${words[@]:2}")
    __gitcomp_nl "$(command git-wt __complete "${args[@]}" 2>/dev/null | grep -v '^:' | grep -v '^-[^-]' | cut -f1)"
}
`

// Zsh hooks.
const zshGitWrapper = `
# Override git command to cd after 'git wt <branch>'
git() {
    if [[ "$1" == "wt" ]]; then
        shift
        local nocd_mode=""
        local nocd_flag=false
        # Check wt.nocd config (supports: true, all, create, false)
        nocd_mode="$(command git config --get wt.nocd 2>/dev/null || true)"
        local existing_worktrees=""
        if [[ "$nocd_mode" == "create" ]]; then
            # Get existing worktree paths before running git wt
            existing_worktrees=$(command git worktree list --porcelain 2>/dev/null | grep '^worktree ' | cut -d' ' -f2- || true)
        fi
        local args=()
        for arg in "$@"; do
            if [[ "$arg" == "--nocd" || "$arg" == "--no-switch-directory" ]]; then
                nocd_flag=true
            fi
            args+=("$arg")
        done
        local result
        result=$(GIT_WT_SHELL_INTEGRATION=1 command git wt "${args[@]}")
        local exit_code=$?
        # Get the last line for cd target
        local last_line
        last_line=$(echo "$result" | tail -n 1)
        if [[ $exit_code -eq 0 && -d "$last_line" ]]; then
            # Print all lines except the last (intermediate paths)
            echo "$result" | sed '$d' | while IFS= read -r line; do
                [[ -n "$line" ]] && echo "$line"
            done
            # Determine whether to cd
            local should_cd=true
            if [[ "$nocd_flag" == "true" ]]; then
                # --nocd flag always prevents cd
                should_cd=false
            elif [[ "$nocd_mode" == "true" || "$nocd_mode" == "all" ]]; then
                # wt.nocd=true/all prevents cd for all operations
                should_cd=false
            elif [[ "$nocd_mode" == "create" ]]; then
                # wt.nocd=create only prevents cd for new worktrees
                if echo "$existing_worktrees" | grep -qxF "$last_line"; then
                    should_cd=true  # existing worktree, allow cd
                else
                    should_cd=false  # new worktree, prevent cd
                fi
            fi
            if [[ "$should_cd" == "true" ]]; then
                cd "$last_line"
            else
                echo "$last_line"
            fi
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
# git wt <branch> completion for zsh with descriptions
_git-wt() {
    local -a completions
    # Pass all previous arguments plus current word to __complete
    local args=("${words[@]:1}")
    while IFS=$'\t' read -r comp desc; do
        [[ "$comp" == :* ]] && continue
        if [[ -n "$desc" ]]; then
            completions+=("${comp}:${desc}")
        else
            completions+=("${comp}")
        fi
    done < <(command git-wt __complete "${args[@]}" 2>/dev/null)
    _describe 'git-wt' completions
}

# Hook into git completion for 'git wt'
_git-wt-wrapper() {
    if (( CURRENT == 2 )); then
        _git  # Let git handle subcommand completion
    elif [[ "${words[2]}" == "wt" ]]; then
        shift words
        (( CURRENT-- ))
        _git-wt
    else
        _git
    fi
}

# Register completions if compdef is available
if (( $+functions[compdef] )); then
    compdef _git-wt git-wt
    compdef _git-wt-wrapper git
fi
`

// Fish hooks.
const fishGitWrapper = `
# Override git command to cd after 'git wt <branch>'
function git --wraps git
    if test "$argv[1]" = "wt"
        set -l nocd_flag false
        # Check wt.nocd config (supports: true, all, create, false)
        set -l nocd_mode (command git config --get wt.nocd 2>/dev/null)
        set -l existing_worktrees
        if test "$nocd_mode" = "create"
            # Get existing worktree paths before running git wt
            set existing_worktrees (command git worktree list --porcelain 2>/dev/null | grep '^worktree ' | cut -d' ' -f2-)
        end
        for arg in $argv[2..]
            if string match -q -- "--nocd" "$arg"; or string match -q -- "--no-switch-directory" "$arg"
                set nocd_flag true
                break
            end
        end
        set -lx GIT_WT_SHELL_INTEGRATION 1
        set -l result (command git wt $argv[2..])
        set -l exit_code $status
        # Get the last line for cd target
        set -l last_line $result[-1]
        if test $exit_code -eq 0 -a -d "$last_line"
            # Print all lines except the last (intermediate paths)
            for line in $result[1..-2]
                printf "%s\n" "$line"
            end
            # Determine whether to cd
            set -l should_cd true
            if test "$nocd_flag" = "true"
                # --nocd flag always prevents cd
                set should_cd false
            else if test "$nocd_mode" = "true" -o "$nocd_mode" = "all"
                # wt.nocd=true/all prevents cd for all operations
                set should_cd false
            else if test "$nocd_mode" = "create"
                # wt.nocd=create only prevents cd for new worktrees
                if contains -- "$last_line" $existing_worktrees
                    set should_cd true  # existing worktree, allow cd
                else
                    set should_cd false  # new worktree, prevent cd
                end
            end
            if test "$should_cd" = "true"
                cd "$last_line"
            else
                printf "%s\n" "$last_line"
            end
        else
            for line in $result
                printf "%s\n" "$line"
            end
            return $exit_code
        end
    else
        command git $argv
    end
end
`

const fishCompletion = `
# git wt <branch> completion for fish
function __fish_git_wt_completions
    set -l cmd (commandline -opc)
    # Pass all arguments after 'git wt' to __complete
    set -l args $cmd[3..]
    set -l cur (commandline -ct)
    command git-wt __complete $args "$cur" 2>/dev/null | string match -rv '^:'
end

# Completions for direct git-wt invocation
function __fish_git_wt_direct_completions
    set -l cmd (commandline -opc)
    # Pass all arguments after 'git-wt' to __complete
    set -l args $cmd[2..]
    set -l cur (commandline -ct)
    command git-wt __complete $args "$cur" 2>/dev/null | string match -rv '^:'
end

function __fish_git_wt_needs_completion
    set -l cmd (commandline -opc)
    test (count $cmd) -ge 2 -a "$cmd[2]" = "wt"
end

# Completions for 'git wt'
complete -x -c git -n '__fish_git_wt_needs_completion' -a '(__fish_git_wt_completions)'

# Completions for direct 'git-wt' command (needed for fish's custom command handler)
complete -x -c git-wt -a '(__fish_git_wt_direct_completions)'
`

// PowerShell hooks.
const powershellGitWrapper = "" +
	"# Override git command to cd after 'git wt <branch>'\n" +
	"function Invoke-Git {\n" +
	"    if ($args[0] -eq \"wt\") {\n" +
	"        $wtArgs = $args[1..($args.Length-1)]\n" +
	"        $nocdFlag = ($wtArgs -contains \"--nocd\") -or ($wtArgs -contains \"--no-switch-directory\")\n" +
	"        # Check wt.nocd config (supports: true, all, create, false)\n" +
	"        $nocdMode = & git.exe config --get wt.nocd 2>$null\n" +
	"        $existingWorktrees = @()\n" +
	"        if ($nocdMode -eq \"create\") {\n" +
	"            # Get existing worktree paths before running git wt\n" +
	"            $existingWorktrees = @(& git.exe worktree list --porcelain 2>$null | Where-Object { $_ -match '^worktree ' } | ForEach-Object { $_ -replace '^worktree ', '' })\n" +
	"        }\n" +
	"        $env:GIT_WT_SHELL_INTEGRATION = \"1\"\n" +
	"        $result = & git.exe wt @wtArgs 2>&1\n" +
	"        $env:GIT_WT_SHELL_INTEGRATION = $null\n" +
	"        # Get the last line for cd target\n" +
	"        $lines = @($result -split \"`n\" | Where-Object { $_ -ne \"\" })\n" +
	"        $lastLine = $lines[-1]\n" +
	"        if ($LASTEXITCODE -eq 0 -and (Test-Path $lastLine -PathType Container)) {\n" +
	"            # Print all lines except the last (intermediate paths)\n" +
	"            if ($lines.Count -gt 1) {\n" +
	"                $lines[0..($lines.Count-2)] | ForEach-Object { Write-Output $_ }\n" +
	"            }\n" +
	"            # Determine whether to cd\n" +
	"            $shouldCd = $true\n" +
	"            if ($nocdFlag) {\n" +
	"                # --nocd flag always prevents cd\n" +
	"                $shouldCd = $false\n" +
	"            } elseif ($nocdMode -eq \"true\" -or $nocdMode -eq \"all\") {\n" +
	"                # wt.nocd=true/all prevents cd for all operations\n" +
	"                $shouldCd = $false\n" +
	"            } elseif ($nocdMode -eq \"create\") {\n" +
	"                # wt.nocd=create only prevents cd for new worktrees\n" +
	"                if ($existingWorktrees -contains $lastLine) {\n" +
	"                    $shouldCd = $true  # existing worktree, allow cd\n" +
	"                } else {\n" +
	"                    $shouldCd = $false  # new worktree, prevent cd\n" +
	"                }\n" +
	"            }\n" +
	"            if ($shouldCd) {\n" +
	"                Set-Location $lastLine\n" +
	"            } else {\n" +
	"                Write-Output $lastLine\n" +
	"            }\n" +
	"        } else {\n" +
	"            Write-Output $result\n" +
	"            return $LASTEXITCODE\n" +
	"        }\n" +
	"    } else {\n" +
	"        & git.exe @args\n" +
	"    }\n" +
	"}\n" +
	"Set-Alias -Name git -Value Invoke-Git -Option AllScope\n"

const powershellCompletion = `
# git wt <branch> completion for PowerShell
$scriptBlock = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $tokens = $commandAst.ToString() -split '\s+'
    if ($tokens.Count -ge 2 -and $tokens[1] -eq "wt") {
        # Pass all arguments after 'git wt' to __complete
        $args = @($tokens[2..($tokens.Count-1)])
        $completions = & git-wt.exe __complete @args 2>$null | Where-Object { $_ -notmatch '^:' }
        $completions | ForEach-Object {
            $parts = $_ -split [char]9, 2
            $completion = $parts[0]
            $tooltip = if ($parts.Count -gt 1) { $parts[1] } else { $parts[0] }
            [System.Management.Automation.CompletionResult]::new($completion, $completion, 'ParameterValue', $tooltip)
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
