// bare_test.go contains E2E tests for bare repository support.
//
// Covered scenarios:
//   - List operation: supported in both bare root and worktrees from bare repos
//   - Add/switch operations: supported in both bare root and worktrees from bare repos
//   - Delete operation: supported from bare-derived worktrees; bare entry itself is protected
package e2e

import (
	"encoding/json"
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

	// --- Tests for list operation in bare repositories ---

	t.Run("direct_bare_list", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt with no arguments (list mode) inside the bare repo
		out, err := runGitWt(t, binPath, bareRepo.Root)
		if err != nil {
			t.Fatalf("expected success for bare repository list, but got error: %v\noutput: %s", err, out)
		}
		if !strings.Contains(out, "(bare)") {
			t.Errorf("output should contain '(bare)' label, got: %s", out)
		}
	})

	t.Run("direct_bare_list_json", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt with --json flag inside the bare repo
		out, err := runGitWt(t, binPath, bareRepo.Root, "--json")
		if err != nil {
			t.Fatalf("expected success for bare repository list --json, but got error: %v\noutput: %s", err, out)
		}

		var entries []struct {
			Path    string `json:"path"`
			Branch  string `json:"branch"`
			Head    string `json:"head"`
			Bare    bool   `json:"bare"`
			Current bool   `json:"current"`
		}
		if err := json.Unmarshal([]byte(out), &entries); err != nil {
			t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, out)
		}
		if len(entries) == 0 {
			t.Fatal("expected at least one entry in JSON output")
		}
		// The first entry should be the bare repo itself
		if !entries[0].Bare {
			t.Errorf("first entry should have bare=true, got: %+v", entries[0])
		}
	})

	t.Run("direct_bare_list_current_marker", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt inside the bare repo - the bare entry should have * marker
		out, err := runGitWt(t, binPath, bareRepo.Root)
		if err != nil {
			t.Fatalf("expected success, but got error: %v\noutput: %s", err, out)
		}

		// Find the line with (bare) and check it has * marker
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "(bare)") {
				if !strings.Contains(line, "*") {
					t.Errorf("bare entry should have * marker when running from bare root, got: %s", line)
				}
				return
			}
		}
		t.Errorf("no line with (bare) found in output: %s", out)
	})

	// --- Tests for list operation in worktrees from bare repositories ---

	t.Run("worktree_from_bare_list", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a worktree from the bare repo using raw git command
		wtPath := filepath.Join(bareRepo.ParentDir(), "wt-main")
		addRawWorktreeFromBare(t, bareRepo.Root, wtPath, "main")

		// Run git-wt with no arguments (list mode) inside the worktree
		out, err := runGitWt(t, binPath, wtPath)
		if err != nil {
			t.Fatalf("expected success for worktree from bare repo list, but got error: %v\noutput: %s", err, out)
		}
		// Should show both the bare entry and the worktree entry
		if !strings.Contains(out, "(bare)") {
			t.Errorf("output should contain '(bare)' label, got: %s", out)
		}
		if !strings.Contains(out, "main") {
			t.Errorf("output should contain 'main' branch, got: %s", out)
		}
	})

	t.Run("worktree_from_bare_list_current_marker", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a worktree from the bare repo
		wtPath := filepath.Join(bareRepo.ParentDir(), "wt-main")
		addRawWorktreeFromBare(t, bareRepo.Root, wtPath, "main")

		// Run git-wt from the worktree - worktree entry should have * marker, bare entry should not
		out, err := runGitWt(t, binPath, wtPath)
		if err != nil {
			t.Fatalf("expected success, but got error: %v\noutput: %s", err, out)
		}

		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(line, "(bare)") {
				if strings.Contains(line, "*") {
					t.Errorf("bare entry should NOT have * marker when running from worktree, got: %s", line)
				}
			}
			if strings.Contains(line, "main") && !strings.Contains(line, "(bare)") {
				if !strings.Contains(line, "*") {
					t.Errorf("worktree entry should have * marker when running from worktree, got: %s", line)
				}
			}
		}
	})

	// --- Tests for bare repos with .git directory name (core.bare = true) ---

	t.Run("dotgit_bare_list", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewDotGitBareTestRepo(t)

		out, err := runGitWt(t, binPath, bareRepo.Root)
		if err != nil {
			t.Fatalf("expected success for dotgit bare repository list, but got error: %v\noutput: %s", err, out)
		}
		if !strings.Contains(out, "(bare)") {
			t.Errorf("output should contain '(bare)' label, got: %s", out)
		}
	})

	t.Run("dotgit_bare_list_json", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewDotGitBareTestRepo(t)

		out, err := runGitWt(t, binPath, bareRepo.Root, "--json")
		if err != nil {
			t.Fatalf("expected success for dotgit bare repository list --json, but got error: %v\noutput: %s", err, out)
		}

		var entries []struct {
			Bare    bool `json:"bare"`
			Current bool `json:"current"`
		}
		if err := json.Unmarshal([]byte(out), &entries); err != nil {
			t.Fatalf("failed to parse JSON output: %v\noutput: %s", err, out)
		}
		if len(entries) == 0 {
			t.Fatal("expected at least one entry in JSON output")
		}
		if !entries[0].Bare {
			t.Errorf("first entry should have bare=true, got: %+v", entries[0])
		}
		if !entries[0].Current {
			t.Errorf("first entry should have current=true when running from bare root, got: %+v", entries[0])
		}
	})

	// --- Tests for add/switch operations in bare repositories ---

	t.Run("direct_bare_add", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt with a branch name (add mode) inside the bare repo
		// Should create a new worktree with a new branch
		stdout, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "feature")
		if err != nil {
			t.Fatalf("expected success for bare repository add, but got error: %v\nstdout: %s", err, stdout)
		}
		wtPath := worktreePath(stdout)
		assertWorktreeExists(t, wtPath)
	})

	t.Run("direct_bare_add_existing_branch", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a branch in the bare repo
		cmd := exec.Command("git", "-C", bareRepo.Root, "branch", "existing-branch", "main")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git branch failed: %v\noutput: %s", err, out)
		}

		// Run git-wt with the existing branch name
		stdout, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "existing-branch")
		if err != nil {
			t.Fatalf("expected success for bare repository add with existing branch, but got error: %v\nstdout: %s", err, stdout)
		}
		wtPath := worktreePath(stdout)
		assertWorktreeExists(t, wtPath)
	})

	t.Run("direct_bare_add_with_start_point", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Run git-wt with branch name and start-point
		stdout, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "feature-from-main", "main")
		if err != nil {
			t.Fatalf("expected success for bare repository add with start-point, but got error: %v\nstdout: %s", err, stdout)
		}
		wtPath := worktreePath(stdout)
		assertWorktreeExists(t, wtPath)
	})

	t.Run("direct_bare_switch_existing", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// First, create a worktree
		stdout1, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "switch-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v\nstdout: %s", err, stdout1)
		}
		createdPath := worktreePath(stdout1)

		// Run git-wt again with the same branch - should switch (return same path)
		stdout2, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "switch-test")
		if err != nil {
			t.Fatalf("expected success for switch to existing worktree, but got error: %v\nstdout: %s", err, stdout2)
		}
		switchPath := worktreePath(stdout2)
		if switchPath != createdPath {
			t.Errorf("switch should return same path as creation\ncreated: %s\nswitch:  %s", createdPath, switchPath)
		}
	})

	t.Run("dotgit_bare_add", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewDotGitBareTestRepo(t)

		// Run git-wt with a branch name inside the dotgit bare repo
		stdout, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "feature")
		if err != nil {
			t.Fatalf("expected success for dotgit bare repository add, but got error: %v\nstdout: %s", err, stdout)
		}
		wtPath := worktreePath(stdout)
		assertWorktreeExists(t, wtPath)
	})

	// --- Tests running inside a worktree created from a bare repository ---

	t.Run("worktree_from_bare_add", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a worktree from the bare repo
		wtPath := filepath.Join(bareRepo.ParentDir(), "wt-main")
		addRawWorktreeFromBare(t, bareRepo.Root, wtPath, "main")

		// Run git-wt with a branch name (add mode) inside the worktree
		// Should succeed: bare-derived worktrees support add
		stdout, _, err := runGitWtStdout(t, binPath, wtPath, "feature2")
		if err != nil {
			t.Fatalf("expected success for worktree from bare repo add, but got error: %v\nstdout: %s", err, stdout)
		}
		newWtPath := worktreePath(stdout)
		assertWorktreeExists(t, newWtPath)
	})

	t.Run("worktree_from_bare_add_copies_files", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Create a worktree from the bare repo
		wtPath := filepath.Join(bareRepo.ParentDir(), "wt-main")
		addRawWorktreeFromBare(t, bareRepo.Root, wtPath, "main")

		// Create an untracked file in the source worktree to test copy behavior
		if err := os.WriteFile(filepath.Join(wtPath, "untracked.txt"), []byte("test content\n"), 0600); err != nil {
			t.Fatalf("failed to create untracked.txt: %v", err)
		}

		// Run git-wt with --copyuntracked to copy untracked files
		stdout, _, err := runGitWtStdout(t, binPath, wtPath, "--copyuntracked", "feature-copy")
		if err != nil {
			t.Fatalf("expected success, but got error: %v\nstdout: %s", err, stdout)
		}
		newWtPath := worktreePath(stdout)

		// Verify the untracked file was copied to the new worktree
		copiedPath := filepath.Join(newWtPath, "untracked.txt")
		if _, err := os.Stat(copiedPath); os.IsNotExist(err) {
			t.Errorf("untracked.txt should be copied to new worktree at %s", copiedPath)
		}
	})

	t.Run("bare_add_chain", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Step 1: Create worktree A from bare root
		stdoutA, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "feature-a")
		if err != nil {
			t.Fatalf("failed to create worktree A from bare root: %v\nstdout: %s", err, stdoutA)
		}
		wtPathA := worktreePath(stdoutA)
		if wtPathA == "" {
			t.Fatal("expected worktree A path in stdout, got empty")
		}

		// Step 2: Create worktree B from worktree A (bare-derived)
		stdoutB, _, err := runGitWtStdout(t, binPath, wtPathA, "feature-b")
		if err != nil {
			t.Fatalf("failed to create worktree B from worktree A: %v\nstdout: %s", err, stdoutB)
		}
		wtPathB := worktreePath(stdoutB)
		assertWorktreeExists(t, wtPathB)

		// Step 3: Switch back to A from bare root (should return existing path)
		stdoutSwitch, _, err := runGitWtStdout(t, binPath, bareRepo.Root, "feature-a")
		if err != nil {
			t.Fatalf("failed to switch to worktree A: %v\nstdout: %s", err, stdoutSwitch)
		}
		switchPath := worktreePath(stdoutSwitch)
		if switchPath != wtPathA {
			t.Errorf("switch should return same path as creation\ncreated: %s\nswitch:  %s", wtPathA, switchPath)
		}
	})

}

