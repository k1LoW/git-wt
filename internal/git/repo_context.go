package git

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
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
	Bare     bool // true if the main repository is bare
	Worktree bool // true if running inside a linked worktree (not the main working tree)
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
// Detection strategy:
//
// Bare detection uses `git worktree list --porcelain`. The first entry always
// represents the main repository; if it has a "bare" line, the repo is bare.
// This works regardless of whether the command is run from the bare root or
// from a linked worktree. See IsBareRepository's original doc comment for the
// detailed rationale of why rev-parse is insufficient.
//
// Worktree detection differs by repo type:
//   - Bare: `git rev-parse --show-toplevel` fails in the bare root (no working
//     tree) but succeeds inside a linked worktree. Success → Worktree=true.
//   - Non-bare: `--show-toplevel` always succeeds. If the resolved path differs
//     from worktrees[0].Path, we are in a linked worktree.
func DetectRepoContext(ctx context.Context) (RepoContext, error) {
	worktrees, err := ListWorktrees(ctx)
	if err != nil {
		return RepoContext{}, err
	}

	rc := RepoContext{}
	if len(worktrees) > 0 && worktrees[0].Bare {
		rc.Bare = true
	}

	if rc.Bare {
		// In a bare repo root, show-toplevel fails. In a linked worktree
		// created from bare, it succeeds.
		if _, err := RepoRoot(ctx); err == nil {
			rc.Worktree = true
		}
	} else {
		// For non-bare repos, compare current toplevel with main worktree path.
		toplevel, err := RepoRoot(ctx)
		if err != nil {
			return RepoContext{}, err
		}
		if len(worktrees) > 0 {
			mainPath, err := filepath.EvalSymlinks(worktrees[0].Path)
			if err != nil {
				mainPath = worktrees[0].Path
			}
			currentPath, err := filepath.EvalSymlinks(toplevel)
			if err != nil {
				currentPath = toplevel
			}
			if mainPath != currentPath {
				rc.Worktree = true
			}
		}
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
	return rc.Bare, nil
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
