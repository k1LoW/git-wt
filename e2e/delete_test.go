// delete_test.go contains worktree/branch deletion tests:
//   - TestE2E_DeleteWorktree: worktree deletion (safe, force, unmerged, multiple)
//   - TestE2E_DeleteBranch: branch-only deletion
//   - TestE2E_DeleteCurrentWorktree: deleting worktree while inside it
package e2e

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/exec"
	"github.com/k1LoW/git-wt/testutil"
)

func TestE2E_DeleteWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("safe_delete", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

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
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "unmerged")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Make a commit in the worktree (will be unmerged)
		commitUnmergedChange(t, wtPath)

		out, err = runGitWt(t, binPath, repo.Root, "-D", "unmerged")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been force deleted")
		}
	})

	t.Run("safe_delete_with_untracked_files", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "untracked-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Add an untracked file (not staged or committed)
		if err := os.WriteFile(filepath.Join(wtPath, "untracked.txt"), []byte("content"), 0600); err != nil {
			t.Fatalf("failed to create untracked file: %v", err)
		}

		// Safe delete should fail with our custom message
		out, err = runGitWt(t, binPath, repo.Root, "-d", "untracked-test")
		if err == nil {
			t.Fatal("git-wt -d should fail when worktree has untracked files")
		}

		if !strings.Contains(out, "has untracked files") {
			t.Errorf("error should mention 'has untracked files', got: %s", out)
		}
		if !strings.Contains(out, "use -D to force deletion") {
			t.Errorf("error should suggest 'use -D to force deletion', got: %s", out)
		}

		// Worktree should still exist
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Error("worktree should NOT have been deleted")
		}

		// Force delete should succeed
		out, err = runGitWt(t, binPath, repo.Root, "-D", "untracked-test")
		if err != nil {
			t.Fatalf("git-wt -D should succeed: %v\noutput: %s", err, out)
		}

		// Worktree should now be deleted
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted with -D")
		}
	})

	t.Run("safe_delete_with_modified_files", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "modified-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Modify a tracked file (README.md exists in the worktree)
		if err := os.WriteFile(filepath.Join(wtPath, "README.md"), []byte("# Modified"), 0600); err != nil {
			t.Fatalf("failed to modify file: %v", err)
		}

		// Safe delete should fail with our custom message
		out, err = runGitWt(t, binPath, repo.Root, "-d", "modified-test")
		if err == nil {
			t.Fatal("git-wt -d should fail when worktree has modified files")
		}

		if !strings.Contains(out, "has modified files") {
			t.Errorf("error should mention 'has modified files', got: %s", out)
		}
		if !strings.Contains(out, "use -D to force deletion") {
			t.Errorf("error should suggest 'use -D to force deletion', got: %s", out)
		}

		// Worktree should still exist
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			t.Error("worktree should NOT have been deleted")
		}

		// Force delete should succeed
		out, err = runGitWt(t, binPath, repo.Root, "-D", "modified-test")
		if err != nil {
			t.Fatalf("git-wt -D should succeed: %v\noutput: %s", err, out)
		}

		// Worktree should now be deleted
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted with -D")
		}
	})

	// PR #64 fix: worktree deletion succeeds even when branch deletion fails
	t.Run("with_unmerged_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "unmerged-branch")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		commitUnmergedChange(t, wtPath)

		// Delete worktree with -d (safe delete) - should succeed but not delete branch
		out, err = runGitWt(t, binPath, repo.Root, "-d", "unmerged-branch")
		if err != nil {
			t.Fatalf("git-wt -d should succeed even when branch deletion fails: %v\noutput: %s", err, out)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}

		if !strings.Contains(out, "failed to delete branch") {
			t.Errorf("output should mention branch deletion failure, got: %s", out)
		}
		if !strings.Contains(out, "use -D to force") {
			t.Errorf("output should suggest using -D to force, got: %s", out)
		}

		// Verify branch still exists
		cmd := exec.Command("git", "branch", "--list", "unmerged-branch")
		cmd.Dir = repo.Root
		branchOut, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if !strings.Contains(string(branchOut), "unmerged-branch") {
			t.Error("branch should still exist after worktree deletion with -d")
		}
	})

	// PR #137 fix: delete worktree from a derived branch should succeed
	// When on branch b (derived from a), deleting worktree a with -d should
	// succeed because a is fully merged into the current HEAD (b).
	t.Run("delete_from_derived_branch", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree a and make a commit
		out, err := runGitWt(t, binPath, repo.Root, "branch-a")
		if err != nil {
			t.Fatalf("failed to create worktree branch-a: %v", err)
		}
		wtPathA := worktreePath(out)

		if err := os.WriteFile(filepath.Join(wtPathA, "a.txt"), []byte("a"), 0600); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		cmd := exec.Command("git", "add", "-A")
		cmd.Dir = wtPathA
		if err := cmd.Run(); err != nil {
			t.Fatalf("git add failed: %v", err)
		}
		cmd = exec.Command("git", "commit", "-m", "commit on a")
		cmd.Dir = wtPathA
		if err := cmd.Run(); err != nil {
			t.Fatalf("git commit failed: %v", err)
		}

		// Create worktree b from branch-a (b includes a's commit)
		out, err = runGitWt(t, binPath, wtPathA, "branch-b")
		if err != nil {
			t.Fatalf("failed to create worktree branch-b: %v", err)
		}
		wtPathB := worktreePath(out)

		// From branch-b, delete worktree branch-a with safe delete
		// This should succeed because branch-a is fully merged into branch-b's HEAD
		out, err = runGitWt(t, binPath, wtPathB, "-d", "branch-a")
		if err != nil {
			t.Fatalf("git-wt -d should succeed when branch is merged into current HEAD: %v\noutput: %s", err, out)
		}

		// Worktree should be deleted
		if _, err := os.Stat(wtPathA); !os.IsNotExist(err) {
			t.Error("worktree branch-a should have been deleted")
		}

		// Branch should also be deleted (fully merged)
		if strings.Contains(out, "failed to delete branch") {
			t.Errorf("branch deletion should succeed since branch-a is merged into branch-b, got: %s", out)
		}

		// Verify branch no longer exists
		cmd = exec.Command("git", "branch", "--list", "branch-a")
		cmd.Dir = repo.Root
		branchOut, err := cmd.Output()
		if err != nil {
			t.Fatalf("git branch --list failed: %v", err)
		}
		if strings.Contains(string(branchOut), "branch-a") {
			t.Error("branch-a should have been deleted since it was fully merged into branch-b")
		}
	})

	// PR #64 fix: shell integration works correctly when branch deletion fails
	t.Run("with_unmerged_branch_shell_integration", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name       string
			shell      string
			scriptFunc func(repoRoot, wtPath, pathDir, branchName string) string
		}{
			{
				name:  "bash",
				shell: "bash",
				scriptFunc: func(repoRoot, wtPath, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
export PATH="%s:$PATH"
eval "$(git wt --init bash)"
cd %q
git wt -d %s
pwd
`, pathDir, wtPath, branchName)
				},
			},
			{
				name:  "zsh",
				shell: "zsh",
				scriptFunc: func(repoRoot, wtPath, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"
cd %q
git wt -d %s
pwd
`, pathDir, wtPath, branchName)
				},
			},
			{
				name:  "fish",
				shell: "fish",
				scriptFunc: func(repoRoot, wtPath, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -x PATH %s $PATH
git wt --init fish | source
cd %q
git wt -d %s
pwd
`, pathDir, wtPath, branchName)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				if _, err := exec.LookPath(tt.shell); err != nil {
					t.Skipf("%s not available", tt.shell)
				}

				repo := testutil.NewTestRepo(t)
				repo.CreateFile("README.md", "# Test")
				repo.Commit("initial commit")

				branchName := fmt.Sprintf("unmerged-%s-test", tt.shell)

				out, err := runGitWt(t, binPath, repo.Root, branchName)
				if err != nil {
					t.Fatalf("failed to create worktree: %v", err)
				}
				wtPath := worktreePath(out)

				commitUnmergedChange(t, wtPath)

				script := tt.scriptFunc(repo.Root, wtPath, filepath.Dir(binPath), branchName)
				cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
				cmdOut, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%s shell integration failed: %v\noutput: %s", tt.shell, err, cmdOut)
				}

				output := strings.TrimSpace(string(cmdOut))
				lines := strings.Split(output, "\n")
				pwd := lines[len(lines)-1]

				if pwd != repo.Root {
					t.Errorf("pwd should be main repo root %q after deleting current worktree, got: %s\nfull output: %s", repo.Root, pwd, output)
				}
			})
		}
	})

	t.Run("multiple", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		outA, err := runGitWt(t, binPath, repo.Root, "del-a")
		if err != nil {
			t.Fatalf("failed to create worktree del-a: %v", err)
		}
		outB, err := runGitWt(t, binPath, repo.Root, "del-b")
		if err != nil {
			t.Fatalf("failed to create worktree del-b: %v", err)
		}
		outC, err := runGitWt(t, binPath, repo.Root, "del-c")
		if err != nil {
			t.Fatalf("failed to create worktree del-c: %v", err)
		}

		pathA := worktreePath(outA)
		pathB := worktreePath(outB)
		pathC := worktreePath(outC)

		out, err := runGitWt(t, binPath, repo.Root, "-D", "del-a", "del-b", "del-c")
		if err != nil {
			t.Fatalf("git-wt -D with multiple args failed: %v\noutput: %s", err, out)
		}

		for _, path := range []string{pathA, pathB, pathC} {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("worktree should have been deleted at %s", path)
			}
		}

		if !strings.Contains(out, "del-a") {
			t.Errorf("output should contain deletion message for 'del-a', got: %s", out)
		}
		if !strings.Contains(out, "del-b") {
			t.Errorf("output should contain deletion message for 'del-b', got: %s", out)
		}
		if !strings.Contains(out, "del-c") {
			t.Errorf("output should contain deletion message for 'del-c', got: %s", out)
		}
	})

	t.Run("multiple_stops_on_error", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		_, err := runGitWt(t, binPath, repo.Root, "stop-a")
		if err != nil {
			t.Fatalf("failed to create worktree stop-a: %v", err)
		}
		outC, err := runGitWt(t, binPath, repo.Root, "stop-c")
		if err != nil {
			t.Fatalf("failed to create worktree stop-c: %v", err)
		}
		pathC := worktreePath(outC)

		// Try to delete: stop-a (exists), stop-b (does NOT exist), stop-c (exists)
		out, err := runGitWt(t, binPath, repo.Root, "-D", "stop-a", "stop-b", "stop-c")
		if err == nil {
			t.Fatal("command should fail when deleting non-existent worktree")
		}

		if !strings.Contains(out, "stop-b") {
			t.Errorf("error should mention 'stop-b', got: %s", out)
		}

		// stop-c should NOT be deleted (execution stopped at stop-b error)
		if _, err := os.Stat(pathC); os.IsNotExist(err) {
			t.Error("stop-c should NOT have been deleted (execution should stop on error)")
		}
	})
}

func TestE2E_DeleteBranch(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("branch_only", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		branchName := "branch-only"
		cmd := exec.Command("git", "branch", branchName)
		cmd.Dir = repo.Root
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to create branch: %v", err)
		}

		cmd = exec.Command("git", "branch", "--list", branchName)
		cmd.Dir = repo.Root
		out, err := cmd.Output()
		if err != nil || !strings.Contains(string(out), branchName) {
			t.Fatalf("branch should exist before deletion")
		}

		wtOut, err := runGitWt(t, binPath, repo.Root, "-D", branchName)
		if err != nil {
			t.Fatalf("failed to delete branch-only: %v, output: %s", err, wtOut)
		}

		if !strings.Contains(wtOut, "Deleted branch") || !strings.Contains(wtOut, "no worktree was associated") {
			t.Errorf("output should indicate branch-only deletion, got: %s", wtOut)
		}

		cmd = exec.Command("git", "branch", "--list", branchName)
		cmd.Dir = repo.Root
		out, err = cmd.Output()
		if err != nil {
			t.Fatalf("failed to list branches: %v", err)
		}
		if strings.Contains(string(out), branchName) {
			t.Error("branch should have been deleted")
		}
	})

	t.Run("not_exists", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "-D", "non-existent-branch")
		if err == nil {
			t.Fatal("should fail when deleting non-existent branch")
		}

		if !strings.Contains(out, "no worktree or branch found") {
			t.Errorf("error should mention 'no worktree or branch found', got: %s", out)
		}
	})
}

// TestE2E_DeleteCurrentWorktree tests deleting the worktree you're currently in.
// This tests the fix for issue #58: safely remove current worktree and return to repository root.
func TestE2E_DeleteCurrentWorktree(t *testing.T) {
	t.Parallel()
	binPath := buildBinary(t)

	t.Run("outputs_main_repo_path_with_shell_integration", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "current-wt")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		cmd := exec.Command(binPath, "-D", "current-wt")
		cmd.Dir = wtPath
		cmd.Env = append(os.Environ(), "GIT_WT_SHELL_INTEGRATION=1")
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("git-wt -D failed: %v\nstderr: %s", err, stderrBuf.String())
		}
		stdout := strings.TrimSpace(stdoutBuf.String())
		if stdout == "" {
			t.Fatal("expected output but got none")
		}

		lines := strings.Split(stdout, "\n")
		lastLine := lines[len(lines)-1]

		if lastLine != repo.Root {
			t.Errorf("last line should be main repo path %q, got %q", repo.Root, lastLine)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}
	})

	t.Run("no_path_without_shell_integration", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "current-wt-no-shell")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		stdout, stderr, err := runGitWtStdout(t, binPath, wtPath, "-D", "current-wt-no-shell")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\nstderr: %s", err, stderr)
		}

		trimmed := strings.TrimSpace(stdout)
		if trimmed == "" {
			t.Fatal("expected output but got none")
		}

		lines := strings.Split(trimmed, "\n")
		lastLine := lines[len(lines)-1]

		if !strings.Contains(lastLine, "Deleted") {
			t.Errorf("last line should be a deletion message containing %q, got %q", "Deleted", lastLine)
		}

		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}
	})

	t.Run("delete_other_worktree_does_not_output_path", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		_, err := runGitWt(t, binPath, repo.Root, "wt-a")
		if err != nil {
			t.Fatalf("failed to create worktree wt-a: %v", err)
		}
		outB, err := runGitWt(t, binPath, repo.Root, "wt-b")
		if err != nil {
			t.Fatalf("failed to create worktree wt-b: %v", err)
		}
		wtPathB := worktreePath(outB)

		stdout, stderr, err := runGitWtStdout(t, binPath, wtPathB, "-D", "wt-a")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\nstderr: %s", err, stderr)
		}

		trimmed := strings.TrimSpace(stdout)
		if trimmed == "" {
			t.Fatal("expected output but got none")
		}

		lines := strings.Split(trimmed, "\n")
		lastLine := lines[len(lines)-1]

		if !strings.Contains(lastLine, "Deleted") {
			t.Errorf("last line should be a deletion message containing %q, got %q", "Deleted", lastLine)
		}
	})

	t.Run("safe_delete_with_shell_integration", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "safe-del-wt")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		cmd := exec.Command(binPath, "-d", "safe-del-wt")
		cmd.Dir = wtPath
		cmd.Env = append(os.Environ(), "GIT_WT_SHELL_INTEGRATION=1")
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("git-wt -d failed: %v\nstderr: %s", err, stderrBuf.String())
		}
		stdout := strings.TrimSpace(stdoutBuf.String())
		if stdout == "" {
			t.Fatal("expected output but got none")
		}

		lines := strings.Split(stdout, "\n")
		lastLine := lines[len(lines)-1]

		if lastLine != repo.Root {
			t.Errorf("last line should be main repo path %q, got %q", repo.Root, lastLine)
		}
	})

	t.Run("shell_integration", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name       string
			shell      string
			scriptFunc func(repoRoot, wtPath, pathDir, branchName string) string
		}{
			{
				name:  "bash",
				shell: "bash",
				scriptFunc: func(repoRoot, wtPath, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
export PATH="%s:$PATH"
eval "$(git wt --init bash)"
cd %q
git wt -D %s
pwd
`, pathDir, wtPath, branchName)
				},
			},
			{
				name:  "zsh",
				shell: "zsh",
				scriptFunc: func(repoRoot, wtPath, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -e
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"
cd %q
git wt -D %s
pwd
`, pathDir, wtPath, branchName)
				},
			},
			{
				name:  "fish",
				shell: "fish",
				scriptFunc: func(repoRoot, wtPath, pathDir, branchName string) string {
					return fmt.Sprintf(`
set -x PATH %s $PATH
git wt --init fish | source
cd %q
git wt -D %s
pwd
`, pathDir, wtPath, branchName)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				if _, err := exec.LookPath(tt.shell); err != nil {
					t.Skipf("%s not available", tt.shell)
				}

				repo := testutil.NewTestRepo(t)
				repo.CreateFile("README.md", "# Test")
				repo.Commit("initial commit")

				branchName := fmt.Sprintf("del-%s-test", tt.shell)

				out, err := runGitWt(t, binPath, repo.Root, branchName)
				if err != nil {
					t.Fatalf("failed to create worktree: %v", err)
				}
				wtPath := worktreePath(out)

				script := tt.scriptFunc(repo.Root, wtPath, filepath.Dir(binPath), branchName)
				cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
				cmdOut, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("%s shell integration failed: %v\noutput: %s", tt.shell, err, cmdOut)
				}

				output := strings.TrimSpace(string(cmdOut))
				lines := strings.Split(output, "\n")
				pwd := lines[len(lines)-1]

				if pwd != repo.Root {
					t.Errorf("pwd should be main repo root %q after deleting current worktree, got: %s", repo.Root, pwd)
				}
			})
		}
	})

	t.Run("delete_with_dot_path", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "dot-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Run from inside the worktree with "." as the target
		cmd := exec.Command(binPath, "-D", ".")
		cmd.Dir = wtPath
		cmd.Env = append(os.Environ(), "GIT_WT_SHELL_INTEGRATION=1")
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("git-wt -D . failed: %v\nstderr: %s", err, stderrBuf.String())
		}

		// Verify worktree was deleted
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}

		// Verify last line is main repo path (for shell integration)
		stdout := strings.TrimSpace(stdoutBuf.String())
		lines := strings.Split(stdout, "\n")
		lastLine := lines[len(lines)-1]
		if lastLine != repo.Root {
			t.Errorf("last line should be main repo path %q, got %q", repo.Root, lastLine)
		}
	})

	t.Run("delete_with_relative_path", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create two worktrees
		_, err := runGitWt(t, binPath, repo.Root, "wt-a")
		if err != nil {
			t.Fatalf("failed to create worktree wt-a: %v", err)
		}
		outB, err := runGitWt(t, binPath, repo.Root, "wt-b")
		if err != nil {
			t.Fatalf("failed to create worktree wt-b: %v", err)
		}
		wtPathA := filepath.Join(filepath.Dir(worktreePath(outB)), "wt-a")
		wtPathB := worktreePath(outB)

		// From inside wt-b, delete wt-a using relative path ../wt-a
		cmd := exec.Command(binPath, "-D", "../wt-a")
		cmd.Dir = wtPathB
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("git-wt -D ../wt-a failed: %v\nstderr: %s", err, stderrBuf.String())
		}

		// Verify wt-a was deleted
		if _, err := os.Stat(wtPathA); !os.IsNotExist(err) {
			t.Error("worktree wt-a should have been deleted")
		}

		// Verify wt-b still exists
		if _, err := os.Stat(wtPathB); os.IsNotExist(err) {
			t.Error("worktree wt-b should still exist")
		}
	})

	t.Run("delete_with_absolute_path", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		out, err := runGitWt(t, binPath, repo.Root, "abs-test")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		if !filepath.IsAbs(wtPath) {
			t.Fatalf("expected absolute path, got %q", wtPath)
		}

		// Delete using absolute path
		cmd := exec.Command(binPath, "-D", wtPath)
		cmd.Dir = repo.Root
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("git-wt -D %s failed: %v\nstderr: %s", wtPath, err, stderrBuf.String())
		}

		// Verify worktree was deleted
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}
	})

	t.Run("delete_worktree_name_takes_precedence_over_filesystem_path", func(t *testing.T) {
		t.Parallel()
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create a worktree named "test"
		out, err := runGitWt(t, binPath, repo.Root, "test")
		if err != nil {
			t.Fatalf("failed to create worktree test: %v", err)
		}
		wtPath := worktreePath(out)

		// Create a local directory named "test" inside the main repo
		// This simulates: repo has a folder "test", and there's also a worktree "test"
		localDir := filepath.Join(repo.Root, "test")
		if err := os.Mkdir(localDir, 0755); err != nil {
			t.Fatalf("failed to create local directory: %v", err)
		}

		// From main repo, delete "test" - should delete the worktree, not be confused by local dir
		cmd := exec.Command(binPath, "-D", "test")
		cmd.Dir = repo.Root
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf
		if err := cmd.Run(); err != nil {
			t.Fatalf("git-wt -D test failed: %v\nstderr: %s", err, stderrBuf.String())
		}

		// Verify worktree was deleted (matched by name)
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree test should have been deleted")
		}

		// Verify local directory still exists (we deleted the worktree, not the local dir)
		if _, err := os.Stat(localDir); os.IsNotExist(err) {
			t.Error("local directory test should still exist")
		}
	})
}
