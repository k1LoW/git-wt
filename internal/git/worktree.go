package git

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

	return worktrees, nil
}

// CurrentLocation returns the path that identifies the current position
// in the worktree list.
//   - bare root: returns MainRepoRoot() (bare repo directory path)
//   - worktree or normal repo: returns CurrentWorktree() (--show-toplevel)
func CurrentLocation(ctx context.Context) (string, error) {
	isBareRoot, err := IsBareRoot(ctx)
	if err != nil {
		return "", err
	}
	if isBareRoot {
		return MainRepoRoot(ctx)
	}
	return CurrentWorktree(ctx)
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
		if wt.Bare {
			continue
		}
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
			return nil, fmt.Errorf("failed to get absolute path for %q: %w", query, err)
		}
		// Resolve symlinks (e.g., macOS /var -> /private/var)
		absPath, err = filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve symlinks for %q: %w", absPath, err)
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

// addWorktreeContext holds pre-computed state shared by AddWorktree and AddWorktreeWithNewBranch.
//
// Invariant: when isBareRoot is true, srcRoot is always empty because bare
// repositories have no working tree to copy files from.
type addWorktreeContext struct {
	isBareRoot bool
	srcRoot    string // empty when isBareRoot is true
}

// prepareAdd detects the repository type (bare vs normal), determines the
// copy source worktree root, and initializes the destination parent directory.
func prepareAdd(ctx context.Context, path string) (*addWorktreeContext, error) {
	isBareRoot, err := IsBareRoot(ctx)
	if err != nil {
		return nil, err
	}

	var srcRoot string
	if !isBareRoot {
		srcRoot, err = CurrentWorktree(ctx)
		if err != nil {
			return nil, err
		}
	}

	parentDir := filepath.Dir(path)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directory: %w", err)
	}
	if err := initBaseDir(parentDir); err != nil {
		return nil, err
	}

	return &addWorktreeContext{isBareRoot: isBareRoot, srcRoot: srcRoot}, nil
}

// copyAfterAdd copies files from the current worktree to the newly created worktree.
// It is a no-op when running from a bare root (no working tree to copy from).
func copyAfterAdd(ctx context.Context, ac *addWorktreeContext, dstPath string, copyOpts CopyOptions) error {
	if ac.isBareRoot {
		return nil
	}

	// Exclude basedir from copy to prevent circular copying, but only when
	// srcRoot is outside the basedir. When srcRoot is inside the basedir
	// (e.g., bare-derived worktree at .wt/main), git ls-files already scopes
	// to srcRoot, so adding the basedir would incorrectly skip all source files.
	parentDir := filepath.Dir(dstPath)
	// Exclude parentDir when srcRoot is outside it (Rel fails or returns ".."),
	// meaning srcRoot is not itself inside the basedir. When srcRoot IS inside
	// the basedir (bare-derived worktree), we must not exclude it or git ls-files
	// would find nothing to copy.
	if rel, err := filepath.Rel(parentDir, ac.srcRoot); err != nil || strings.HasPrefix(rel, "..") {
		copyOpts.ExcludeDirs = append(copyOpts.ExcludeDirs, parentDir)
	}

	if err := CopyFilesToWorktree(ctx, ac.srcRoot, dstPath, copyOpts, os.Stderr); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}
	return nil
}

// AddWorktree creates a new worktree for the given branch.
func AddWorktree(ctx context.Context, path, branch string, copyOpts CopyOptions) error {
	ac, err := prepareAdd(ctx, path)
	if err != nil {
		return err
	}

	cmd, err := gitCommand(ctx, "worktree", "add", path, branch)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return copyAfterAdd(ctx, ac, path, copyOpts)
}

// AddWorktreeWithNewBranch creates a new worktree with a new branch.
// If startPoint is specified, the new branch will be created from that commit/branch.
func AddWorktreeWithNewBranch(ctx context.Context, path, branch, startPoint string, copyOpts CopyOptions) error {
	ac, err := prepareAdd(ctx, path)
	if err != nil {
		return err
	}

	args := []string{"worktree", "add", "-b", branch, path}
	if startPoint != "" {
		args = append(args, startPoint)
	}

	cmd, err := gitCommand(ctx, args...)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return copyAfterAdd(ctx, ac, path, copyOpts)
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

// RunRemover executes a custom remover command to remove a worktree directory.
// The worktree path is passed safely as a positional argument via sh -c.
func RunRemover(ctx context.Context, remover string, wtPath string, dir string, w io.Writer) error {
	cmd := exec.CommandContext(ctx, "sh", "-c", remover+` "$1"`, "--", wtPath)
	cmd.Dir = dir
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("remover %q failed: %w", remover, err)
	}
	return nil
}

// PruneWorktrees runs 'git worktree prune' to clean up stale worktree entries.
func PruneWorktrees(ctx context.Context) error {
	cmd, err := gitCommand(ctx, "worktree", "prune")
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
