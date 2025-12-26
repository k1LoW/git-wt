package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/git-wt/testutil"
)

// buildBinary builds git-wt binary for testing and returns the path.
func buildBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "git-wt")

	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	return binPath
}

// runGitWt runs git-wt command and returns stdout.
func runGitWt(t *testing.T, binPath, dir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// getWorktreePath extracts the worktree path from git-wt output.
// The path is the last line of output (after git messages).
func getWorktreePath(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

func TestE2E_ListWorktrees(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	out, err := runGitWt(t, binPath, repo.Root)
	if err != nil {
		t.Fatalf("git-wt failed: %v\noutput: %s", err, out)
	}

	// Should contain the main worktree
	if !strings.Contains(out, repo.Root) {
		t.Errorf("output should contain repo root %q, got: %s", repo.Root, out)
	}

	if !strings.Contains(out, "main") {
		t.Errorf("output should contain 'main' branch, got: %s", out)
	}
}

func TestE2E_CreateWorktree(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create new worktree with new branch
	out, err := runGitWt(t, binPath, repo.Root, "feature-branch")
	if err != nil {
		t.Fatalf("git-wt feature-branch failed: %v\noutput: %s", err, out)
	}

	// Output should be the path to the new worktree
	if !strings.Contains(out, "feature-branch") {
		t.Errorf("output should contain worktree path with 'feature-branch', got: %s", out)
	}

	// Extract the actual path (last line of output)
	wtPath := getWorktreePath(out)

	// Verify the directory was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("worktree directory was not created at %s", wtPath)
	}
}

func TestE2E_SwitchWorktree(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree
	out1, err := runGitWt(t, binPath, repo.Root, "feature")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\noutput: %s", err, out1)
	}
	wtPath := getWorktreePath(out1)

	// Running again should return the same path (switch behavior)
	out2, err := runGitWt(t, binPath, repo.Root, "feature")
	if err != nil {
		t.Fatalf("git-wt feature failed: %v\noutput: %s", err, out2)
	}

	// Second run should only output the path (no git messages since worktree exists)
	if out2 != wtPath {
		t.Errorf("expected same path %q, got %q", wtPath, out2)
	}
}

func TestE2E_DeleteWorktree(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree
	out, err := runGitWt(t, binPath, repo.Root, "to-delete")
	if err != nil {
		t.Fatalf("failed to create worktree: %v", err)
	}
	wtPath := getWorktreePath(out)

	// Verify it exists
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatalf("worktree should exist at %s", wtPath)
	}

	// Delete worktree
	out, err = runGitWt(t, binPath, repo.Root, "-d", "to-delete")
	if err != nil {
		t.Fatalf("git-wt -d failed: %v\noutput: %s", err, out)
	}

	// Verify worktree was deleted
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree should have been deleted")
	}
}

func TestE2E_ForceDeleteWorktree(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree with unmerged changes
	out, err := runGitWt(t, binPath, repo.Root, "unmerged")
	if err != nil {
		t.Fatalf("failed to create worktree: %v", err)
	}
	wtPath := getWorktreePath(out)

	// Make a commit in the worktree (will be unmerged)
	if err := os.WriteFile(filepath.Join(wtPath, "new.txt"), []byte("content"), 0600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = wtPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "unmerged commit")
	cmd.Dir = wtPath
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Force delete worktree
	out, err = runGitWt(t, binPath, repo.Root, "-D", "unmerged")
	if err != nil {
		t.Fatalf("git-wt -D failed: %v\noutput: %s", err, out)
	}

	// Verify worktree was deleted
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("worktree should have been force deleted")
	}
}

func TestE2E_InitScript(t *testing.T) {
	binPath := buildBinary(t)

	tests := []struct {
		shell    string
		contains []string
	}{
		{
			shell:    "bash",
			contains: []string{"# git-wt shell hook for bash", "_git_wt()"},
		},
		{
			shell:    "zsh",
			contains: []string{"# git-wt shell hook for zsh", "_git_wt()"},
		},
		{
			shell:    "fish",
			contains: []string{"# git-wt shell hook for fish", "function git --wraps git"},
		},
		{
			shell:    "powershell",
			contains: []string{"# git-wt shell hook for PowerShell", "Invoke-Git"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			// Can run from any directory for --init
			out, err := runGitWt(t, binPath, t.TempDir(), "--init", tt.shell)
			if err != nil {
				t.Fatalf("git-wt --init %s failed: %v\noutput: %s", tt.shell, err, out)
			}

			for _, s := range tt.contains {
				if !strings.Contains(out, s) {
					t.Errorf("output should contain %q, got: %s", s, out)
				}
			}
		})
	}
}

