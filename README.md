# git-wt ![Coverage](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/git-wt/coverage.svg) ![Code to Test Ratio](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/git-wt/ratio.svg) ![Test Execution Time](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/git-wt/time.svg)

A Git subcommand that makes `git worktree` simple.

## Usage

``` console
$ git wt                       # List all worktrees
$ git wt <branch|worktree>     # Switch to worktree (create worktree/branch if needed)
$ git wt -d <branch|worktree>  # Delete worktree and branch (safe)
$ git wt -D <branch|worktree>  # Force delete worktree and branch
```

> [!NOTE]
> The default branch (e.g., main, master) is protected from accidental deletion.
> - If the default branch has a worktree, the worktree is deleted but the branch is preserved.
> - If the default branch has no worktree, deletion is blocked entirely.
> - Use `--allow-delete-default` to override this protection and delete the branch.

## Install

**go install:**

``` console
$ go install github.com/k1LoW/git-wt@latest
```

**homebrew tap:**

``` console
$ brew install k1LoW/tap/git-wt
```

**manually:**

Download binary from [releases page](https://github.com/k1LoW/git-wt/releases)

## Shell Integration

Add the following to your shell config to enable worktree switching and completion:

**zsh (~/.zshrc):**

``` zsh
eval "$(git wt --init zsh)"
```

**bash (~/.bashrc):** (experimental)

``` bash
eval "$(git wt --init bash)"
```

**fish (~/.config/fish/config.fish):** (experimental)

``` fish
git wt --init fish | source
```

**powershell ($PROFILE):** (experimental)

``` powershell
Invoke-Expression (git wt --init powershell | Out-String)
```

> [!IMPORTANT]
> The shell integration creates a `git()` wrapper function to enable automatic directory switching with `git wt <branch>`. This wrapper intercepts only `git wt <branch>` commands and passes all other git commands through unchanged. If you have other tools or customizations that also wrap the `git` command, there may be conflicts.

If you want only completion without the `git()` wrapper (no automatic directory switching), use the `--nocd` option:

``` zsh
eval "$(git wt --init zsh --nocd)"
```

You can also use `--nocd` with `git wt <branch>` to create/switch to a worktree without changing the current directory:

``` console
$ git wt --nocd feature-branch
/path/to/worktree/feature-branch  # prints path but stays in current directory
```

## Configuration

Configuration is done via `git config`. All config options can be overridden with flags for a single invocation.

#### `wt.basedir` / `--basedir`

Worktree base directory.

``` console
$ git config wt.basedir "../{gitroot}-worktrees"
# or override for a single invocation
$ git wt --basedir="/tmp/worktrees" feature-branch
```

Supported template variables:
- `{gitroot}`: repository root directory name

Default: `../{gitroot}-wt`

#### `wt.copyignored` / `--copyignored`

Copy files ignored by `.gitignore` (e.g., `.env`) to new worktrees.

``` console
$ git config wt.copyignored true
# or override for a single invocation
$ git wt --copyignored feature-branch
$ git wt --copyignored=false feature-branch  # explicitly disable
```

Default: `false`

#### `wt.copyuntracked` / `--copyuntracked`

Copy untracked files (not yet added to git) to new worktrees.

``` console
$ git config wt.copyuntracked true
# or override for a single invocation
$ git wt --copyuntracked feature-branch
$ git wt --copyuntracked=false feature-branch  # explicitly disable
```

Default: `false`

#### `wt.copymodified` / `--copymodified`

Copy modified files (tracked but with uncommitted changes) to new worktrees.

``` console
$ git config wt.copymodified true
# or override for a single invocation
$ git wt --copymodified feature-branch
$ git wt --copymodified=false feature-branch  # explicitly disable
```

Default: `false`

#### `wt.nocopy` / `--nocopy`

Exclude files matching patterns from copying. Uses `.gitignore` syntax.

``` console
$ git config --add wt.nocopy "*.log"
$ git config --add wt.nocopy "vendor/"
# or override for a single invocation (multiple patterns supported)
$ git wt --copyignored --nocopy "*.log" --nocopy "vendor/" feature-branch
```

Supported patterns (same as `.gitignore`):
- `*.log`: wildcard matching
- `vendor/`: directory matching
- `**/temp`: match in any directory
- `/config.local`: relative to git root

#### `wt.copy` / `--copy`

Always copy files matching patterns, even if they are gitignored. Uses `.gitignore` syntax.

``` console
$ git config --add wt.copy "*.code-workspace"
$ git config --add wt.copy ".vscode/"
# or override for a single invocation (multiple patterns supported)
$ git wt --copy "*.code-workspace" --copy ".vscode/" feature-branch
```

This is useful when you want to copy specific IDE files (like VS Code workspace files) without enabling `wt.copyignored` for all gitignored files.

> [!NOTE]
> If the same file matches both `wt.copy` and `wt.nocopy`, `wt.nocopy` takes precedence.

#### `wt.hook` / `--hook`

Commands to run after creating a new worktree. Hooks run in the new worktree directory.

``` console
$ git config --add wt.hook "npm install"
$ git config --add wt.hook "go generate ./..."
# or override for a single invocation (multiple hooks supported)
$ git wt --hook "npm install" feature-branch
```

> [!NOTE]
> - Hooks only run when **creating** a new worktree, not when switching to an existing one.
> - If a hook fails, execution stops immediately and `git wt` exits with an error (shell integration will not `cd` to the worktree).

#### `wt.nocd` / `--nocd`

Do not change directory to the worktree. Only print the worktree path.

Supported values for `wt.nocd` config:
- `true` or `all`: Never cd to worktree (both new and existing).
- `create`: Only prevent cd when creating new worktrees (allow cd to existing worktrees).
- `false` (default): Always cd to worktree.

``` console
# Prevent cd only for new worktrees (allow cd to existing)
$ git config wt.nocd create

# Never cd to any worktree
$ git config wt.nocd true

# Use --nocd flag for a single invocation (always prevents cd)
$ git wt --nocd feature-branch
```

> [!NOTE]
> - The `--nocd` flag always prevents cd regardless of config value.
> - Using `--nocd` with `--init` disables the `git()` wrapper entirely (only shell completion is output). The `wt.nocd` config does not affect `--init` output.

## Recipes

### peco

You can use [peco](https://github.com/peco/peco) for interactive worktree selection:

``` console
$ git wt $(git wt | tail -n +2 | peco | awk '{print $(NF-1)}')
```

### fzf

You can use [fzf](https://github.com/junegunn/fzf) for interactive worktree selection:

#### bash/zsh

``` console
$ cd $(git-wt | fzf | awk '{print $1}')
```

#### fish

``` console
$ cd (git-wt | fzf | awk '{print $1}')
```

### tmux

When creating a new worktree, open and switch to a new tmux window named `{repo}:{branch}`. The working directory will be the new worktree:

``` console
$ git config wt.nocd create
$ git config --add wt.hook 'tmux neww -c "$PWD" -n "$(basename -s .git `git remote get-url origin`):$(git branch --show-current)"'
```

- `wt.nocd create`: Prevents automatic directory change when creating new worktrees (tmux opens a new window instead), but still allows cd when switching to existing worktrees.
- `wt.hook 'tmux neww ...'`: Creates a new tmux window (`neww`) with `-c "$PWD"` setting the working directory to the new worktree, and `-n "..."` naming the window as `{repo}:{branch}`.
