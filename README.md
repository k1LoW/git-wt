# git-wt

A Git subcommand that makes `git worktree` simple.

## Usage

``` console
$ git wt                    # List all worktrees
$ git wt <branch>           # Switch to worktree (create if not exists)
$ git wt -d <branch>        # Delete worktree and branch (safe)
$ git wt -D <branch>        # Force delete worktree and branch
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
eval "$(git-wt --init zsh)"
```

**bash (~/.bashrc):** (experimental)

``` bash
eval "$(git-wt --init bash)"
```

**fish (~/.config/fish/config.fish):** (experimental)

``` fish
git-wt --init fish | source
```

**powershell ($PROFILE):** (experimental)

``` powershell
Invoke-Expression (git-wt --init powershell | Out-String)
```

> [!IMPORTANT]
> The shell integration creates a `git()` wrapper function to enable automatic directory switching with `git wt <branch>`. This wrapper intercepts only `git wt <branch>` commands and passes all other git commands through unchanged. If you have other tools or customizations that also wrap the `git` command, there may be conflicts.

If you want only completion without the `git()` wrapper (no automatic directory switching), use the `--no-switch-directory` option:

``` zsh
eval "$(git-wt --init zsh --no-switch-directory)"
```

## Configuration

Configuration is done via `git config`.

#### `wt.basedir`

Worktree base directory.

``` console
$ git config wt.basedir "../{gitroot}-worktrees"
```

Supported template variables:
- `{gitroot}`: repository root directory name

Default: `../{gitroot}-wt`

#### `wt.copyignored`

Copy files ignored by `.gitignore` (e.g., `.env`) to new worktrees.

``` console
$ git config wt.copyignored true
```

Default: `false`

#### `wt.copyuntracked`

Copy untracked files (not yet added to git) to new worktrees.

``` console
$ git config wt.copyuntracked true
```

Default: `false`

#### `wt.copymodified`

Copy modified files (tracked but with uncommitted changes) to new worktrees.

``` console
$ git config wt.copymodified true
```

Default: `false`
