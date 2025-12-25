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

**bash (~/.bashrc):**

``` bash
eval "$(git-wt --init bash)"
```

**zsh (~/.zshrc):**

``` zsh
eval "$(git-wt --init zsh)"
```

**fish (~/.config/fish/config.fish):**

``` fish
git-wt --init fish | source
```

**powershell ($PROFILE):**

``` powershell
Invoke-Expression (git-wt --init powershell | Out-String)
```

## Configuration

### `wt.basedir`

Set worktree base directory via `git config`:

``` console
$ git config wt.basedir "../{gitroot}-worktrees"
```

Supported template variables:
- `{gitroot}`: repository root directory name

Default: `../{gitroot}-wt`
