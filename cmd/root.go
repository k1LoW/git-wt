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
	deleteFlag        bool
	forceDeleteFlag   bool
	initShell         string
	noSwitchDirectory bool
	// Config override flags.
	basedirFlag       string
	copyignoredFlag   bool
	copyuntrackedFlag bool
	copymodifiedFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "git wt [branch|worktree]",
	Short: "A Git subcommand that makes 'git worktree' simple",
	Long: `git-wt is a Git subcommand that makes 'git worktree' simple.

Examples:
  git wt                List all worktrees
  git wt <branch|worktree>  Switch to worktree (create worktree/branch if needed)
  git wt -d <branch|worktree>  Delete worktree and branch (safe)
  git wt -D <branch|worktree>  Force delete worktree and branch

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
    Default: ../{gitroot}-wt
    Example: git config wt.basedir "../{gitroot}-worktrees"

  wt.copyignored (--copyignored)
    Copy .gitignore'd files (e.g., .env) to new worktrees.
    Default: false

  wt.copyuntracked (--copyuntracked)
    Copy untracked files to new worktrees.
    Default: false

  wt.copymodified (--copymodified)
    Copy modified files to new worktrees.
    Default: false`,
	RunE:              runRoot,
	Args:              cobra.MaximumNArgs(1),
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
	rootCmd.Flags().BoolVarP(&deleteFlag, "delete", "d", false, "Delete worktree and branch (safe delete, only if merged)")
	rootCmd.Flags().BoolVarP(&forceDeleteFlag, "force-delete", "D", false, "Force delete worktree and branch")
	rootCmd.Flags().StringVar(&initShell, "init", "", "Output shell initialization script (bash, zsh, fish, powershell)")
	rootCmd.Flags().BoolVar(&noSwitchDirectory, "no-switch-directory", false, "Do not add git() wrapper for automatic directory switching (use with --init)")
	// Config override flags.
	rootCmd.Flags().StringVar(&basedirFlag, "basedir", "", "Override wt.basedir config (worktree base directory)")
	rootCmd.Flags().BoolVar(&copyignoredFlag, "copyignored", false, "Override wt.copyignored config (copy .gitignore'd files)")
	rootCmd.Flags().BoolVar(&copyuntrackedFlag, "copyuntracked", false, "Override wt.copyuntracked config (copy untracked files)")
	rootCmd.Flags().BoolVar(&copymodifiedFlag, "copymodified", false, "Override wt.copymodified config (copy modified files)")
}

func runRoot(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Handle init flag
	if initShell != "" {
		return runInit(initShell, noSwitchDirectory)
	}

	// No arguments: list worktrees
	if len(args) == 0 {
		return listWorktrees(ctx)
	}

	branch := args[0]

	// Handle delete flags
	if forceDeleteFlag {
		return deleteWorktree(ctx, branch, true)
	}
	if deleteFlag {
		return deleteWorktree(ctx, branch, false)
	}

	// Default: create or switch to worktree
	return handleWorktree(ctx, cmd, branch)
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

	return cfg, nil
}

func completeBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ctx := cmd.Context()

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

	// Add branches and directory names from existing worktrees
	worktrees, err := git.ListWorktrees(ctx)
	if err == nil {
		for _, wt := range worktrees {
			// Add branch name
			if wt.Branch != "" && wt.Branch != "(detached)" {
				if _, exists := seen[wt.Branch]; !exists {
					seen[wt.Branch] = struct{}{}
					completions = append(completions, wt.Branch)
				}
			}

			// Add worktree directory name (relative path from base dir)
			if baseDir != "" {
				relPath, err := filepath.Rel(baseDir, wt.Path)
				if err == nil && !strings.HasPrefix(relPath, "..") {
					if _, exists := seen[relPath]; !exists {
						seen[relPath] = struct{}{}
						completions = append(completions, relPath)
					}
				}
			}
		}
	}

	// Add local branches
	branches, err := git.ListBranches(ctx)
	if err == nil {
		for _, branch := range branches {
			if _, exists := seen[branch]; !exists {
				seen[branch] = struct{}{}
				completions = append(completions, branch)
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

func listWorktrees(ctx context.Context) error {
	worktrees, err := git.ListWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	currentPath, err := git.CurrentWorktree(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current worktree: %w", err)
	}

	table := tablewriter.NewTable(os.Stdout, tablewriter.WithHeader([]string{"", "PATH", "BRANCH", "HEAD"}), tablewriter.WithRendition(tw.Rendition{
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
		if err := table.Append([]string{marker, wt.Path, wt.Branch, wt.Head}); err != nil {
			return fmt.Errorf("failed to append row: %w", err)
		}
	}

	if err := table.Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}
	return nil
}

func deleteWorktree(ctx context.Context, branch string, force bool) error {
	// Find worktree by branch or directory name
	wt, err := git.FindWorktreeByBranchOrDir(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to find worktree: %w", err)
	}

	if wt == nil {
		return fmt.Errorf("no worktree found for branch or directory %q", branch)
	}

	// Remove worktree
	if err := git.RemoveWorktree(ctx, wt.Path, force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete branch (only if it exists as a local branch)
	// Let git branch -d/-D handle the merge check
	exists, err := git.LocalBranchExists(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if exists {
		if err := git.DeleteBranch(ctx, branch, force); err != nil {
			return fmt.Errorf("failed to delete branch (use -D to force): %w", err)
		}
		fmt.Printf("Deleted worktree and branch %q\n", branch)
	} else {
		fmt.Printf("Deleted worktree %q (branch did not exist locally)\n", branch)
	}
	return nil
}

func handleWorktree(ctx context.Context, cmd *cobra.Command, branch string) error {
	// Load config with flag overrides
	cfg, err := loadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if worktree already exists for this branch or directory name
	wt, err := git.FindWorktreeByBranchOrDir(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to find worktree: %w", err)
	}

	if wt != nil {
		// Worktree exists, print path for shell integration
		fmt.Println(wt.Path)
		return nil
	}

	// Get worktree path
	wtPath, err := git.WorktreePathFor(ctx, cfg.BaseDir, branch)
	if err != nil {
		return fmt.Errorf("failed to get worktree path: %w", err)
	}

	// Build copy options from config
	copyOpts := git.CopyOptions{
		CopyIgnored:   cfg.CopyIgnored,
		CopyUntracked: cfg.CopyUntracked,
		CopyModified:  cfg.CopyModified,
	}

	// Check if branch exists
	exists, err := git.BranchExists(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to check branch: %w", err)
	}

	if exists {
		// Branch exists, create worktree with existing branch
		if err := git.AddWorktree(ctx, wtPath, branch, copyOpts); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}
	} else {
		// Branch doesn't exist, create new branch and worktree
		if err := git.AddWorktreeWithNewBranch(ctx, wtPath, branch, copyOpts); err != nil {
			return fmt.Errorf("failed to create worktree with new branch: %w", err)
		}
	}

	// Print path for shell integration
	fmt.Println(wtPath)
	return nil
}
