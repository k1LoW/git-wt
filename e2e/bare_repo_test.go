// bare_repo_test.go contains tests for bare repository support:
//   - TestE2E_BareRepo_ListWorktrees: listing worktrees in a bare repo
//   - TestE2E_BareRepo_CreateWorktree: creating worktrees from a bare repo
//   - TestE2E_BareRepo_DeleteWorktree: deleting worktrees in a bare repo
package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/k1LoW/git-wt/testutil"
)

func TestE2E_BareRepo_ListWorktrees(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("basic", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewBareTestRepo(t)

		out, err := runGitWt(t, binPath, repo.Root)
		if err != nil {
			t.Fatalf("git-wt failed: %v\noutput: %s", err, out)
		}

		// Should contain the bare worktree path
		if !strings.Contains(out, repo.Root) {
			t.Errorf("output should contain bare repo root %q, got: %s", repo.Root, out)
		}

		// Bare entry should show branch and HEAD (not empty)
		if !strings.Contains(out, "main") {
			t.Errorf("output should contain branch 'main' for bare entry, got: %s", out)
		}

		// Bare entry should be labeled as (bare) in BRANCH column
		if !strings.Contains(out, "(bare)") {
			t.Errorf("output should contain '(bare)' marker for bare entry, got: %s", out)
		}
	})

	t.Run("with_worktrees", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewBareTestRepo(t)

		// Create a worktree
		_, err := runGitWt(t, binPath, repo.Root, "feature-a")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root)
		if err != nil {
			t.Fatalf("git-wt failed: %v\noutput: %s", err, out)
		}

		if !strings.Contains(out, repo.Root) {
			t.Errorf("output should contain bare repo root, got: %s", out)
		}
		if !strings.Contains(out, "feature-a") {
			t.Errorf("output should contain 'feature-a', got: %s", out)
		}
	})
}

func TestE2E_BareRepo_CreateWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("new_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewBareTestRepo(t)

		out, err := runGitWt(t, binPath, repo.Root, "feature-branch")
		if err != nil {
			t.Fatalf("git-wt feature-branch failed: %v\noutput: %s", err, out)
		}

		wtPath := worktreePath(out)
		if !strings.Contains(wtPath, "feature-branch") {
			t.Errorf("output should contain worktree path with 'feature-branch', got: %s", wtPath)
		}

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Errorf("worktree directory was not created at %s", wtPath)
		}
	})

	t.Run("existing_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewBareTestRepo(t)

		repo.Git("branch", "existing-branch")

		out, err := runGitWt(t, binPath, repo.Root, "existing-branch")
		if err != nil {
			t.Fatalf("failed to create worktree for existing branch: %v\noutput: %s", err, out)
		}

		wtPath := worktreePath(out)
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Errorf("worktree was not created at %s", wtPath)
		}
	})

	// Regression: "git wt main" in a bare repo should create a worktree,
	// not match the bare entry itself.
	t.Run("head_branch_creates_worktree", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewBareTestRepo(t)

		out, err := runGitWt(t, binPath, repo.Root, "main")
		if err != nil {
			t.Fatalf("git-wt main failed: %v\noutput: %s", err, out)
		}

		wtPath := worktreePath(out)

		// Should create a new worktree, not return the bare repo path
		if wtPath == repo.Root {
			t.Error("git-wt main should create a new worktree, not return the bare repo path")
		}
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Errorf("worktree was not created at %s", wtPath)
		}
	})
}

func TestE2E_BareRepo_DeleteWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("safe_delete", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewBareTestRepo(t)

		out, err := runGitWt(t, binPath, repo.Root, "to-delete")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Fatalf("worktree should exist at %s", wtPath)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-d", "to-delete")
		if err != nil {
			t.Fatalf("git-wt -d failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}
	})

	t.Run("force_delete", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewBareTestRepo(t)

		out, err := runGitWt(t, binPath, repo.Root, "force-del")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Add an untracked file to make worktree dirty
		if err := os.WriteFile(wtPath+"/dirty.txt", []byte("dirty"), 0600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		out, err = runGitWt(t, binPath, repo.Root, "-D", "force-del")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been force deleted")
		}
	})
}
