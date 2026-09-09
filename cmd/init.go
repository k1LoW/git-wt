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
        local rename_flag=false
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
            if [[ "$arg" == "-m" || "$arg" == "-M" || "$arg" == "--move" || "$arg" == "--force-move" ]]; then
                rename_flag=true
            fi
            args+=("$arg")
        done
        local result
        result=$(GIT_WT_SHELL_INTEGRATION=1 command git wt "${args[@]}")
        local exit_code=$?
        # Get the last line for cd target
        local last_line
        last_line=$(echo "$result" | tail -n 1)
        if [[ -d "$last_line" ]]; then
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
                if [[ "$rename_flag" == "true" ]]; then
                    should_cd=true  # rename targets existing worktree at new path
                elif echo "$existing_worktrees" | grep -qxF "$last_line"; then
                    should_cd=true  # existing worktree, allow cd
                else
                    should_cd=false  # new worktree, prevent cd
                fi
            fi
            if [[ "$should_cd" == "true" ]]; then
                cd "$last_line" || return $?
            else
                echo "$last_line"
            fi
            # Pass git-wt's exit code through. A hook can fail after the worktree
            # was created, and cd should still happen in that case.
            return $exit_code
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
        local rename_flag=false
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
            if [[ "$arg" == "-m" || "$arg" == "-M" || "$arg" == "--move" || "$arg" == "--force-move" ]]; then
                rename_flag=true
            fi
            args+=("$arg")
        done
        local result
        result=$(GIT_WT_SHELL_INTEGRATION=1 command git wt "${args[@]}")
        local exit_code=$?
        # Get the last line for cd target
        local last_line
        last_line=$(echo "$result" | tail -n 1)
        if [[ -d "$last_line" ]]; then
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
                if [[ "$rename_flag" == "true" ]]; then
                    should_cd=true  # rename targets existing worktree at new path
                elif echo "$existing_worktrees" | grep -qxF "$last_line"; then
                    should_cd=true  # existing worktree, allow cd
                else
                    should_cd=false  # new worktree, prevent cd
                fi
            fi
            if [[ "$should_cd" == "true" ]]; then
                cd "$last_line" || return $?
            else
                echo "$last_line"
            fi
            # Pass git-wt's exit code through. A hook can fail after the worktree
            # was created, and cd should still happen in that case.
            return $exit_code
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
        set -l rename_flag false
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
            end
            if string match -q -- "-m" "$arg"; or string match -q -- "-M" "$arg"; or string match -q -- "--move" "$arg"; or string match -q -- "--force-move" "$arg"
                set rename_flag true
            end
        end
        set -lx GIT_WT_SHELL_INTEGRATION 1
        set -l result (command git wt $argv[2..])
        set -l exit_code $status
        # Get the last line for cd target. A failing git wt can print nothing at
        # all, so the index is guarded rather than assumed to exist.
        set -l last_line ""
        if test (count $result) -gt 0
            set last_line $result[-1]
        end
        if test -d "$last_line"
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
                if test "$rename_flag" = "true"
                    set should_cd true  # rename targets existing worktree at new path
                else if contains -- "$last_line" $existing_worktrees
                    set should_cd true  # existing worktree, allow cd
                else
                    set should_cd false  # new worktree, prevent cd
                end
            end
            if test "$should_cd" = "true"
                cd "$last_line"; or return $status
            else
                printf "%s\n" "$last_line"
            end
            # Pass git-wt's exit code through. A hook can fail after the worktree
            # was created, and cd should still happen in that case.
            return $exit_code
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
	"        $renameFlag = ($wtArgs -contains \"-m\") -or ($wtArgs -contains \"-M\") -or ($wtArgs -contains \"--move\") -or ($wtArgs -contains \"--force-move\")\n" +
	"        # Check wt.nocd config (supports: true, all, create, false)\n" +
	"        $nocdMode = & git.exe config --get wt.nocd 2>$null\n" +
	"        $existingWorktrees = @()\n" +
	"        if ($nocdMode -eq \"create\") {\n" +
	"            # Get existing worktree paths before running git wt\n" +
	"            $existingWorktrees = @(& git.exe worktree list --porcelain 2>$null | Where-Object { $_ -match '^worktree ' } | ForEach-Object { $_ -replace '^worktree ', '' })\n" +
	"        }\n" +
	"        # Function-scoped, so a caller running with Stop does not turn git-wt's\n" +
	"        # own stderr or its non-zero exit into a terminating error in here.\n" +
	"        $ErrorActionPreference = \"Continue\"\n" +
	"        $env:GIT_WT_SHELL_INTEGRATION = \"1\"\n" +
	"        # stderr is left to flow to the terminal so that only the worktree path\n" +
	"        # is captured. Merging it with 2>&1 put git's progress output and the\n" +
	"        # error line into $result, where the last line is then not the path.\n" +
	"        $result = & git.exe wt @wtArgs\n" +
	"        $exitCode = $LASTEXITCODE\n" +
	"        $env:GIT_WT_SHELL_INTEGRATION = $null\n" +
	"        # Get the last line for cd target\n" +
	"        $lines = @($result -split \"`n\" | Where-Object { $_ -ne \"\" })\n" +
	"        $lastLine = $lines[-1]\n" +
	"        if ($lastLine -and (Test-Path -LiteralPath $lastLine -PathType Container)) {\n" +
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
	"                if ($renameFlag) {\n" +
	"                    $shouldCd = $true  # rename targets existing worktree at new path\n" +
	"                } elseif ($existingWorktrees -contains $lastLine) {\n" +
	"                    $shouldCd = $true  # existing worktree, allow cd\n" +
	"                } else {\n" +
	"                    $shouldCd = $false  # new worktree, prevent cd\n" +
	"                }\n" +
	"            }\n" +
	"            if ($shouldCd) {\n" +
	"                Set-Location -LiteralPath $lastLine\n" +
	"                # Set-Location reports a failure as a non-terminating error, so a\n" +
	"                # failed cd has to replace the exit code to stay visible.\n" +
	"                if (-not $?) { $exitCode = 1 }\n" +
	"            } else {\n" +
	"                Write-Output $lastLine\n" +
	"            }\n" +
	"        } else {\n" +
	"            Write-Output $result\n" +
	"        }\n" +
	"        # Pass git-wt's exit code through. A hook can fail after the worktree was\n" +
	"        # created, and cd should still happen in that case. Returning the code would\n" +
	"        # write the number to stdout, and a function cannot set its own exit code,\n" +
	"        # so the failure status is re-established through Write-Error.\n" +
	"        $global:LASTEXITCODE = $exitCode\n" +
	"        if ($exitCode -ne 0) {\n" +
	"            Write-Error \"git wt exited with code $exitCode\" -ErrorAction Continue\n" +
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
	var header, wrapper, completion string
	switch shell {
	case "bash":
		header = "# git-wt shell hook for bash\n"
		wrapper = bashGitWrapper
		completion = bashCompletion
	case "zsh":
		header = "# git-wt shell hook for zsh\n"
		wrapper = zshGitWrapper
		completion = zshCompletion
	case "fish":
		header = "# git-wt shell hook for fish\n"
		wrapper = fishGitWrapper
		completion = fishCompletion
	case "powershell":
		header = "# git-wt shell hook for PowerShell\n"
		wrapper = powershellGitWrapper
		completion = powershellCompletion
	default:
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
	}
	if _, err := io.WriteString(os.Stdout, header); err != nil {
		return err
	}
	if !ignoreSwitchDirectory {
		if _, err := io.WriteString(os.Stdout, wrapper); err != nil {
			return err
		}
	}
	_, err := io.WriteString(os.Stdout, completion)
	return err
}
