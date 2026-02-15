// bare_test.go contains E2E tests for bare repository support.
//
// Covered scenarios:
//   - List operation: supported in both bare root and worktrees from bare repos
//   - Add/switch, delete operations: not yet supported, should return errors
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
			t.Errorf("output should contain '(bare)' label, got: %s", out) //nostyle:errorstrings
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
		cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
		}
		t.Cleanup(func() { os.RemoveAll(wtPath) })

		// Run git-wt with no arguments (list mode) inside the worktree
		out, err := runGitWt(t, binPath, wtPath)
		if err != nil {
			t.Fatalf("expected success for worktree from bare repo list, but got error: %v\noutput: %s", err, out)
		}
		// Should show both the bare entry and the worktree entry
		if !strings.Contains(out, "(bare)") {
			t.Errorf("output should contain '(bare)' label, got: %s", out) //nostyle:errorstrings
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
		cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
		}
		t.Cleanup(func() { os.RemoveAll(wtPath) })

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
			t.Errorf("output should contain '(bare)' label, got: %s", out) //nostyle:errorstrings
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

	// --- Tests for operations that still don't support bare repositories ---

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
	// (operations that still don't support bare repos)

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
