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
	"fmt"
	"os"

	"github.com/k1LoW/git-wt/internal/git"
	"github.com/k1LoW/git-wt/version"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
)

var (
	deleteFlag      bool
	forceDeleteFlag bool
	initShell       string
)

var rootCmd = &cobra.Command{
	Use:   "git wt [branch]",
	Short: "A Git subcommand that makes 'git worktree' simple",
	Long: `git-wt is a Git subcommand that makes 'git worktree' simple.

Examples:
  git wt                List all worktrees
  git wt <branch>       Switch to worktree (create if not exists)
  git wt -d <branch>    Delete worktree and branch (safe)
  git wt -D <branch>    Force delete worktree and branch

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
  Set worktree base directory via git config:
    git config wt.basedir "../{gitroot}-worktrees"

  Supported template variables:
    {gitroot}  repository root directory name

  Default: ../{gitroot}-wt`,
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
}

func runRoot(cmd *cobra.Command, args []string) error {
	// Handle init flag
	if initShell != "" {
		return runInit(cmd, initShell)
	}

	// No arguments: list worktrees
	if len(args) == 0 {
		return listWorktrees()
	}

	branch := args[0]

	// Handle delete flags
	if forceDeleteFlag {
		return deleteWorktree(branch, true)
	}
	if deleteFlag {
		return deleteWorktree(branch, false)
	}

	// Default: create or switch to worktree
	return handleWorktree(branch)
}

func completeBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Collect unique branch names
	seen := make(map[string]struct{})
	var completions []string

	// Add branches from existing worktrees
	worktrees, err := git.ListWorktrees()
	if err == nil {
		for _, wt := range worktrees {
			if wt.Branch != "" && wt.Branch != "(detached)" {
				if _, exists := seen[wt.Branch]; !exists {
					seen[wt.Branch] = struct{}{}
					completions = append(completions, wt.Branch)
				}
			}
		}
	}

	// Add local branches
	branches, err := git.ListBranches()
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

func listWorktrees() error {
	worktrees, err := git.ListWorktrees()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	currentPath, err := git.GetCurrentWorktree()
	if err != nil {
		return fmt.Errorf("failed to get current worktree: %w", err)
	}

	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithHeader([]string{"", "PATH", "BRANCH", "HEAD"}),
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
		}),
	)

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

func deleteWorktree(branch string, force bool) error {
	// Find worktree by branch
	wt, err := git.FindWorktreeByBranch(branch)
	if err != nil {
		return fmt.Errorf("failed to find worktree: %w", err)
	}

	if wt == nil {
		return fmt.Errorf("no worktree found for branch '%s'", branch)
	}

	// Remove worktree
	if err := git.RemoveWorktree(wt.Path, force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Delete branch (only if it exists as a local branch)
	// Let git branch -d/-D handle the merge check
	exists, err := git.LocalBranchExists(branch)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if exists {
		if err := git.DeleteBranch(branch, force); err != nil {
			return fmt.Errorf("failed to delete branch (use -D to force): %w", err)
		}
		fmt.Printf("Deleted worktree and branch '%s'\n", branch)
	} else {
		fmt.Printf("Deleted worktree '%s' (branch did not exist locally)\n", branch)
	}
	return nil
}

func handleWorktree(branch string) error {
	// Check if worktree already exists for this branch
	wt, err := git.FindWorktreeByBranch(branch)
	if err != nil {
		return fmt.Errorf("failed to find worktree: %w", err)
	}

	if wt != nil {
		// Worktree exists, print path for shell integration
		fmt.Println(wt.Path)
		return nil
	}

	// Get worktree path
	wtPath, err := git.GetWorktreePath(branch)
	if err != nil {
		return fmt.Errorf("failed to get worktree path: %w", err)
	}

	// Check if branch exists
	exists, err := git.BranchExists(branch)
	if err != nil {
		return fmt.Errorf("failed to check branch: %w", err)
	}

	if exists {
		// Branch exists, create worktree with existing branch
		if err := git.AddWorktree(wtPath, branch); err != nil {
			return fmt.Errorf("failed to create worktree: %w", err)
		}
	} else {
		// Branch doesn't exist, create new branch and worktree
		if err := git.AddWorktreeWithNewBranch(wtPath, branch); err != nil {
			return fmt.Errorf("failed to create worktree with new branch: %w", err)
		}
	}

	// Print path for shell integration
	fmt.Println(wtPath)
	return nil
}
