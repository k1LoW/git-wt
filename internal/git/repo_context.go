package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RepoContext describes the type and location within a git repository.
//
// The four possible states are:
//
//	{Bare: false, Worktree: false} — main working tree of a normal repository
//	{Bare: false, Worktree: true}  — linked worktree of a normal repository
//	{Bare: true,  Worktree: false} — bare repository root (no working tree)
//	{Bare: true,  Worktree: true}  — linked worktree created from a bare repository
type RepoContext struct {
	bare     bool   // true if the main repository is bare
	worktree bool   // true if running inside a linked worktree (not the main working tree)
	dir      string // working directory at detection time (used for cache invalidation)
}

type repoContextKey struct{}

// WithRepoContext stores a RepoContext in the given context for later retrieval.
func WithRepoContext(ctx context.Context, rc RepoContext) context.Context {
	return context.WithValue(ctx, repoContextKey{}, &rc)
}

// RepoContextFrom retrieves a cached RepoContext from ctx.
// It returns nil if no value is stored or if the current working directory
// differs from the directory recorded at detection time.
func RepoContextFrom(ctx context.Context) *RepoContext {
	rc, ok := ctx.Value(repoContextKey{}).(*RepoContext)
	if !ok {
		return nil
	}
	if cwd, err := os.Getwd(); err != nil || cwd != rc.dir {
		return nil
	}
	return rc
}

// ErrBareRepository is a sentinel error returned when a bare repository is
// detected but the requested operation does not support bare repositories.
//
// Bare repositories lack a working tree, so many git-wt operations
// (list, add/switch, delete) cannot function correctly in them.
// Support for bare repositories is tracked in the linked issue.
var ErrBareRepository = errors.New(
	"bare repositories are not currently supported by git-wt\n" +
		"For more information, see: https://github.com/k1LoW/git-wt/issues/130",
)

// DetectRepoContext detects whether the current repository is bare and whether
// the current working directory is inside a linked worktree.
//
// Detection uses `git rev-parse --is-bare-repository --git-dir --git-common-dir`
// in a single process invocation:
//
//   - Bare: detected by combining two signals with OR:
//     1. --is-bare-repository flag (reliable at bare root, but returns false
//     in bare-derived worktrees)
//     2. filepath.Base(gitCommonDir) != ".git" (catches bare repos named
//     without .git suffix, but misses bare repos where git-common-dir
//     is a .git subdirectory)
//   - Worktree: gitDir != gitCommonDir
//     In the main working tree (or bare root), both are equal. In a linked
//     worktree, gitDir points to a worktrees/X subdirectory.
func DetectRepoContext(ctx context.Context) (RepoContext, error) {
	if cached := RepoContextFrom(ctx); cached != nil {
		return *cached, nil
	}

	cmd, err := gitCommand(ctx, "rev-parse", "--is-bare-repository", "--path-format=absolute", "--git-dir", "--git-common-dir")
	if err != nil {
		return RepoContext{}, err
	}
	out, err := cmd.Output()
	if err != nil {
		return RepoContext{}, err
	}
	lines := strings.SplitN(strings.TrimSpace(string(out)), "\n", 3)
	if len(lines) != 3 {
		return RepoContext{}, fmt.Errorf("unexpected output from git rev-parse: %q", string(out))
	}

	isBareFlag := lines[0] == "true"
	gitDir := lines[1]
	gitCommonDir := lines[2]

	rc := RepoContext{
		bare:     isBareFlag || filepath.Base(gitCommonDir) != ".git",
		worktree: gitDir != gitCommonDir,
	}

	if cwd, err := os.Getwd(); err == nil {
		rc.dir = cwd
	}

	return rc, nil
}

// IsBareRepository reports whether the main repository is bare.
// It is a convenience wrapper around DetectRepoContext.
func IsBareRepository(ctx context.Context) (bool, error) {
	rc, err := DetectRepoContext(ctx)
	if err != nil {
		return false, err
	}
	return rc.bare, nil
}

// AssertNotBareRepository returns ErrBareRepository if the current repository
// is bare. This is used as a guard at the beginning of operations that do not
// support bare repositories.
//
// When bare repository support is added for a specific operation, its guard
// call can simply be removed. This design allows staged (per-operation)
// enablement of bare repository support.
func AssertNotBareRepository(ctx context.Context) error {
	isBare, err := IsBareRepository(ctx)
	if err != nil {
		return fmt.Errorf("failed to check repository type: %w", err)
	}
	if isBare {
		return ErrBareRepository
	}
	return nil
}

// IsNormalMain reports whether the current directory is the main working tree
// of a normal (non-bare) repository.
func IsNormalMain(ctx context.Context) (bool, error) {
	rc, err := DetectRepoContext(ctx)
	if err != nil {
		return false, err
	}
	return !rc.bare && !rc.worktree, nil
}

// IsNormalWorktree reports whether the current directory is a linked worktree
// of a normal (non-bare) repository.
func IsNormalWorktree(ctx context.Context) (bool, error) {
	rc, err := DetectRepoContext(ctx)
	if err != nil {
		return false, err
	}
	return !rc.bare && rc.worktree, nil
}

// IsBareRoot reports whether the current directory is a bare repository root
// (no working tree).
func IsBareRoot(ctx context.Context) (bool, error) {
	rc, err := DetectRepoContext(ctx)
	if err != nil {
		return false, err
	}
	return rc.bare && !rc.worktree, nil
}

// IsBareWorktree reports whether the current directory is a linked worktree
// created from a bare repository.
func IsBareWorktree(ctx context.Context) (bool, error) {
	rc, err := DetectRepoContext(ctx)
	if err != nil {
		return false, err
	}
	return rc.bare && rc.worktree, nil
}

// gitDirs returns the git-dir and git-common-dir for the current repository.
// Both paths are returned as absolute paths resolved by git.
// git-dir points to the .git directory (or worktrees/X subdirectory for linked worktrees).
// git-common-dir points to the shared .git directory of the main repository.
func gitDirs(ctx context.Context) (gitDir, gitCommonDir string, err error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--path-format=absolute", "--git-dir", "--git-common-dir")
	if err != nil {
		return "", "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", "", err
	}
	lines := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)
	if len(lines) != 2 {
		return "", "", fmt.Errorf("unexpected output from git rev-parse: %q", string(out))
	}
	return lines[0], lines[1], nil
}

// ShowPrefix returns the path prefix of the current directory relative to the repository root.
// It runs "git rev-parse --show-prefix" and strips the trailing slash.
// Returns an empty string when at the repository root.
func ShowPrefix(ctx context.Context) (string, error) {
	cmd, err := gitCommand(ctx, "rev-parse", "--show-prefix")
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSpace(string(out)), "/"), nil
}

// MainRepoRoot returns the root directory of the main git repository.
// Unlike RepoRoot, this returns the main repository root even when called from a worktree.
//
// For normal repositories, git-common-dir is ".git" inside the repo root,
// so the parent directory is returned. For bare repositories, git-common-dir
// IS the repository directory itself, so it is returned directly.
func MainRepoRoot(ctx context.Context) (string, error) {
	_, gitCommonDir, err := gitDirs(ctx)
	if err != nil {
		return "", err
	}
	if filepath.Base(gitCommonDir) == ".git" {
		return filepath.Dir(gitCommonDir), nil
	}
	return gitCommonDir, nil
}

// RepoName returns the name of the current git repository (directory name).
func RepoName(ctx context.Context) (string, error) {
	root, err := MainRepoRoot(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Base(root), nil
}
