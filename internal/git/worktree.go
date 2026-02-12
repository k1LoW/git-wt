package git

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/k1LoW/exec"
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

	// Populate Head and Branch for bare entries, since
	// "git worktree list --porcelain" omits them for bare worktrees.
	for i := range worktrees {
		if worktrees[i].Bare && worktrees[i].Head == "" {
			if head, err := resolveHead(ctx, worktrees[i].Path); err == nil {
				worktrees[i].Head = head
			}
		}
		if worktrees[i].Bare && worktrees[i].Branch == "" {
			if branch, err := resolveHEADBranch(ctx, worktrees[i].Path); err == nil {
				worktrees[i].Branch = branch
			}
		}
	}

	return worktrees, nil
}

// resolveHead returns the abbreviated HEAD commit hash for the repository at dir.
func resolveHead(ctx context.Context, dir string) (string, error) {
	cmd, err := gitCommand(ctx, "-C", dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	head := strings.TrimSpace(string(out))
	if len(head) >= 7 {
		return head[:7], nil
	}
	return head, nil
}

// resolveHEADBranch returns the branch name that HEAD points to for the repository at dir.
// Returns DetachedMarker when HEAD is detached (symbolic-ref exits with code 1).
func resolveHEADBranch(ctx context.Context, dir string) (string, error) {
	cmd, err := gitCommand(ctx, "-C", dir, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		// symbolic-ref returns exit code 1 when HEAD is detached.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return DetachedMarker, nil
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// CurrentWorktree returns the path of the current worktree.
// For bare repositories, it falls back to MainRepoRoot since there is no working tree.
func CurrentWorktree(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		// --show-toplevel fails in bare repositories (exit 128).
		bare, bareErr := IsBareRepo(ctx)
		if bareErr != nil {
			return "", errors.Join(err, bareErr)
		}
		if bare {
			return MainRepoRoot(ctx)
		}
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
		if !wt.Bare && wt.Branch == branch {
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

	// First, try to find by branch name (skip bare entries — they are not switchable worktrees)
	for _, wt := range worktrees {
		if !wt.Bare && wt.Branch != DetachedMarker && wt.Branch == query {
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
		if wt.Bare {
			continue
		}
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
			if wt.Bare {
				continue
			}
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
	// Check if this is a bare repository (no working tree to copy from)
	bare, err := IsBareRepo(ctx)
	if err != nil {
		return err
	}

	// Get source root before creating worktree (only needed for non-bare repos)
	var srcRoot string
	if !bare {
		srcRoot, err = RepoRoot(ctx)
		if err != nil {
			return err
		}
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

	// Skip file copy for bare repositories (no working tree to copy from)
	if !bare {
		// Exclude basedir from copy to prevent circular copying
		copyOpts.ExcludeDirs = append(copyOpts.ExcludeDirs, parentDir)

		// Copy files to new worktree
		if err := CopyFilesToWorktree(ctx, srcRoot, path, copyOpts); err != nil {
			return fmt.Errorf("failed to copy files: %w", err)
		}
	}

	return nil
}

// AddWorktreeWithNewBranch creates a new worktree with a new branch.
// If startPoint is specified, the new branch will be created from that commit/branch.
func AddWorktreeWithNewBranch(ctx context.Context, path, branch, startPoint string, copyOpts CopyOptions) error {
	// Check if this is a bare repository (no working tree to copy from)
	bare, err := IsBareRepo(ctx)
	if err != nil {
		return err
	}

	// Get source root before creating worktree (only needed for non-bare repos)
	var srcRoot string
	if !bare {
		srcRoot, err = RepoRoot(ctx)
		if err != nil {
			return err
		}
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

	// Skip file copy for bare repositories (no working tree to copy from)
	if !bare {
		// Exclude basedir from copy to prevent circular copying
		copyOpts.ExcludeDirs = append(copyOpts.ExcludeDirs, parentDir)

		// Copy files to new worktree
		if err := CopyFilesToWorktree(ctx, srcRoot, path, copyOpts); err != nil {
			return fmt.Errorf("failed to copy files: %w", err)
		}
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
