package git

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DetachedMarker is used to indicate a detached HEAD state.
// This is an invalid branch name to avoid confusion with actual branch names.
const DetachedMarker = "[detached]"

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
			current.Branch = DetachedMarker
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
// It first tries to match by branch name, then by directory name (relative path from base dir),
// and finally by filesystem path (relative or absolute).
func FindWorktreeByBranchOrDir(ctx context.Context, query string) (*Worktree, error) {
	worktrees, err := ListWorktrees(ctx)
	if err != nil {
		return nil, err
	}

	// First, try to find by branch name
	for _, wt := range worktrees {
		if wt.Branch != DetachedMarker && wt.Branch == query {
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

	// Finally, if query is a valid filesystem path, resolve it and try to match
	if info, err := os.Stat(query); err == nil && info.IsDir() {
		absPath, err := filepath.Abs(query)
		if err != nil {
			return nil, nil
		}
		// Resolve symlinks (e.g., macOS /var -> /private/var)
		absPath, err = filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, nil
		}
		for _, wt := range worktrees {
			wtPath, err := filepath.EvalSymlinks(wt.Path)
			if err != nil {
				continue
			}
			if wtPath == absPath {
				return &wt, nil
			}
		}
	}

	return nil, nil
}

// WorktreeDirName returns the directory name of a worktree (relative path from base dir).
func WorktreeDirName(ctx context.Context, wt *Worktree) (string, error) {
	cfg, err := LoadConfig(ctx)
	if err != nil {
		return "", err
	}
	baseDir, err := ExpandBaseDir(ctx, cfg.BaseDir)
	if err != nil {
		return "", err
	}
	relPath, err := filepath.Rel(baseDir, wt.Path)
	if err != nil {
		return "", err
	}
	return relPath, nil
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

	if err := initBaseDir(parentDir); err != nil {
		return err
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

	// Exclude basedir from copy to prevent circular copying
	copyOpts.ExcludeDirs = append(copyOpts.ExcludeDirs, parentDir)

	// Copy files to new worktree
	if err := CopyFilesToWorktree(ctx, srcRoot, path, copyOpts); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

// AddWorktreeWithNewBranch creates a new worktree with a new branch.
// If startPoint is specified, the new branch will be created from that commit/branch.
func AddWorktreeWithNewBranch(ctx context.Context, path, branch, startPoint string, copyOpts CopyOptions) error {
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

	if err := initBaseDir(parentDir); err != nil {
		return err
	}

	// Build command arguments
	args := []string{"worktree", "add", "-b", branch, path}
	if startPoint != "" {
		args = append(args, startPoint)
	}

	cmd, err := gitCommand(ctx, args...)
	if err != nil {
		return err
	}
	// Output git messages to stderr so stdout only contains the path for shell integration
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	// Exclude basedir from copy to prevent circular copying
	copyOpts.ExcludeDirs = append(copyOpts.ExcludeDirs, parentDir)

	// Copy files to new worktree
	if err := CopyFilesToWorktree(ctx, srcRoot, path, copyOpts); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	return nil
}

// initBaseDir initializes the basedir with .gitignore and README.md files.
// It creates these files only if they don't already exist.
func initBaseDir(baseDir string) error {
	gitignorePath := filepath.Join(baseDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte("*\n"), 0600); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	readmePath := filepath.Join(baseDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		readmeContent := `# Git worktrees added by ` + "`git wt`" + `

This directory contains Git worktrees created with ` + "`git wt`" + `.

- Do NOT edit files here from parent directory contexts.
- Each subdirectory is an independent Git worktree and should be opened
  and operated on directly.
- Depending on your configuration, this directory may be placed under a Git repository.
  A ` + "`.gitignore`" + ` file ensures everything under it is ignored in that case.
`
		if err := os.WriteFile(readmePath, []byte(readmeContent), 0600); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
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
