// bare_test.go contains E2E tests for bare repository detection.
//
// git-wt does not currently support bare repositories. These tests verify
// that clear error messages are returned when running git-wt commands
// in two scenarios:
//   - Directly inside a bare repository
//   - Inside a worktree created from a bare repository
//
// Detection mechanism:
//   `git worktree list --porcelain` outputs a "bare" line for the first entry
//   of a bare repository. This flag is present regardless of whether the command
//   is run from the bare repo itself or from one of its worktrees.
//   See internal/git/repo_context.go for the detection implementation.
package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/exec"
	"github.com/k1LoW/git-wt/testutil"
)

func TestE2E_BareRepository(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	// --- Tests running directly inside a bare repository ---

	t.Run("direct_bare_list", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt with no arguments (list mode) inside the bare repo
		out, err := runGitWt(t, binPath, bareRepo.Root)
		if err == nil {
			t.Fatalf("expected error for bare repository, but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare") {
			t.Errorf("error message should mention 'bare', got: %s", out)
		}
	})

	t.Run("direct_bare_add", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt with a branch name (add/switch mode) inside the bare repo
		out, err := runGitWt(t, binPath, bareRepo.Root, "feature")
		if err == nil {
			t.Fatalf("expected error for bare repository, but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare") {
			t.Errorf("error message should mention 'bare', got: %s", out)
		}
	})

	t.Run("direct_bare_delete", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt with -d flag (delete mode) inside the bare repo
		out, err := runGitWt(t, binPath, bareRepo.Root, "-d", "main")
		if err == nil {
			t.Fatalf("expected error for bare repository, but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare") {
			t.Errorf("error message should mention 'bare', got: %s", out)
		}
	})

	// --- Tests running inside a worktree created from a bare repository ---
	//
	// When a worktree is created from a bare repository, `git worktree list --porcelain`
	// still reports the first (main) entry as "bare". This means the detection logic
	// works identically whether running from the bare repo or its worktree.

	t.Run("worktree_from_bare_list", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a worktree from the bare repo using raw git command
		// (git-wt itself should reject this, so we use git directly)
		wtPath := filepath.Join(bareRepo.ParentDir(), "wt-main")
		cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
		}
		t.Cleanup(func() { os.RemoveAll(wtPath) })

		// Run git-wt with no arguments (list mode) inside the worktree
		out, err := runGitWt(t, binPath, wtPath)
		if err == nil {
			t.Fatalf("expected error for worktree from bare repo, but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare") {
			t.Errorf("error message should mention 'bare', got: %s", out)
		}
	})

	t.Run("worktree_from_bare_add", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a worktree from the bare repo
		wtPath := filepath.Join(bareRepo.ParentDir(), "wt-main")
		cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
		}
		t.Cleanup(func() { os.RemoveAll(wtPath) })

		// Run git-wt with a branch name (add/switch mode) inside the worktree
		out, err := runGitWt(t, binPath, wtPath, "feature")
		if err == nil {
			t.Fatalf("expected error for worktree from bare repo, but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare") {
			t.Errorf("error message should mention 'bare', got: %s", out)
		}
	})

	t.Run("worktree_from_bare_delete", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a worktree from the bare repo
		wtPath := filepath.Join(bareRepo.ParentDir(), "wt-main")
		cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
		}
		t.Cleanup(func() { os.RemoveAll(wtPath) })

		// Run git-wt with -d flag (delete mode) inside the worktree
		out, err := runGitWt(t, binPath, wtPath, "-d", "main")
		if err == nil {
			t.Fatalf("expected error for worktree from bare repo, but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare") {
			t.Errorf("error message should mention 'bare', got: %s", out)
		}
	})
}
