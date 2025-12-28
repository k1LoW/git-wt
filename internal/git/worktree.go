package git

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree.
type Worktree struct {
	Path   string
	Branch string
	Head   string
	Bare   bool
}

// ListWorktrees returns a list of all worktrees.
func ListWorktrees(ctx context.Context) ([]Worktree, error) {
	cmd, err := gitCommand(ctx, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var worktrees []Worktree
	var current Worktree

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
			}
			current = Worktree{}
			continue
		}

		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			head := strings.TrimPrefix(line, "HEAD ")
			if len(head) >= 7 {
				current.Head = head[:7]
			} else {
				current.Head = head
			}
		case strings.HasPrefix(line, "branch "):
			branch := strings.TrimPrefix(line, "branch ")
			// Remove refs/heads/ prefix
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		case line == "bare":
			current.Bare = true
		case line == "detached":
			current.Branch = "(detached)"
		}
	}

	// Add the last worktree if exists
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees, nil
}

// CurrentWorktree returns the path of the current worktree.
func CurrentWorktree(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// FindWorktreeByBranch finds a worktree by branch name.
func FindWorktreeByBranch(ctx context.Context, branch string) (*Worktree, error) {
	worktrees, err := ListWorktrees(ctx)
	if err != nil {
		return nil, err
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return &wt, nil
		}
	}
	return nil, nil
}

// FindWorktreeByBranchOrDir finds a worktree by branch name or directory name.
// It first tries to match by branch name, then by directory name (relative path from base dir).
func FindWorktreeByBranchOrDir(ctx context.Context, query string) (*Worktree, error) {
	worktrees, err := ListWorktrees(ctx)
	if err != nil {
		return nil, err
	}

	// First, try to find by branch name
	for _, wt := range worktrees {
		if wt.Branch == query {
			return &wt, nil
		}
	}

	// Get worktree base directory for relative path comparison
	cfg, err := LoadConfig(ctx)
	if err != nil {
		return nil, err
	}
	baseDir, err := ExpandBaseDir(ctx, cfg.BaseDir)
	if err != nil {
		return nil, err
	}

	// Then, try to find by directory name (relative path from base dir)
	for _, wt := range worktrees {
		relPath, err := filepath.Rel(baseDir, wt.Path)
		if err != nil {
			continue
		}
		// Skip if the path is outside the base dir (starts with ..)
		if strings.HasPrefix(relPath, "..") {
			continue
		}
		if relPath == query {
			return &wt, nil
		}
	}

	return nil, nil
}

// AddWorktree creates a new worktree for the given branch.
func AddWorktree(ctx context.Context, path, branch string, copyOpts CopyOptions) error {
	// Get source root before creating worktree
	srcRoot, err := RepoRoot(ctx)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd, err := gitCommand(ctx, "worktree", "add", path, branch)
	if err != nil {
		return err
	}
	// Output git messages to stderr so stdout only contains the path for shell integration
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Copy files to new worktree
	if err := CopyFilesToWorktree(ctx, srcRoot, path, copyOpts); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

// AddWorktreeWithNewBranch creates a new worktree with a new branch.
func AddWorktreeWithNewBranch(ctx context.Context, path, branch string, copyOpts CopyOptions) error {
	// Get source root before creating worktree
	srcRoot, err := RepoRoot(ctx)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd, err := gitCommand(ctx, "worktree", "add", "-b", branch, path)
	if err != nil {
		return err
	}
	// Output git messages to stderr so stdout only contains the path for shell integration
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Copy files to new worktree
	if err := CopyFilesToWorktree(ctx, srcRoot, path, copyOpts); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

// RemoveWorktree removes a worktree.
func RemoveWorktree(ctx context.Context, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd, err := gitCommand(ctx, args...)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
