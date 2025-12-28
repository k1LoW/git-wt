# git-wt ![Coverage](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/git-wt/coverage.svg) ![Code to Test Ratio](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/git-wt/ratio.svg) ![Test Execution Time](https://raw.githubusercontent.com/k1LoW/octocovs/main/badges/k1LoW/git-wt/time.svg)

A Git subcommand that makes `git worktree` simple.

## Usage

``` console
$ git wt                    # List all worktrees
$ git wt <branch|worktree>  # Switch to worktree (create worktree/branch if needed)
$ git wt -d <branch|worktree>  # Delete worktree and branch (safe)
$ git wt -D <branch|worktree>  # Force delete worktree and branch
```

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

If you want only completion without the `git()` wrapper (no automatic directory switching), use the `--no-switch-directory` option:

``` zsh
eval "$(git wt --init zsh --no-switch-directory)"
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
