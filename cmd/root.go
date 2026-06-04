/*
Copyright © 2025 Ken'ichiro Oyama <k1lowxb@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1LoW/git-wt/internal/git"
	"github.com/k1LoW/git-wt/version"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

var (
	deleteFlag      bool
	forceDeleteFlag bool
	moveFlag        bool
	forceMoveFlag   bool
	initShell       string
	nocd            bool
	branchFlag      string
	// Config override flags.
	basedirFlag       string
	copyignoredFlag   bool
	copyuntrackedFlag bool
	copymodifiedFlag  bool
	nocopyFlag        []string
	copyFlag           []string
	symlinkFlag        []string
	hookFlag           []string
	deleteHookFlag     []string
	removerFlag        string
	allowDeleteDefault bool
	relativeFlag       bool
	jsonFlag           bool
)

var rootCmd = &cobra.Command{
	Use:   "git wt [branch|worktree] [start-point]",
	Short: "A Git subcommand that makes 'git worktree' simple",
	Long: `git-wt is a Git subcommand that makes 'git worktree' simple.

Examples:
  git wt                                         List all worktrees
  git wt <branch|worktree|path>                  Switch to worktree (create worktree/branch if needed)
  git wt <branch|worktree|path> <start-point>    Create worktree from start-point (e.g., origin/main)
  git wt -b <branch> <worktree>                  Create worktree with a different branch name
  git wt -d <branch|worktree|path>...            Delete worktree and branch (safe)
  git wt -D <branch|worktree|path>...            Force delete worktree and branch
  git wt -m [<old>] <new>                        Rename worktree directory and branch (safe)
  git wt -M [<old>] <new>                        Force rename (overwrite existing branch, allow moving dirty/locked worktrees)

Note: The default branch (e.g., main, master) is protected from accidental deletion or rename.
      Pass --allow-delete-default to override the protection in any of the cases below.
      - With worktree: -d removes the worktree but keeps the branch by default; -m/-M refuses to rename by default.
      - Without worktree: deletion is refused by default.

Shell Integration:
  Add the following to your shell config to enable worktree switching and completion:

  # bash (~/.bashrc)
  eval "$(git-wt --init bash)"

  # zsh (~/.zshrc)
  eval "$(git-wt --init zsh)"

  # fish (~/.config/fish/config.fish)
  git-wt --init fish | source

  # powershell ($PROFILE)
  Invoke-Expression (git-wt --init powershell | Out-String)

Configuration:
  Configuration is done via git config. All config options can be overridden
  with flags for a single invocation.

  wt.basedir (--basedir)
    Worktree base directory.
    Supported template variables: {gitroot} (repository root directory name)
    Default: .wt
    Example: git config wt.basedir "../{gitroot}-wt"

  wt.copyignored (--copyignored)
    Copy .gitignore'd files (e.g., .env) to new worktrees.
    Default: false

  wt.copyuntracked (--copyuntracked)
    Copy untracked files to new worktrees.
    Default: false

  wt.copymodified (--copymodified)
    Copy modified files to new worktrees.
    Default: false

  wt.copy (--copy)
    Patterns for files to always copy, even if gitignored (gitignore syntax).
    Can be specified multiple times. Useful for copying specific IDE files.
    Example: git config --add wt.copy "*.code-workspace"
             git config --add wt.copy ".vscode/"

  wt.nocopy (--nocopy)
    Patterns for files to exclude from copying (gitignore syntax).
    Can be specified multiple times.
    Note: If the same file matches both wt.copy and wt.nocopy, wt.nocopy takes precedence.
    Example: git config --add wt.nocopy "*.log"
             git config --add wt.nocopy "vendor/"

  wt.symlink (--symlink)
    Patterns for directories to symlink instead of copy (gitignore syntax).
    Matching top-level directories are symlinked to the source, sharing the
    same files. This is much faster than copying but changes affect all worktrees.
    Can be specified multiple times.
    Example: git config --add wt.symlink "node_modules/"

  wt.hook (--hook)
    Commands to run after creating a new worktree.
    Can be specified multiple times. Hooks run in the new worktree directory.
    Note: Hooks do NOT run when switching to an existing worktree.
    Example: git config --add wt.hook "npm install"
             git config --add wt.hook "go generate ./..."

  wt.deletehook (--deletehook)
    Commands to run before deleting a worktree.
    Can be specified multiple times. Hooks run in the worktree directory
    before it is removed, so you can perform cleanup (e.g., push branches).
    Note: Hooks do NOT run when deleting a branch without a worktree.
    Example: git config --add wt.deletehook "git push origin --delete $(git branch --show-current)"

  wt.remover (--remover)
    Custom command to remove the worktree directory instead of 'git worktree remove'.
    The worktree path is passed as an argument to the command.
    After the command completes, 'git worktree prune' is run automatically.
    Useful for faster deletion via trash commands (e.g., trash, trash-put).
    Default: (not set, uses 'git worktree remove')
    Example: git config wt.remover "trash-put"

  wt.nocd (--nocd)
    Do not change directory to the worktree. Only print the worktree path.
    Supported values:
      - true, all: Never cd to worktree (both new and existing)
      - create: Only prevent cd when creating new worktrees (allow cd to existing)
      - false (default): Always cd to worktree
    Note: --nocd flag always prevents cd regardless of config value.
    Using --nocd with --init disables git() wrapper (wt.nocd config does not).
    Example: git config wt.nocd create

  wt.relative (--relative)
    Append the current subdirectory path to the worktree output path.
    When running from a subdirectory, the output path will include the
    subdirectory relative to the repository root (like git diff --relative).
    Falls back to worktree root if the subdirectory does not exist in the worktree.
    Default: false
    Example: git config wt.relative true`,
	RunE:              runRoot,
	Args:              cobra.ArbitraryArgs,
	ValidArgsFunction: completeBranches,
	SilenceUsage:      true,
	Version:           version.Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s version {{.Version}}\n", version.Name))

	// Disable Cobra's default "completion" subcommand.
	// git-wt uses its own shell integration via --init flag instead.
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete worktree and branch by name or path (safe delete, only if merged)")
	rootCmd.Flags().BoolVarP(&forceDeleteFlag, "force-delete", "D", false, "Force delete worktree and branch by name or path")
	rootCmd.Flags().BoolVarP(&moveFlag, "move", "m", false, "Rename worktree directory and branch (safe rename)")
	rootCmd.Flags().BoolVarP(&forceMoveFlag, "force-move", "M", false, "Force rename worktree directory and branch (allow overwriting existing branch and moving dirty/locked worktrees)")
	rootCmd.Flags().StringVar(&initShell, "init", "", "Output shell initialization script (bash, zsh, fish, powershell)")
	rootCmd.Flags().BoolVar(&nocd, "nocd", false, "Do not change directory to the worktree (also disables git() wrapper when used with --init)")
	rootCmd.Flags().BoolVar(&nocd, "no-switch-directory", false, "")
	if err := rootCmd.Flags().MarkDeprecated("no-switch-directory", "use --nocd instead"); err != nil {
		panic(err) //nostyle:dontpanic
	}
	rootCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "Use a different branch name than the worktree directory name")
	// Config override flags.
	rootCmd.Flags().StringVar(&basedirFlag, "basedir", "", "Override wt.basedir config (worktree base directory)")
	rootCmd.Flags().BoolVar(&copyignoredFlag, "copyignored", false, "Override wt.copyignored config (copy .gitignore'd files)")
	rootCmd.Flags().BoolVar(&copyuntrackedFlag, "copyuntracked", false, "Override wt.copyuntracked config (copy untracked files)")
	rootCmd.Flags().BoolVar(&copymodifiedFlag, "copymodified", false, "Override wt.copymodified config (copy modified files)")
	rootCmd.Flags().StringArrayVar(&nocopyFlag, "nocopy", nil, "Exclude files matching pattern from copying (can be specified multiple times)")
	rootCmd.Flags().StringArrayVar(&copyFlag, "copy", nil, "Always copy files matching pattern (can be specified multiple times)")
	rootCmd.Flags().StringArrayVar(&symlinkFlag, "symlink", nil, "Symlink directories matching pattern instead of copying (can be specified multiple times)")
	rootCmd.Flags().StringArrayVar(&hookFlag, "hook", nil, "Run command after creating new worktree (can be specified multiple times)")
	rootCmd.Flags().StringArrayVar(&deleteHookFlag, "deletehook", nil, "Run command before deleting a worktree (can be specified multiple times)")
	rootCmd.Flags().StringVar(&removerFlag, "remover", "", "Custom command to remove worktree directory (e.g., trash-put)")
	rootCmd.Flags().BoolVar(&allowDeleteDefault, "allow-delete-default", false, "Allow deletion of the default branch (main, master)")
	rootCmd.Flags().BoolVar(&relativeFlag, "relative", false, "Append current subdirectory to worktree path (like git diff --relative)")
	rootCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output in JSON format")
}

func runRoot(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Handle init flag (only respects --nocd flag, not wt.nocd config)
	if initShell != "" {
		return runInit(initShell, nocd)
	}

	// Detect repo context once and thread it through context.
	// Subsequent calls to DetectRepoContext will reuse the cached value
	// instead of spawning git processes again.
	rc, err := git.DetectRepoContext(ctx)
	if err != nil {
		return err
	}
	ctx = git.WithRepoContext(ctx, rc)

	// No arguments: list worktrees
	if len(args) == 0 {
		return listWorktrees(ctx)
	}

	// Handle delete flags (multiple arguments allowed)
	if forceDeleteFlag {
		if branchFlag != "" {
			return fmt.Errorf("cannot use -b/--branch with -D/--force-delete")
		}
		if moveFlag || forceMoveFlag {
			return fmt.Errorf("cannot combine -m/-M with -d/-D")
		}
		args = uniqueArgs(args)
		return deleteWorktrees(ctx, cmd, args, true)
	}
	if deleteFlag {
		if branchFlag != "" {
			return fmt.Errorf("cannot use -b/--branch with -d/--delete")
		}
		if moveFlag || forceMoveFlag {
			return fmt.Errorf("cannot combine -m/-M with -d/-D")
		}
		args = uniqueArgs(args)
		return deleteWorktrees(ctx, cmd, args, false)
	}

	// Handle move/rename flags
	if moveFlag || forceMoveFlag {
		if branchFlag != "" {
			return fmt.Errorf("cannot use -b/--branch with -m/-M")
		}
		return moveWorktree(ctx, cmd, args, forceMoveFlag)
	}

	// For create/switch: validate argument count (like git branch)
	// git wt <branch> [<start-point>]
	if len(args) > 2 {
		return fmt.Errorf("too many arguments: expected <branch> [<start-point>], got %d arguments", len(args))
	}

	wtName := args[0]
	var startPoint string
	if len(args) == 2 {
		startPoint = args[1]
	}

	branchName := branchFlag
	if branchName == "" {
		branchName = wtName
	}

	// Default: create or switch to worktree
	return handleWorktree(ctx, cmd, wtName, branchName, startPoint)
}

// loadConfig loads config from git config and applies flag overrides.
func loadConfig(ctx context.Context, cmd *cobra.Command) (git.Config, error) {
	cfg, err := git.LoadConfig(ctx)
	if err != nil {
		return cfg, err
	}

	// Apply flag overrides
	if cmd.Flags().Changed("basedir") {
		cfg.BaseDir = basedirFlag
	}
	if cmd.Flags().Changed("copyignored") {
		cfg.CopyIgnored = copyignoredFlag
	}
	if cmd.Flags().Changed("copyuntracked") {
		cfg.CopyUntracked = copyuntrackedFlag
	}
	if cmd.Flags().Changed("copymodified") {
		cfg.CopyModified = copymodifiedFlag
	}
	if cmd.Flags().Changed("nocopy") {
		cfg.NoCopy = nocopyFlag
	}
	if cmd.Flags().Changed("copy") {
		cfg.Copy = copyFlag
	}
	if cmd.Flags().Changed("symlink") {
		cfg.Symlink = symlinkFlag
	}
	if cmd.Flags().Changed("hook") {
		cfg.Hooks = hookFlag
	}
	if cmd.Flags().Changed("deletehook") {
		cfg.DeleteHooks = deleteHookFlag
	}
	if cmd.Flags().Changed("remover") {
		cfg.Remover = removerFlag
	}
	if cmd.Flags().Changed("relative") {
		cfg.Relative = relativeFlag
	}

	return cfg, nil
}

func completeBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := cmd.Context()

	// For second argument (start-point), complete with branches including remote
	if len(args) == 1 && !deleteFlag && !forceDeleteFlag {
		return completeStartPoint(ctx)
	}

	// For delete flags, allow multiple arguments (same completion as first arg)
	// For first argument or delete mode, complete with worktrees and local branches

	// Collect unique branch names and worktree directory names
	seen := make(map[string]struct{})
	var completions []string

	// Get worktree base directory for relative path calculation
	cfg, err := loadConfig(ctx, cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	baseDir, err := git.ExpandBaseDir(ctx, cfg.BaseDir)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	// Track which names are worktrees
	worktreeNames := make(map[string]struct{})

	var commitMessages map[string]string
	if msgs, err := git.BranchCommitMessages(ctx, "refs/heads"); err == nil {
		commitMessages = msgs
	}

	// Add branches and directory names from existing worktrees
	worktrees, err := git.ListWorktrees(ctx)
	if err == nil {
		for _, wt := range worktrees {
			// Get worktree directory name (relative path from base dir)
			var wtDirName string
			if baseDir != "" {
				if relPath, err := filepath.Rel(baseDir, wt.Path); err == nil && !strings.HasPrefix(relPath, "..") {
					wtDirName = relPath
				}
			}

			// Add branch name with [branch: worktree=dir] or [worktree: branch=name] marker
			if wt.Branch != "" && wt.Branch != git.DetachedMarker {
				if _, exists := seen[wt.Branch]; !exists {
					seen[wt.Branch] = struct{}{}
					worktreeNames[wt.Branch] = struct{}{}
					wtInfo := wt.Path
					if wtDirName != "" {
						wtInfo = wtDirName
					}
					var desc string
					// If worktree dir name matches branch name, use [worktree: branch=name] format
					if wtDirName == wt.Branch {
						desc = fmt.Sprintf("[worktree: branch=%s]", wt.Branch)
						if msg := commitMessages[wt.Branch]; msg != "" {
							desc = fmt.Sprintf("[worktree: branch=%s] %s", wt.Branch, truncateString(msg, 40))
						}
					} else {
						desc = fmt.Sprintf("[branch: worktree=%s]", wtInfo)
						if msg := commitMessages[wt.Branch]; msg != "" {
							desc = fmt.Sprintf("[branch: worktree=%s] %s", wtInfo, truncateString(msg, 40))
						}
					}
					completions = append(completions, fmt.Sprintf("%s\t%s", wt.Branch, desc))
				}
			}

			// Add worktree directory name with [worktree: branch=name] or [worktree: name] marker
			if wtDirName != "" {
				if _, exists := seen[wtDirName]; !exists {
					seen[wtDirName] = struct{}{}
					worktreeNames[wtDirName] = struct{}{}
					branchInfo := wt.Branch
					if branchInfo == "" || branchInfo == git.DetachedMarker {
						branchInfo = "detached"
					}
					var desc string
					// If worktree dir name matches branch name, use simpler [worktree: name] format
					if wtDirName == branchInfo {
						desc = fmt.Sprintf("[worktree: %s]", wtDirName)
						if msg := commitMessages[wt.Branch]; msg != "" {
							desc = fmt.Sprintf("[worktree: %s] %s", wtDirName, truncateString(msg, 40))
						}
					} else {
						desc = fmt.Sprintf("[worktree: branch=%s]", branchInfo)
						if msg := commitMessages[wt.Branch]; msg != "" {
							desc = fmt.Sprintf("[worktree: branch=%s] %s", branchInfo, truncateString(msg, 40))
						}
					}
					completions = append(completions, fmt.Sprintf("%s\t%s", wtDirName, desc))
				}
			}
		}
	}

	// Add local branches (not already added as worktrees)
	branches, err := git.ListBranches(ctx)
	if err == nil {
		for _, branch := range branches {
			if _, exists := seen[branch]; !exists {
				seen[branch] = struct{}{}
				desc := "[branch]"
				if msg := commitMessages[branch]; msg != "" {
					desc = "[branch] " + truncateString(msg, 40)
				}
				completions = append(completions, fmt.Sprintf("%s\t%s", branch, desc))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// completeStartPoint returns completion for start-point (second argument).
// It includes local branches, remote branches, and tags.
func completeStartPoint(ctx context.Context) ([]string, cobra.ShellCompDirective) {
	seen := make(map[string]struct{})
	var completions []string

	var commitMessages map[string]string
	if msgs, err := git.BranchCommitMessages(ctx, "refs/heads", "refs/remotes"); err == nil {
		commitMessages = msgs
	}

	// Add local branches
	branches, err := git.ListBranches(ctx)
	if err == nil {
		for _, branch := range branches {
			if _, exists := seen[branch]; !exists {
				seen[branch] = struct{}{}
				desc := "[branch]"
				if msg := commitMessages[branch]; msg != "" {
					desc = "[branch] " + truncateString(msg, 40)
				}
				completions = append(completions, fmt.Sprintf("%s\t%s", branch, desc))
			}
		}
	}

	// Add remote branches
	remoteBranches, err := git.ListRemoteBranches(ctx)
	if err == nil {
		for _, branch := range remoteBranches {
			if _, exists := seen[branch]; !exists {
				seen[branch] = struct{}{}
				desc := "[remote]"
				if msg := commitMessages[branch]; msg != "" {
					desc = "[remote] " + truncateString(msg, 40)
				}
				completions = append(completions, fmt.Sprintf("%s\t%s", branch, desc))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// truncateString truncates a string to maxLen runes, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(r[:maxLen])
	}
	return string(r[:maxLen-3]) + "..."
}

func listWorktrees(ctx context.Context) error {
	worktrees, err := git.ListWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	currentPath, err := git.CurrentLocation(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current location: %w", err)
	}

	if jsonFlag {
		return printJSON(os.Stdout, worktrees, currentPath)
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader([]string{"", "PATH", "BRANCH", "HEAD"}),
		tablewriter.WithHeaderAlignment(tw.AlignLeft),
		tablewriter.WithHeaderPaddingPerColumn([]tw.Padding{tw.PaddingNone}),
		tablewriter.WithRowPaddingPerColumn([]tw.Padding{tw.PaddingNone}),
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{
				Left:   tw.Off,
				Right:  tw.Off,
				Top:    tw.Off,
				Bottom: tw.Off,
			},
			Settings: tw.Settings{
				Separators: tw.Separators{
					ShowHeader:     tw.Off,
					ShowFooter:     tw.Off,
					BetweenRows:    tw.Off,
					BetweenColumns: tw.Off,
				},
				Lines: tw.Lines{
					ShowTop:        tw.Off,
					ShowBottom:     tw.Off,
					ShowHeaderLine: tw.Off,
					ShowFooterLine: tw.Off,
				},
			},
		}))

	for _, wt := range worktrees {
		marker := ""
		if wt.Path == currentPath {
			marker = "*"
		}
		branch := wt.Branch
		if wt.Bare {
			branch = "(bare)"
		}
		if err := table.Append([]string{marker, wt.Path, branch, wt.Head}); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}
	return nil
}

func deleteWorktrees(ctx context.Context, cmd *cobra.Command, branches []string, force bool) error {
	// Get main repo root before any deletion (needed for running git commands after worktree removal)
	mainRoot, err := git.MainRepoRoot(ctx)
	if err != nil {
		return fmt.Errorf("failed to get main repository root: %w", err)
	}

	cfg, err := loadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if current directory is one of the worktrees being deleted
	currentWt, err := git.CurrentWorktree(ctx)
	if err != nil {
		currentWt = "" // Not in a worktree, continue
	}

	var needCdToMain bool

	for _, branch := range branches {
		// Find worktree by branch or directory name
		wt, err := git.FindWorktreeByBranchOrDir(ctx, branch)
		if err != nil {
			return fmt.Errorf("failed to find worktree: %w", err)
		}

		// Case 1: Worktree exists - remove worktree and optionally branch
		if wt != nil {
			// Check if we're deleting the current worktree
			if currentWt != "" && wt.Path == currentWt {
				needCdToMain = true
			}

			// Get worktree directory name before removal
			wtDir, err := git.WorktreeDirName(ctx, wt)
			if err != nil {
				return fmt.Errorf("failed to get worktree directory name: %w", err)
			}

			// Check branch existence and default branch status before removal
			branchExists, err := git.LocalBranchExists(ctx, wt.Branch)
			if err != nil {
				return fmt.Errorf("failed to check branch existence: %w", err)
			}

			// Check if this is the default branch (must be done before worktree removal)
			var isDefault bool
			if branchExists {
				isDefault, err = git.IsDefaultBranch(ctx, wt.Branch)
				if err != nil {
					return fmt.Errorf("failed to check default branch: %w", err)
				}
			}

			// Check for modified or untracked files (only for safe delete)
			if !force {
				modifiedFiles, err := git.ListModifiedFiles(ctx, wt.Path)
				if err != nil {
					return fmt.Errorf("failed to check for modified files: %w", err)
				}
				if len(modifiedFiles) > 0 {
					return fmt.Errorf("worktree %q has modified files, use -D to force deletion", branch)
				}

				untrackedFiles, err := git.ListUntrackedFiles(ctx, wt.Path)
				if err != nil {
					return fmt.Errorf("failed to check for untracked files: %w", err)
				}
				if len(untrackedFiles) > 0 {
					return fmt.Errorf("worktree %q has untracked files, use -D to force deletion", branch)
				}
			}

			// Run delete hooks before worktree removal (directory still exists)
			if err := git.RunHooks(ctx, cfg.DeleteHooks, wt.Path, os.Stderr); err != nil {
				return fmt.Errorf("delete hook failed for worktree %q: %w", branch, err)
			}

			// Remove worktree
			if cfg.Remover != "" {
				if err := git.RunRemover(ctx, cfg.Remover, wt.Path, mainRoot, os.Stderr); err != nil {
					return fmt.Errorf("remover failed for worktree %q: %w", branch, err)
				}
				if err := git.PruneWorktrees(ctx); err != nil {
					return fmt.Errorf("git worktree prune failed after remover for %q: %w", branch, err)
				}
			} else {
				if err := git.RemoveWorktree(ctx, wt.Path, force); err != nil {
					return fmt.Errorf("failed to remove worktree: %w", err)
				}
			}

			// Delete branch (only if it exists as a local branch)
			// Let git branch -d/-D handle the merge check
			// If we deleted the current worktree, run git from mainRoot since cwd no longer exists.
			// needCdToMain stays true for subsequent iterations as well.
			if branchExists {
				dir := ""
				if needCdToMain {
					dir = mainRoot
				}
				if isDefault && !allowDeleteDefault {
					// Default branch is protected - only delete worktree
					if wtDir == wt.Branch {
						fmt.Printf("Deleted worktree %q (branch is default, not deleted)\n", wt.Branch)
					} else {
						fmt.Printf("Deleted worktree %q (branch %q is default, not deleted)\n", wtDir, wt.Branch)
					}
				} else if err := git.DeleteBranchInDir(ctx, wt.Branch, force, dir); err != nil {
					// Treat as non-fatal since worktree removal succeeded
					if wtDir == wt.Branch {
						fmt.Printf("Deleted worktree, but failed to delete branch %q (use -D to force)\n", wt.Branch)
					} else {
						fmt.Printf("Deleted worktree %q, but failed to delete branch %q (use -D to force)\n", wtDir, wt.Branch)
					}
				} else {
					if wtDir == wt.Branch {
						fmt.Printf("Deleted worktree and branch %q\n", wt.Branch)
					} else {
						fmt.Printf("Deleted worktree %q and branch %q\n", wtDir, wt.Branch)
					}
				}
			} else {
				fmt.Printf("Deleted worktree %q (branch %q did not exist locally)\n", wtDir, wt.Branch)
			}
			continue
		}

		// Check if the target matches the bare repository entry
		isBareEntry, err := git.IsBareEntry(ctx, branch)
		if err != nil {
			return fmt.Errorf("failed to check bare entry: %w", err)
		}
		if isBareEntry {
			return fmt.Errorf("cannot delete bare repository entry %q: the bare repository root cannot be removed as a worktree", branch)
		}

		// Case 2: No worktree - try to delete branch only
		exists, err := git.LocalBranchExists(ctx, branch)
		if err != nil {
			return fmt.Errorf("failed to check branch existence: %w", err)
		}

		if !exists {
			return fmt.Errorf("no worktree or branch found for %q", branch)
		}

		// Check if this is the default branch (protected when no worktree exists)
		isDefault, err := git.IsDefaultBranch(ctx, branch)
		if err != nil {
			return fmt.Errorf("failed to check default branch: %w", err)
		}
		if isDefault && !allowDeleteDefault {
			return fmt.Errorf("cannot delete default branch %q: use --allow-delete-default to override", branch)
		}

		if err := git.DeleteBranch(ctx, branch, force); err != nil {
			return fmt.Errorf("failed to delete branch (use -D to force): %w", err)
		}
		fmt.Printf("Deleted branch %q (no worktree was associated)\n", branch)
	}

	// If we deleted the current worktree, print main repo path for shell integration to cd
	// Only output if shell integration is active (GIT_WT_SHELL_INTEGRATION=1)
	if needCdToMain && os.Getenv("GIT_WT_SHELL_INTEGRATION") == "1" {
		fmt.Println(mainRoot)
	}

	return nil
}

// moveWorktree renames a worktree's directory and its associated branch in
// a single operation. It accepts either one argument (the new name, applied
// to the current worktree) or two arguments (old, new).
func moveWorktree(ctx context.Context, cmd *cobra.Command, args []string, force bool) error {
	if len(args) == 0 || len(args) > 2 {
		return fmt.Errorf("expected 1 or 2 arguments for -m/-M: <newname> or <oldname> <newname>, got %d", len(args))
	}

	var oldQuery, newName string
	if len(args) == 1 {
		newName = args[0]
	} else {
		oldQuery = args[0]
		newName = args[1]
	}

	cfg, err := loadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	baseDir, err := git.ExpandBaseDir(ctx, cfg.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to expand basedir: %w", err)
	}
	// Resolve symlinks for reliable path comparison (macOS /var vs /private/var).
	if resolved, err := filepath.EvalSymlinks(baseDir); err == nil {
		baseDir = resolved
	}

	// Resolve source worktree.
	var src *git.Worktree
	if oldQuery == "" {
		// Single-argument form: rename the current linked worktree. The main
		// working tree cannot be renamed via this command.
		rc, err := git.DetectRepoContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to detect repository context: %w", err)
		}
		if !rc.IsLinkedWorktree() {
			return fmt.Errorf("current directory is not a linked worktree; specify the worktree to rename explicitly")
		}
		cur, err := git.CurrentWorktree(ctx)
		if err != nil {
			return fmt.Errorf("failed to get current worktree: %w", err)
		}
		src, err = git.FindWorktreeByBranchOrDir(ctx, cur)
		if err != nil {
			return fmt.Errorf("failed to find current worktree: %w", err)
		}
		if src == nil {
			return fmt.Errorf("current directory is not a linked worktree; specify the worktree to rename explicitly")
		}
	} else {
		src, err = git.FindWorktreeByBranchOrDir(ctx, oldQuery)
		if err != nil {
			return fmt.Errorf("failed to find worktree: %w", err)
		}
		if src == nil {
			// Fallback: try matching <oldQuery> as a directory name relative
			// to the already-resolved (override-aware) baseDir.
			// FindWorktreeByBranchOrDir uses internal git.LoadConfig for its
			// basedir-relative match path, so a `--basedir` flag override
			// does not reach it. Resolving to an absolute path here lets the
			// helper's path-matching strategy pick the worktree up.
			candidate := filepath.Join(baseDir, oldQuery)
			src, err = git.FindWorktreeByBranchOrDir(ctx, candidate)
			if err != nil {
				return fmt.Errorf("failed to find worktree: %w", err)
			}
			if src == nil {
				return fmt.Errorf("no worktree found for %q", oldQuery)
			}
		}
	}

	if src.Branch == "" || src.Branch == git.DetachedMarker {
		return fmt.Errorf("cannot rename worktree at %q: it has no branch (detached HEAD)", src.Path)
	}

	// Protect the default branch unless overridden.
	isDefault, err := git.IsDefaultBranch(ctx, src.Branch)
	if err != nil {
		return fmt.Errorf("failed to check default branch: %w", err)
	}
	if isDefault && !allowDeleteDefault {
		return fmt.Errorf("cannot rename default branch %q: use --allow-delete-default to override", src.Branch)
	}

	// Resolve old and new paths.
	oldPath, err := filepath.Abs(src.Path)
	if err != nil {
		return fmt.Errorf("failed to resolve old worktree path: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(oldPath); err == nil {
		oldPath = resolved
	}

	// Build newPath off the already-symlink-resolved baseDir so the prefix
	// matches oldPath (also symlink-resolved). Going through
	// git.WorktreePathFor here would re-expand cfg.BaseDir without resolving
	// symlinks, leaving newPath at e.g. /var/... while oldPath is
	// /private/var/... on macOS — which silently breaks the samePath check.
	newPath := filepath.Clean(filepath.Join(baseDir, newName))

	if oldPath == newPath && src.Branch == newName {
		return fmt.Errorf("worktree %q is already named %q", src.Branch, newName)
	}

	// Validate the new branch name BEFORE touching the filesystem. Otherwise
	// `git worktree move` can succeed and `git branch -m` then fail on an
	// invalid name (e.g., "foo..bar"), leaving a worktree directory renamed
	// while the branch is still on the old name.
	if src.Branch != newName {
		if err := git.CheckBranchNameFormat(ctx, newName); err != nil {
			return err
		}
	}

	// Pre-flight checks for target collisions. Branch collisions are skipped
	// under -M because `git branch -M` can overwrite a target branch; target
	// directory collisions are checked unconditionally because `git worktree
	// move --force` does NOT overwrite an existing destination — its --force
	// only relaxes dirty/locked-worktree checks. The check is also skipped
	// when newPath resolves to the same directory as oldPath (the
	// branch-only rename case, e.g. a worktree created via `-b` whose
	// directory is already named newName).
	samePath := oldPath == newPath
	if !samePath {
		info, err := os.Stat(newPath)
		switch {
		case err == nil:
			if info.IsDir() {
				return fmt.Errorf("target worktree directory %q already exists", newPath)
			}
			return fmt.Errorf("target path %q exists and is not a directory", newPath)
		case !os.IsNotExist(err):
			return fmt.Errorf("failed to stat target path %q: %w", newPath, err)
		}
	}
	if !force && src.Branch != newName {
		exists, err := git.LocalBranchExists(ctx, newName)
		if err != nil {
			return fmt.Errorf("failed to check branch existence: %w", err)
		}
		if exists {
			return fmt.Errorf("branch %q already exists (use -M to force)", newName)
		}
	}

	// Detect whether we are currently inside the worktree being renamed.
	curWt, _ := git.CurrentWorktree(ctx) //nostyle:handlerrors
	inside := false
	if curWt != "" {
		if resolved, err := filepath.EvalSymlinks(curWt); err == nil {
			curWt = resolved
		}
		inside = curWt == oldPath
	}

	// Move the directory first. Skip when source and target paths are
	// identical (only the branch changes).
	if !samePath {
		if err := git.MoveWorktree(ctx, oldPath, newPath, force); err != nil {
			return fmt.Errorf("failed to move worktree: %w", err)
		}
		// Clean up now-empty parent directories under basedir (e.g., the "feat/"
		// left behind when renaming "feat/foo" out of basedir/feat/foo).
		oldParent := filepath.Dir(oldPath)
		if err := git.RemoveEmptyParents(oldParent, baseDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to clean up empty parent directories: %v\n", err)
		}
	}

	// Rename the branch. Run from the new worktree path so the command works
	// even when the old cwd has just been removed.
	if src.Branch != newName {
		if err := git.RenameBranch(ctx, src.Branch, newName, force, newPath); err != nil {
			return fmt.Errorf("failed to rename branch: %w", err)
		}
	}

	// User-facing message.
	if src.Branch == newName {
		fmt.Fprintf(os.Stderr, "Moved worktree to %q\n", newPath)
	} else {
		fmt.Fprintf(os.Stderr, "Renamed worktree to %q and branch to %q\n", newPath, newName)
	}

	// Shell integration: when the current worktree was renamed, print the new
	// path so the shell wrapper can cd there. The shell wrapper inspects
	// -m/-M to keep this consistent with wt.nocd=create (existing worktree,
	// not a fresh creation).
	if inside && os.Getenv("GIT_WT_SHELL_INTEGRATION") == "1" {
		fmt.Println(newPath)
	}
	return nil
}

func handleWorktree(ctx context.Context, cmd *cobra.Command, wtName, branchName, startPoint string) error {
	// Load config with flag overrides
	cfg, err := loadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check for legacy basedir migration (only if --basedir flag is not set)
	if !cmd.Flags().Changed("basedir") {
		newBaseDir, err := checkLegacyBaseDir(ctx, cfg.BaseDir)
		if err != nil {
			return fmt.Errorf("failed to check legacy basedir: %w", err)
		}
		if newBaseDir != "" {
			cfg.BaseDir = newBaseDir
		}
	}

	// Build copy options from config
	copyOpts := git.CopyOptions{
		CopyIgnored:   cfg.CopyIgnored,
		CopyUntracked: cfg.CopyUntracked,
		CopyModified:  cfg.CopyModified,
		NoCopy:        cfg.NoCopy,
		Copy:          cfg.Copy,
		Symlink:       cfg.Symlink,
	}

	// Check if worktree already exists for this branch or directory name
	wt, err := git.FindWorktreeByBranchOrDir(ctx, branchName)
	if err != nil {
		return fmt.Errorf("failed to find worktree: %w", err)
	}
	if wt == nil && branchName != wtName {
		// Also try finding by worktree directory name
		wt, err = git.FindWorktreeByBranchOrDir(ctx, wtName)
		if err != nil {
			return fmt.Errorf("failed to find worktree: %w", err)
		}
	}

	if wt != nil {
		if startPoint != "" {
			return fmt.Errorf("worktree for branch %q already exists at %s (start-point %q is not allowed when switching to an existing worktree)", wt.Branch, wt.Path, startPoint)
		}
		// Worktree exists, switch to it
		fmt.Println(resolveRelative(ctx, wt.Path, cfg.Relative))
		return nil
	}

	// Get worktree path using the worktree name (not the branch name)
	wtPath, err := git.WorktreePathFor(ctx, cfg.BaseDir, wtName)
	if err != nil {
		return fmt.Errorf("failed to get worktree path: %w", err)
	}

	// Check if branch exists
	exists, err := git.BranchExists(ctx, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch: %w", err)
	}

	if exists {
		if startPoint != "" {
			return fmt.Errorf("branch %q already exists (start-point %q is not allowed for existing branches)", branchName, startPoint)
		}
		// Branch exists, create worktree with existing branch
		if err := git.AddWorktree(ctx, wtPath, branchName, copyOpts); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}
	} else {
		// Branch doesn't exist, create new branch and worktree
		if err := git.AddWorktreeWithNewBranch(ctx, wtPath, branchName, startPoint, copyOpts); err != nil {
			return fmt.Errorf("failed to create worktree with new branch: %w", err)
		}
	}

	// Run hooks after creating new worktree
	if err := git.RunHooks(ctx, cfg.Hooks, wtPath, os.Stderr); err != nil {
		// Print path but return error so shell integration won't cd
		fmt.Println(resolveRelative(ctx, wtPath, cfg.Relative))
		return err
	}

	// Print path to stdout
	fmt.Println(resolveRelative(ctx, wtPath, cfg.Relative))
	return nil
}

func resolveRelative(ctx context.Context, wtPath string, relative bool) string {
	if !relative {
		return wtPath
	}
	prefix, err := git.ShowPrefix(ctx)
	if err != nil || prefix == "" {
		return wtPath
	}
	resolved := filepath.Join(wtPath, prefix)
	info, err := os.Stat(resolved)
	if err != nil || !info.IsDir() {
		return wtPath
	}
	return resolved
}

// uniqueArgs removes duplicates from args while preserving order.
func uniqueArgs(args []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(args))
	for _, arg := range args {
		if _, ok := seen[arg]; !ok {
			seen[arg] = struct{}{}
			result = append(result, arg)
		}
	}
	return result
}

const legacyBaseDirPattern = "../{gitroot}-wt"

// checkLegacyBaseDir checks if the user should be notified about the basedir default change.
// Returns the basedir to use (empty string means use the current config value).
func checkLegacyBaseDir(ctx context.Context, currentBaseDir string) (string, error) {
	configured, err := git.IsBaseDirConfigured(ctx)
	if err != nil {
		return "", err
	}
	if configured {
		return "", nil
	}

	legacyPath, err := git.ExpandBaseDir(ctx, legacyBaseDirPattern)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(legacyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if !info.IsDir() {
		return "", nil
	}

	return "", fmt.Errorf("the default value for wt.basedir has changed from %q to %q and an existing directory was found at %q\nTo continue using the existing directory, run: git config wt.basedir %q\nOr delete the existing directory to use the new default", legacyBaseDirPattern, ".wt", legacyPath, legacyBaseDirPattern)
}