func TestE2E_InitScript_NoSwitchDirectory(t *testing.T) {
	binPath := buildBinary(t)

	out, err := runGitWt(t, binPath, t.TempDir(), "--init", "bash", "--no-switch-directory")
	if err != nil {
		t.Fatalf("git-wt --init bash --no-switch-directory failed: %v\noutput: %s", err, out)
	}

	// Should not contain the git wrapper function
	if strings.Contains(out, "git() {") {
		t.Error("output should not contain git wrapper when --no-switch-directory is used")
	}

	// Should still contain completion
	if !strings.Contains(out, "_git_wt()") {
		t.Error("output should contain completion function")
	}
}

func TestE2E_InitScript_UnsupportedShell(t *testing.T) {
	binPath := buildBinary(t)

	_, err := runGitWt(t, binPath, t.TempDir(), "--init", "unsupported")
	if err == nil {
		t.Error("expected error for unsupported shell")
	}
}

func TestE2E_WorktreeFromWorktree(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create first worktree
	out1, err := runGitWt(t, binPath, repo.Root, "feature1")
	if err != nil {
		t.Fatalf("failed to create first worktree: %v", err)
	}
	wt1Path := getWorktreePath(out1)

	// Create second worktree from within first worktree
	out2, err := runGitWt(t, binPath, wt1Path, "feature2")
	if err != nil {
		t.Fatalf("failed to create second worktree from worktree: %v\noutput: %s", err, out2)
	}
	wt2Path := getWorktreePath(out2)

	// Verify second worktree was created
	if _, err := os.Stat(wt2Path); os.IsNotExist(err) {
		t.Errorf("second worktree was not created at %s", wt2Path)
	}
}

func TestE2E_CopyIgnored(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n")
	repo.Commit("initial commit")

	// Create ignored file
	repo.CreateFile(".env", "SECRET=supersecret")

	// Enable copyignored
	repo.Git("config", "wt.copyignored", "true")

	// Create worktree
	out, err := runGitWt(t, binPath, repo.Root, "with-env")
	if err != nil {
		t.Fatalf("failed to create worktree: %v", err)
	}
	wtPath := getWorktreePath(out)

	// Verify .env was copied
	envPath := filepath.Join(wtPath, ".env")
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf(".env was not copied to worktree: %v", err)
	}

	if string(content) != "SECRET=supersecret" {
		t.Errorf(".env content = %q, want %q", string(content), "SECRET=supersecret")
	}
}

func TestE2E_CustomBaseDir(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Set custom basedir
	customBase := filepath.Join(repo.ParentDir(), "custom-wt-dir")
	repo.Git("config", "wt.basedir", customBase)

	// Create worktree
	out, err := runGitWt(t, binPath, repo.Root, "custom-branch")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
	}
	wtPath := getWorktreePath(out)

	// Verify worktree was created in custom location
	expectedPath := filepath.Join(customBase, "custom-branch")
	if wtPath != expectedPath {
		t.Errorf("worktree path = %q, want %q", wtPath, expectedPath)
	}

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("worktree was not created at custom path %s", wtPath)
	}
}

func TestE2E_Version(t *testing.T) {
	binPath := buildBinary(t)

	out, err := runGitWt(t, binPath, t.TempDir(), "--version")
	if err != nil {
		t.Fatalf("git-wt --version failed: %v\noutput: %s", err, out)
	}

	// Version output format: "git version X.Y.Z" (from cobra)
	if !strings.Contains(out, "version") {
		t.Errorf("expected version output, got: %s", out)
	}
}

func TestE2E_Help(t *testing.T) {
	binPath := buildBinary(t)

	out, err := runGitWt(t, binPath, t.TempDir(), "--help")
	if err != nil {
		t.Fatalf("git-wt --help failed: %v\noutput: %s", err, out)
	}

	expectedStrings := []string{
		"git wt [branch]",
		"--delete",
		"--force-delete",
		"--init",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(out, s) {
			t.Errorf("help output should contain %q", s)
		}
	}
}

func TestE2E_ExistingBranch(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a branch without worktree
	repo.Git("branch", "existing-branch")

	// Create worktree for existing branch
	out, err := runGitWt(t, binPath, repo.Root, "existing-branch")
	if err != nil {
		t.Fatalf("failed to create worktree for existing branch: %v\noutput: %s", err, out)
	}
	wtPath := getWorktreePath(out)

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("worktree was not created at %s", wtPath)
	}

	// Verify it's using the existing branch (not creating a new one)
	// The branch should still exist after worktree creation
	restore := repo.Chdir()
	defer restore()

	cmd := exec.Command("git", "branch", "--list", "existing-branch")
	branchOut, err := cmd.Output()
	if err != nil {
		t.Fatalf("git branch --list failed: %v", err)
	}

	if !strings.Contains(string(branchOut), "existing-branch") {
		t.Error("existing-branch should still exist")
	}
}