// TestE2E_BareDelete tests delete operations in bare repository environments.
func TestE2E_BareDelete(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	// --- Success: delete linked worktree from bare root ---

	t.Run("bare_root_safe_delete", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)
		wtPath := createBareWorktree(t, binPath, bareRepo.Root, "feature-del")

		// Safe delete from bare root
		out, err := runGitWt(t, binPath, bareRepo.Root, "-d", "feature-del")
		if err != nil {
			t.Fatalf("git-wt -d failed: %v\noutput: %s", err, out)
		}

		// Worktree should be deleted
		assertWorktreeDeleted(t, wtPath)
		if !strings.Contains(out, "Deleted") {
			t.Errorf("output should contain 'Deleted', got: %s", out)
		}
	})

	t.Run("bare_root_force_delete", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)
		wtPath := createBareWorktree(t, binPath, bareRepo.Root, "unmerged-del")
		commitUnmergedChange(t, wtPath)

		// Force delete from bare root
		out, err := runGitWt(t, binPath, bareRepo.Root, "-D", "unmerged-del")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\noutput: %s", err, out)
		}

		// Worktree should be deleted
		assertWorktreeDeleted(t, wtPath)
	})

	// --- Error: attempting to delete bare entry itself ---

	t.Run("bare_root_delete_self", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Try to delete main (the bare entry's branch) from bare root
		out, err := runGitWt(t, binPath, bareRepo.Root, "-d", "main")
		if err == nil {
			t.Fatalf("expected error when deleting bare entry, but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare repository entry") {
			t.Errorf("error message should mention 'bare repository entry', got: %s", out)
		}
	})

	t.Run("bare_root_delete_self_dot", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)

		// Try to delete "." (current directory = bare root) from bare root
		out, err := runGitWt(t, binPath, bareRepo.Root, "-d", ".")
		if err == nil {
			t.Fatalf("expected error when deleting bare entry via '.', but succeeded with output: %s", out)
		}
		if !strings.Contains(out, "bare repository entry") {
			t.Errorf("error message should mention 'bare repository entry', got: %s", out)
		}
	})

	// --- Success: delete from bare-derived worktree ---

	t.Run("bare_worktree_delete_other", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)
		wtPathA := createBareWorktree(t, binPath, bareRepo.Root, "wt-a")
		wtPathB := createBareWorktree(t, binPath, bareRepo.Root, "wt-b")

		// From worktree B, delete worktree A
		out, err := runGitWt(t, binPath, wtPathB, "-d", "wt-a")
		if err != nil {
			t.Fatalf("git-wt -d failed: %v\noutput: %s", err, out)
		}

		// Worktree A should be deleted
		assertWorktreeDeleted(t, wtPathA)
		// Worktree B should still exist
		if _, err := os.Stat(wtPathB); os.IsNotExist(err) {
			t.Error("worktree B should still exist")
		}
	})

	t.Run("bare_worktree_delete_current", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)
		wtPath := createBareWorktree(t, binPath, bareRepo.Root, "current-del")

		// Delete current worktree from inside it (with shell integration)
		stdout, stderr, err := runGitWtWithShellIntegration(t, binPath, wtPath, "-D", "current-del")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\nstderr: %s", err, stderr)
		}

		// Worktree should be deleted
		assertWorktreeDeleted(t, wtPath)

		// Shell integration should output bare root path
		assertLastLine(t, stdout, bareRepo.Root)
	})

	// --- Delete last worktree, cd back to bare root ---

	t.Run("bare_delete_last_worktree_cd", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)
		wtPath := createBareWorktree(t, binPath, bareRepo.Root, "only-wt")

		// Delete from inside the worktree (with shell integration)
		stdout, stderr, err := runGitWtWithShellIntegration(t, binPath, wtPath, "-D", "only-wt")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\nstderr: %s", err, stderr)
		}

		// Worktree should be deleted
		assertWorktreeDeleted(t, wtPath)

		// Shell integration should output bare root path
		assertLastLine(t, stdout, bareRepo.Root)
	})

	// --- Error: modified files ---

	t.Run("bare_worktree_delete_modified", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewBareTestRepo(t)
		wtPath := createBareWorktree(t, binPath, bareRepo.Root, "mod-wt")

		// Add an untracked file
		if err := os.WriteFile(filepath.Join(wtPath, "untracked.txt"), []byte("content"), 0600); err != nil {
			t.Fatalf("failed to create untracked file: %v", err)
		}

		// Safe delete should fail
		out, err := runGitWt(t, binPath, bareRepo.Root, "-d", "mod-wt")
		if err == nil {
			t.Fatal("git-wt -d should fail when worktree has untracked files")
		}
		if !strings.Contains(out, "use -D to force deletion") {
			t.Errorf("error should suggest '-D to force deletion', got: %s", out)
		}

		// Worktree should still exist
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Error("worktree should NOT have been deleted")
		}
	})

	// --- dotgit bare variant ---

	t.Run("dotgit_bare_delete", func(t *testing.T) {
		t.Parallel()
		bareRepo := testutil.NewDotGitBareTestRepo(t)
		wtPath := createBareWorktree(t, binPath, bareRepo.Root, "dotgit-del")

		// Safe delete from bare root
		out, err := runGitWt(t, binPath, bareRepo.Root, "-d", "dotgit-del")
		if err != nil {
			t.Fatalf("git-wt -d failed: %v\noutput: %s", err, out)
		}

		// Worktree should be deleted
		assertWorktreeDeleted(t, wtPath)
	})
}
