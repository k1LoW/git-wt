package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/k1LoW/exec"

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

// runGitWt runs git-wt command and returns combined output (stdout + stderr).
func runGitWt(t *testing.T, binPath, dir string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// runGitWtStdout runs git-wt command and returns stdout only.
// This is important for shell integration tests where only stdout is captured.
func runGitWtStdout(t *testing.T, binPath, dir string, args ...string) (stdout string, stderr string, err error) {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	return strings.TrimSpace(stdoutBuf.String()), strings.TrimSpace(stderrBuf.String()), err
}

// worktreePath extracts the worktree path from git-wt output.
// The path is the last line of output (after git messages).
func worktreePath(output string) string {
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

// TestE2E_ListWorktrees_TableFormat tests that git-wt table output is properly formatted
// with newlines preserved (regression test for fish shell hook issue, ref: PR #14).
func TestE2E_ListWorktrees_TableFormat(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create multiple worktrees to ensure table has multiple rows
	_, err := runGitWt(t, binPath, repo.Root, "feature-a")
	if err != nil {
		t.Fatalf("failed to create worktree feature-a: %v", err)
	}
	_, err = runGitWt(t, binPath, repo.Root, "feature-b")
	if err != nil {
		t.Fatalf("failed to create worktree feature-b: %v", err)
	}

	// Run git wt (list worktrees)
	out, err := runGitWt(t, binPath, repo.Root)
	if err != nil {
		t.Fatalf("git-wt failed: %v\noutput: %s", err, out)
	}

	// Table output should have multiple lines (header + at least 3 worktrees)
	lines := strings.Split(out, "\n")
	if len(lines) < 4 {
		t.Errorf("table output should have at least 4 lines (header + 3 worktrees), got %d lines:\n%s", len(lines), out)
	}

	// Each worktree should be on its own line (not collapsed into one line)
	// Check that main, feature-a, and feature-b are on separate lines
	var mainLine, featureALine, featureBLine string
	for _, line := range lines {
		if strings.Contains(line, "main") {
			mainLine = line
		}
		if strings.Contains(line, "feature-a") {
			featureALine = line
		}
		if strings.Contains(line, "feature-b") {
			featureBLine = line
		}
	}

	if mainLine == "" {
		t.Error("main branch should be on its own line")
	}
	if featureALine == "" {
		t.Error("feature-a branch should be on its own line")
	}
	if featureBLine == "" {
		t.Error("feature-b branch should be on its own line")
	}

	// Ensure they are on different lines (not all on the same line)
	if mainLine == featureALine || mainLine == featureBLine || featureALine == featureBLine {
		t.Errorf("each worktree should be on its own line, but found duplicates:\nmain: %q\nfeature-a: %q\nfeature-b: %q", mainLine, featureALine, featureBLine)
	}
}

// verifyTableFormat checks that table output has proper newline formatting.
// Returns error if the output appears to be collapsed into a single line.
func verifyTableFormat(t *testing.T, output string) {
	t.Helper()

	lines := strings.Split(output, "\n")

	// Should have at least 4 lines (header + 3 worktrees)
	if len(lines) < 4 {
		t.Errorf("table output should have at least 4 lines, got %d lines:\n%s", len(lines), output)
	}

	// Each worktree should be on its own line
	var mainLine, featureALine, featureBLine string
	for _, line := range lines {
		if strings.Contains(line, "main") {
			mainLine = line
		}
		if strings.Contains(line, "feature-a") {
			featureALine = line
		}
		if strings.Contains(line, "feature-b") {
			featureBLine = line
		}
	}

	if mainLine == "" {
		t.Error("main branch should be on its own line")
	}
	if featureALine == "" {
		t.Error("feature-a branch should be on its own line")
	}
	if featureBLine == "" {
		t.Error("feature-b branch should be on its own line")
	}

	// Ensure they are on different lines
	if mainLine == featureALine || mainLine == featureBLine || featureALine == featureBLine {
		t.Errorf("each worktree should be on its own line, but found duplicates:\nmain: %q\nfeature-a: %q\nfeature-b: %q", mainLine, featureALine, featureBLine)
	}
}

// TestE2E_ListWorktrees_TableFormat_Shell tests table output formatting via shell integration.
// Regression test for PR #14 which fixed fish hook output formatting.
func TestE2E_ListWorktrees_TableFormat_Shell(t *testing.T) {
	binPath := buildBinary(t)

	tests := []struct {
		name       string
		shell      string
		scriptFunc func(repoRoot, pathDir, binPath string) string
	}{
		{
			name:  "bash",
			shell: "bash",
			scriptFunc: func(repoRoot, pathDir, binPath string) string {
				return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"
git wt
`, repoRoot, pathDir)
			},
		},
		{
			name:  "zsh",
			shell: "zsh",
			scriptFunc: func(repoRoot, pathDir, binPath string) string {
				return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"
git wt
`, repoRoot, pathDir)
			},
		},
		{
			name:  "fish",
			shell: "fish",
			scriptFunc: func(repoRoot, pathDir, binPath string) string {
				return fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source
git wt
`, repoRoot, pathDir)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exec.LookPath(tt.shell); err != nil {
				t.Skipf("%s not available", tt.shell)
			}

			repo := testutil.NewTestRepo(t)
			repo.CreateFile("README.md", "# Test")
			repo.Commit("initial commit")

			// Create worktrees first (without shell integration)
			_, err := runGitWt(t, binPath, repo.Root, "feature-a")
			if err != nil {
				t.Fatalf("failed to create worktree feature-a: %v", err)
			}
			_, err = runGitWt(t, binPath, repo.Root, "feature-b")
			if err != nil {
				t.Fatalf("failed to create worktree feature-b: %v", err)
			}

			script := tt.scriptFunc(repo.Root, filepath.Dir(binPath), binPath)
			cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s shell integration failed: %v\noutput: %s", tt.shell, err, out)
			}

			verifyTableFormat(t, string(out))
		})
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
	wtPath := worktreePath(out)

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
	wtPath := worktreePath(out1)

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
	wtPath := worktreePath(out)

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
	wtPath := worktreePath(out)

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
			contains: []string{"# git-wt shell hook for zsh", "_git-wt()"},
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

func TestE2E_InitScript_Nocd(t *testing.T) {
	binPath := buildBinary(t)

	out, err := runGitWt(t, binPath, t.TempDir(), "--init", "bash", "--nocd")
	if err != nil {
		t.Fatalf("git-wt --init bash --nocd failed: %v\noutput: %s", err, out)
	}

	// Should not contain the git wrapper function
	if strings.Contains(out, "git() {") {
		t.Error("output should not contain git wrapper when --nocd is used")
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

	// Set local basedir to avoid global config influence
	repo.Git("config", "wt.basedir", "../{gitroot}-wt")

	// Create first worktree
	out1, err := runGitWt(t, binPath, repo.Root, "feature1")
	if err != nil {
		t.Fatalf("failed to create first worktree: %v", err)
	}
	wt1Path := worktreePath(out1)

	// Create second worktree from within first worktree
	out2, err := runGitWt(t, binPath, wt1Path, "feature2")
	if err != nil {
		t.Fatalf("failed to create second worktree from worktree: %v\noutput: %s", err, out2)
	}
	wt2Path := worktreePath(out2)

	// Verify second worktree was created
	if _, err := os.Stat(wt2Path); os.IsNotExist(err) {
		t.Errorf("second worktree was not created at %s", wt2Path)
	}

	// Verify worktree paths are created under the same basedir
	expectedWt1Path := filepath.Clean(filepath.Join(repo.Root, "../repo-wt/feature1"))
	if wt1Path != expectedWt1Path {
		t.Errorf("first worktree path = %q, want %q", wt1Path, expectedWt1Path)
	}

	expectedWt2Path := filepath.Clean(filepath.Join(repo.Root, "../repo-wt/feature2"))
	if wt2Path != expectedWt2Path {
		t.Errorf("second worktree path = %q, want %q", wt2Path, expectedWt2Path)
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
	wtPath := worktreePath(out)

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
	wtPath := worktreePath(out)

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
		"git wt [branch|worktree] [start-point]",
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
	wtPath := worktreePath(out)

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

// TestE2E_ShellIntegration_StdoutFormat tests that git-wt output is compatible
// with shell integration (stdout contains only the path, suitable for cd).
func TestE2E_ShellIntegration_StdoutFormat(t *testing.T) {
	binPath := buildBinary(t)

	t.Run("list_worktrees_stdout_is_not_directory", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		stdout, _, err := runGitWtStdout(t, binPath, repo.Root)
		if err != nil {
			t.Fatalf("git-wt failed: %v", err)
		}

		// List output should NOT be a valid directory path
		// (it's a table, so shell integration should not cd)
		info, err := os.Stat(stdout)
		if err == nil && info.IsDir() {
			t.Errorf("list output should not be a valid directory, got: %s", stdout)
		}
	})

	t.Run("create_worktree_stdout_is_directory", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "feature-shell")
		if err != nil {
			t.Fatalf("git-wt feature-shell failed: %v\nstderr: %s", err, stderr)
		}

		// stdout should be exactly one line (the path)
		lines := strings.Split(stdout, "\n")
		if len(lines) != 1 {
			t.Errorf("stdout should be exactly 1 line, got %d lines: %q", len(lines), stdout)
		}

		// stdout should be a valid directory
		info, err := os.Stat(stdout)
		if err != nil {
			t.Errorf("stdout path does not exist: %v", err)
		} else if !info.IsDir() {
			t.Errorf("stdout should be a directory, got: %s", stdout)
		}

		// stderr should contain git messages (not empty for new worktree)
		if stderr == "" {
			t.Log("warning: stderr is empty (git worktree add usually outputs to stderr)")
		}
	})

	t.Run("switch_worktree_stdout_is_directory", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree first
		_, _, err := runGitWtStdout(t, binPath, repo.Root, "existing-wt")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Switch to existing worktree
		stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "existing-wt")
		if err != nil {
			t.Fatalf("git-wt existing-wt failed: %v\nstderr: %s", err, stderr)
		}

		// stdout should be exactly one line
		lines := strings.Split(stdout, "\n")
		if len(lines) != 1 {
			t.Errorf("stdout should be exactly 1 line, got %d lines: %q", len(lines), stdout)
		}

		// stdout should be a valid directory
		info, err := os.Stat(stdout)
		if err != nil {
			t.Errorf("stdout path does not exist: %v", err)
		} else if !info.IsDir() {
			t.Errorf("stdout should be a directory, got: %s", stdout)
		}

		// stderr should be empty for existing worktree (no git operation)
		if stderr != "" {
			t.Logf("note: stderr is not empty for existing worktree: %s", stderr)
		}
	})

	t.Run("delete_worktree_stdout_is_not_directory", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree first
		_, _, err := runGitWtStdout(t, binPath, repo.Root, "to-delete-shell")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		// Delete worktree
		stdout, _, err := runGitWtStdout(t, binPath, repo.Root, "-d", "to-delete-shell")
		if err != nil {
			t.Fatalf("git-wt -d failed: %v", err)
		}

		// Delete output should NOT be a valid directory
		// (it's a message, so shell integration should not cd)
		info, err := os.Stat(stdout)
		if err == nil && info.IsDir() {
			t.Errorf("delete output should not be a valid directory, got: %s", stdout)
		}
	})
}

// TestE2E_ShellIntegration_Bash tests the actual shell integration with bash.
func TestE2E_ShellIntegration_Bash(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	binPath := buildBinary(t)
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Test that shell integration works: eval the init script and run git wt
	script := fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"

# Test: git wt <branch> should cd to the worktree
git wt shell-bash-test
pwd
`, repo.Root, filepath.Dir(binPath))

	cmd := exec.Command("bash", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("bash shell integration failed: %v\noutput: %s", err, out)
	}

	output := strings.TrimSpace(string(out))
	// The last line should be the worktree path
	lines := strings.Split(output, "\n")
	pwd := lines[len(lines)-1]

	if !strings.Contains(pwd, "shell-bash-test") {
		t.Errorf("pwd should contain worktree path, got: %s", pwd)
	}
}

// TestE2E_ShellIntegration_Zsh tests the actual shell integration with zsh.
func TestE2E_ShellIntegration_Zsh(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not available")
	}

	binPath := buildBinary(t)
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	script := fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"

# Test: git wt <branch> should cd to the worktree
git wt shell-zsh-test
pwd
`, repo.Root, filepath.Dir(binPath))

	cmd := exec.Command("zsh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("zsh shell integration failed: %v\noutput: %s", err, out)
	}

	output := strings.TrimSpace(string(out))
	lines := strings.Split(output, "\n")
	pwd := lines[len(lines)-1]

	if !strings.Contains(pwd, "shell-zsh-test") {
		t.Errorf("pwd should contain worktree path, got: %s", pwd)
	}
}

// TestE2E_ShellIntegration_Fish tests the actual shell integration with fish.
func TestE2E_ShellIntegration_Fish(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish not available")
	}

	binPath := buildBinary(t)
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	script := fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source

# Test: git wt <branch> should cd to the worktree
git wt shell-fish-test
pwd
`, repo.Root, filepath.Dir(binPath))

	cmd := exec.Command("fish", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("fish shell integration failed: %v\noutput: %s", err, out)
	}

	output := strings.TrimSpace(string(out))
	lines := strings.Split(output, "\n")
	pwd := lines[len(lines)-1]

	if !strings.Contains(pwd, "shell-fish-test") {
		t.Errorf("pwd should contain worktree path, got: %s", pwd)
	}
}

// TestE2E_BasedirFlag tests the --basedir flag overrides wt.basedir config.
func TestE2E_BasedirFlag(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Set config basedir to one location
	configBase := filepath.Join(repo.ParentDir(), "config-wt-dir")
	repo.Git("config", "wt.basedir", configBase)

	// Use flag to override to different location
	flagBase := filepath.Join(repo.ParentDir(), "flag-wt-dir")

	// Create worktree with --basedir flag
	out, err := runGitWt(t, binPath, repo.Root, "--basedir", flagBase, "flag-branch")
	if err != nil {
		t.Fatalf("failed to create worktree with --basedir flag: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify worktree was created in flag location, not config location
	expectedPath := filepath.Join(flagBase, "flag-branch")
	if wtPath != expectedPath {
		t.Errorf("worktree path = %q, want %q", wtPath, expectedPath)
	}

	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Errorf("worktree was not created at flag path %s", wtPath)
	}

	// Verify it was NOT created in config location
	configPath := filepath.Join(configBase, "flag-branch")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Errorf("worktree should not have been created at config path %s", configPath)
	}
}

// TestE2E_CopyignoredFlag tests the --copyignored flag overrides wt.copyignored config.
func TestE2E_CopyignoredFlag(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n")
	repo.Commit("initial commit")

	// Create ignored file
	repo.CreateFile(".env", "SECRET=flagtest")

	// Config is NOT set (default false)
	// Use --copyignored flag

	out, err := runGitWt(t, binPath, repo.Root, "--copyignored", "copyignored-flag-test")
	if err != nil {
		t.Fatalf("failed to create worktree with --copyignored flag: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify .env was copied
	envPath := filepath.Join(wtPath, ".env")
	content, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf(".env was not copied to worktree despite --copyignored flag: %v", err)
	}

	if string(content) != "SECRET=flagtest" {
		t.Errorf(".env content = %q, want %q", string(content), "SECRET=flagtest")
	}
}

// TestE2E_CopyuntrackedFlag tests the --copyuntracked flag.
func TestE2E_CopyuntrackedFlag(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create untracked file
	repo.CreateFile("untracked.txt", "untracked-flag-content")

	// Use --copyuntracked flag
	out, err := runGitWt(t, binPath, repo.Root, "--copyuntracked", "copyuntracked-flag-test")
	if err != nil {
		t.Fatalf("failed to create worktree with --copyuntracked flag: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify untracked file was copied
	untrackedPath := filepath.Join(wtPath, "untracked.txt")
	content, err := os.ReadFile(untrackedPath)
	if err != nil {
		t.Fatalf("untracked.txt was not copied to worktree despite --copyuntracked flag: %v", err)
	}

	if string(content) != "untracked-flag-content" {
		t.Errorf("untracked.txt content = %q, want %q", string(content), "untracked-flag-content")
	}
}

// TestE2E_CopymodifiedFlag tests the --copymodified flag.
func TestE2E_CopymodifiedFlag(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile("tracked.txt", "original-content")
	repo.Commit("initial commit")

	// Modify tracked file
	repo.CreateFile("tracked.txt", "modified-flag-content")

	// Use --copymodified flag
	out, err := runGitWt(t, binPath, repo.Root, "--copymodified", "copymodified-flag-test")
	if err != nil {
		t.Fatalf("failed to create worktree with --copymodified flag: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify modified file was copied with modifications
	trackedPath := filepath.Join(wtPath, "tracked.txt")
	content, err := os.ReadFile(trackedPath)
	if err != nil {
		t.Fatalf("tracked.txt was not copied to worktree: %v", err)
	}

	if string(content) != "modified-flag-content" {
		t.Errorf("tracked.txt content = %q, want %q", string(content), "modified-flag-content")
	}
}

// TestE2E_MultipleCopyFlags tests combining multiple copy flags.
func TestE2E_MultipleCopyFlags(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n")
	repo.CreateFile("tracked.txt", "original")
	repo.Commit("initial commit")

	// Create various files
	repo.CreateFile(".env", "SECRET=multi")
	repo.CreateFile("untracked.txt", "untracked-multi")
	repo.CreateFile("tracked.txt", "modified-multi")

	// Use multiple flags
	out, err := runGitWt(t, binPath, repo.Root, "--copyignored", "--copyuntracked", "--copymodified", "multi-flag-test")
	if err != nil {
		t.Fatalf("failed to create worktree with multiple flags: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify all files were copied
	tests := []struct {
		file    string
		content string
	}{
		{".env", "SECRET=multi"},
		{"untracked.txt", "untracked-multi"},
		{"tracked.txt", "modified-multi"},
	}

	for _, tt := range tests {
		path := filepath.Join(wtPath, tt.file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s was not copied to worktree: %v", tt.file, err)
			continue
		}
		if string(content) != tt.content {
			t.Errorf("%s content = %q, want %q", tt.file, string(content), tt.content)
		}
	}
}

// TestE2E_FlagOverridesConfig tests that flags override config values.
func TestE2E_FlagOverridesConfig(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n")
	repo.Commit("initial commit")

	// Set copyignored to false in config
	repo.Git("config", "wt.copyignored", "false")

	// Create ignored file
	repo.CreateFile(".env", "SECRET=override")

	// Without flag, .env should NOT be copied
	out1, err := runGitWt(t, binPath, repo.Root, "no-flag-test")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\noutput: %s", err, out1)
	}
	wtPath1 := worktreePath(out1)

	envPath1 := filepath.Join(wtPath1, ".env")
	if _, err := os.Stat(envPath1); !os.IsNotExist(err) {
		t.Error(".env should NOT have been copied without --copyignored flag")
	}

	// With flag, .env SHOULD be copied
	out2, err := runGitWt(t, binPath, repo.Root, "--copyignored", "with-flag-test")
	if err != nil {
		t.Fatalf("failed to create worktree with flag: %v\noutput: %s", err, out2)
	}
	wtPath2 := worktreePath(out2)

	envPath2 := filepath.Join(wtPath2, ".env")
	if _, err := os.Stat(envPath2); os.IsNotExist(err) {
		t.Error(".env SHOULD have been copied with --copyignored flag")
	}
}

// TestE2E_ShellIntegration_Nocd tests that --nocd flag
// prevents cd when used with git wt <branch> via shell integration.
func TestE2E_ShellIntegration_Nocd(t *testing.T) {
	binPath := buildBinary(t)

	tests := []struct {
		name       string
		shell      string
		scriptFunc func(repoRoot, pathDir string) string
	}{
		{
			name:  "bash",
			shell: "bash",
			scriptFunc: func(repoRoot, pathDir string) string {
				return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"

# Test: git wt --nocd <branch> should NOT cd to the worktree
git wt --nocd nocd-bash-test
pwd
`, repoRoot, pathDir)
			},
		},
		{
			name:  "zsh",
			shell: "zsh",
			scriptFunc: func(repoRoot, pathDir string) string {
				return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"

# Test: git wt --nocd <branch> should NOT cd to the worktree
git wt --nocd nocd-zsh-test
pwd
`, repoRoot, pathDir)
			},
		},
		{
			name:  "fish",
			shell: "fish",
			scriptFunc: func(repoRoot, pathDir string) string {
				return fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source

# Test: git wt --nocd <branch> should NOT cd to the worktree
git wt --nocd nocd-fish-test
pwd
`, repoRoot, pathDir)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exec.LookPath(tt.shell); err != nil {
				t.Skipf("%s not available", tt.shell)
			}

			repo := testutil.NewTestRepo(t)
			repo.CreateFile("README.md", "# Test")
			repo.Commit("initial commit")

			script := tt.scriptFunc(repo.Root, filepath.Dir(binPath))
			cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s shell integration with --nocd failed: %v\noutput: %s", tt.shell, err, out)
			}

			output := strings.TrimSpace(string(out))
			lines := strings.Split(output, "\n")
			pwd := lines[len(lines)-1]

			// pwd should be the original repo root, NOT the new worktree
			if strings.Contains(pwd, "nocd-"+tt.name+"-test") {
				t.Errorf("pwd should NOT contain worktree path when --nocd is used, got: %s", pwd)
			}
			if pwd != repo.Root {
				t.Errorf("pwd should be original repo root %q, got: %s", repo.Root, pwd)
			}
		})
	}
}

// TestE2E_NocdConfig tests that wt.nocd config prevents cd to the worktree.
func TestE2E_NocdConfig(t *testing.T) {
	binPath := buildBinary(t)

	tests := []struct {
		name       string
		shell      string
		scriptFunc func(repoRoot, pathDir string) string
	}{
		{
			name:  "bash",
			shell: "bash",
			scriptFunc: func(repoRoot, pathDir string) string {
				return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init bash)"

# Test: git wt <branch> with wt.nocd=true should NOT cd to the worktree
git wt nocd-config-bash-test
pwd
`, repoRoot, pathDir)
			},
		},
		{
			name:  "zsh",
			shell: "zsh",
			scriptFunc: func(repoRoot, pathDir string) string {
				return fmt.Sprintf(`
set -e
cd %q
export PATH="%s:$PATH"
eval "$(git wt --init zsh)"

# Test: git wt <branch> with wt.nocd=true should NOT cd to the worktree
git wt nocd-config-zsh-test
pwd
`, repoRoot, pathDir)
			},
		},
		{
			name:  "fish",
			shell: "fish",
			scriptFunc: func(repoRoot, pathDir string) string {
				return fmt.Sprintf(`
cd %q
set -x PATH %s $PATH
git wt --init fish | source

# Test: git wt <branch> with wt.nocd=true should NOT cd to the worktree
git wt nocd-config-fish-test
pwd
`, repoRoot, pathDir)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := exec.LookPath(tt.shell); err != nil {
				t.Skipf("%s not available", tt.shell)
			}

			repo := testutil.NewTestRepo(t)
			repo.CreateFile("README.md", "# Test")
			repo.Commit("initial commit")

			// Set wt.nocd in config
			repo.Git("config", "wt.nocd", "true")

			script := tt.scriptFunc(repo.Root, filepath.Dir(binPath))
			cmd := exec.Command(tt.shell, "-c", script) //#nosec G204
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s shell integration with wt.nocd config failed: %v\noutput: %s", tt.shell, err, out)
			}

			output := strings.TrimSpace(string(out))
			lines := strings.Split(output, "\n")
			pwd := lines[len(lines)-1]

			// With wt.nocd=true config, pwd should remain in original repo root
			if strings.Contains(pwd, "nocd-config-"+tt.name+"-test") {
				t.Errorf("pwd should NOT contain worktree path when wt.nocd=true, got: %s", pwd)
			}
			if pwd != repo.Root {
				t.Errorf("pwd should be original repo root %q, got: %s", repo.Root, pwd)
			}
		})
	}
}

// TestE2E_NocdConfigWithInit tests that --init ignores wt.nocd config and always outputs git() wrapper.
// The wt.nocd config only affects cd behavior at runtime, not the init output.
func TestE2E_NocdConfigWithInit(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Set wt.nocd in config
	repo.Git("config", "wt.nocd", "true")

	restore := repo.Chdir()
	defer restore()

	out, err := runGitWt(t, binPath, repo.Root, "--init", "bash")
	if err != nil {
		t.Fatalf("git-wt --init bash failed: %v\noutput: %s", err, out)
	}

	// With wt.nocd=true config, --init should still output git wrapper
	// (wt.nocd only affects runtime cd behavior, not init output)
	if !strings.Contains(out, "git() {") {
		t.Error("output should contain git wrapper even when wt.nocd=true config is set")
	}
}

// TestE2E_HookFlag tests the --hook flag runs commands after creating a new worktree.
func TestE2E_HookFlag(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree with --hook flag that creates a marker file
	out, err := runGitWt(t, binPath, repo.Root, "--hook", "touch hook-marker.txt", "hook-test")
	if err != nil {
		t.Fatalf("failed to create worktree with --hook flag: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify the hook created the marker file
	markerPath := filepath.Join(wtPath, "hook-marker.txt")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("hook-marker.txt was not created by hook")
	}
}

// TestE2E_HookConfig tests the wt.hook config runs commands after creating a new worktree.
func TestE2E_HookConfig(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Set hook in config
	repo.Git("config", "--add", "wt.hook", "touch config-hook-marker.txt")

	// Create worktree
	out, err := runGitWt(t, binPath, repo.Root, "hook-config-test")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify the hook created the marker file
	markerPath := filepath.Join(wtPath, "config-hook-marker.txt")
	if _, err := os.Stat(markerPath); os.IsNotExist(err) {
		t.Error("config-hook-marker.txt was not created by hook from config")
	}
}

// TestE2E_MultipleHooks tests that multiple hooks run in order.
func TestE2E_MultipleHooks(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree with multiple hooks
	out, err := runGitWt(t, binPath, repo.Root, "--hook", "echo first > order.txt", "--hook", "echo second >> order.txt", "multi-hook-test")
	if err != nil {
		t.Fatalf("failed to create worktree with multiple hooks: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify both hooks ran in order
	orderPath := filepath.Join(wtPath, "order.txt")
	content, err := os.ReadFile(orderPath)
	if err != nil {
		t.Fatalf("order.txt was not created: %v", err)
	}

	expected := "first\nsecond\n"
	if string(content) != expected {
		t.Errorf("order.txt content = %q, want %q", string(content), expected)
	}
}

// TestE2E_HookNotRunOnExistingWorktree tests that hooks do NOT run when switching to existing worktree.
func TestE2E_HookNotRunOnExistingWorktree(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree first (without hook)
	out, err := runGitWt(t, binPath, repo.Root, "existing-hook-test")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Switch to existing worktree with hook - hook should NOT run
	out2, err := runGitWt(t, binPath, repo.Root, "--hook", "touch should-not-exist.txt", "existing-hook-test")
	if err != nil {
		t.Fatalf("failed to switch to worktree: %v\noutput: %s", err, out2)
	}

	// Verify the hook did NOT create the file
	markerPath := filepath.Join(wtPath, "should-not-exist.txt")
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("hook should NOT have run when switching to existing worktree")
	}
}

// TestE2E_HookFlagOverridesConfig tests that --hook flag overrides wt.hook config.
func TestE2E_HookFlagOverridesConfig(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Set hook in config
	repo.Git("config", "--add", "wt.hook", "touch config-marker.txt")

	// Create worktree with --hook flag (should override config)
	out, err := runGitWt(t, binPath, repo.Root, "--hook", "touch flag-marker.txt", "hook-override-test")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\noutput: %s", err, out)
	}
	wtPath := worktreePath(out)

	// Verify flag hook ran
	flagMarkerPath := filepath.Join(wtPath, "flag-marker.txt")
	if _, err := os.Stat(flagMarkerPath); os.IsNotExist(err) {
		t.Error("flag-marker.txt should have been created by --hook flag")
	}

	// Verify config hook did NOT run (flag overrides config)
	configMarkerPath := filepath.Join(wtPath, "config-marker.txt")
	if _, err := os.Stat(configMarkerPath); !os.IsNotExist(err) {
		t.Error("config-marker.txt should NOT have been created (--hook flag overrides config)")
	}
}

// TestE2E_HookFailureExitsWithError tests that hook failures cause exit code 1 and stop execution.
func TestE2E_HookFailureExitsWithError(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree with a failing hook followed by a successful hook
	stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "--hook", "exit 1", "--hook", "touch after-failure.txt", "hook-failure-test")

	// Command should fail with exit code 1
	if err == nil {
		t.Fatal("command should fail when hook fails")
	}

	wtPath := strings.TrimSpace(stdout)

	// Verify stderr contains error about the failed hook
	if !strings.Contains(stderr, "hook") || !strings.Contains(stderr, "failed") {
		t.Errorf("stderr should contain error about failed hook, got: %s", stderr)
	}

	// Verify the second hook did NOT run (execution stops on first failure)
	markerPath := filepath.Join(wtPath, "after-failure.txt")
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Error("second hook should NOT have run after first hook failed")
	}

	// Verify worktree was still created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree should have been created even though hook failed")
	}
}

// TestE2E_HookOutputToStderr tests that hook output goes to stderr (not stdout).
func TestE2E_HookOutputToStderr(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create worktree with a hook that outputs to stdout
	stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "--hook", "echo hook-output-test", "hook-output-test")
	if err != nil {
		t.Fatalf("failed to create worktree: %v\nstderr: %s", err, stderr)
	}

	// stdout should only contain the worktree path (for shell integration)
	lines := strings.Split(stdout, "\n")
	if len(lines) != 1 {
		t.Errorf("stdout should be exactly 1 line (worktree path), got %d lines: %q", len(lines), stdout)
	}

	// Hook output should be in stderr
	if !strings.Contains(stderr, "hook-output-test") {
		t.Errorf("hook output should be in stderr, got stderr: %s", stderr)
	}
}

// TestE2E_Complete tests the __complete command output.
func TestE2E_Complete(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create some branches for completion
	repo.Git("branch", "feature-one")
	repo.Git("branch", "feature-two")

	t.Run("empty_input_returns_branches", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain branch names
		if !strings.Contains(out, "main") {
			t.Errorf("completion should contain 'main', got: %s", out)
		}
		if !strings.Contains(out, "feature-one") {
			t.Errorf("completion should contain 'feature-one', got: %s", out)
		}
		if !strings.Contains(out, "feature-two") {
			t.Errorf("completion should contain 'feature-two', got: %s", out)
		}
	})

	t.Run("partial_input_filters_branches", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "feat")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain matching branches
		if !strings.Contains(out, "feature-one") {
			t.Errorf("completion should contain 'feature-one', got: %s", out)
		}
		if !strings.Contains(out, "feature-two") {
			t.Errorf("completion should contain 'feature-two', got: %s", out)
		}
	})

	t.Run("dash_input_returns_flags", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "-")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain flags with descriptions (tab-separated)
		expectedFlags := []string{
			"--delete",
			"--force-delete",
			"--init",
			"--basedir",
			"--copyignored",
			"--copyuntracked",
			"--copymodified",
			"--hook",
			"--nocopy",
			"--nocd",
			"-d",
			"-D",
		}

		for _, flag := range expectedFlags {
			if !strings.Contains(out, flag) {
				t.Errorf("completion should contain %q, got: %s", flag, out)
			}
		}

		// Verify tab-separated format (flag\tdescription)
		lines := strings.Split(out, "\n")
		hasTabSeparated := false
		for _, line := range lines {
			if strings.Contains(line, "\t") && strings.HasPrefix(line, "-") {
				hasTabSeparated = true
				break
			}
		}
		if !hasTabSeparated {
			t.Errorf("completion should have tab-separated descriptions, got: %s", out)
		}
	})

	t.Run("double_dash_input_returns_long_flags", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "--")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain long flags
		if !strings.Contains(out, "--delete") {
			t.Errorf("completion should contain '--delete', got: %s", out)
		}
		if !strings.Contains(out, "--basedir") {
			t.Errorf("completion should contain '--basedir', got: %s", out)
		}
	})

	t.Run("branch_completion_has_description", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Branches without worktree should have [branch] description
		lines := strings.Split(out, "\n")
		hasBranchDesc := false
		for _, line := range lines {
			if strings.Contains(line, "[branch]") {
				hasBranchDesc = true
				break
			}
		}
		if !hasBranchDesc {
			t.Errorf("branch completion should have [branch] description, got: %s", out)
		}

		// main branch (worktree) should have [branch: worktree=...] description
		hasWorktreeDesc := false
		for _, line := range lines {
			if strings.HasPrefix(line, "main\t") && strings.Contains(line, "[branch: worktree=") {
				hasWorktreeDesc = true
				break
			}
		}
		if !hasWorktreeDesc {
			t.Errorf("main branch should have [branch: worktree=...] description, got: %s", out)
		}
	})

	t.Run("branch_completion_has_commit_message", func(t *testing.T) {
		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should contain "initial commit" from the commit message
		if !strings.Contains(out, "initial commit") {
			t.Errorf("branch completion should include commit message, got: %s", out)
		}
	})

	t.Run("worktree_completion_has_description", func(t *testing.T) {
		// Create a worktree - the directory name will be the same as branch name
		_, err := runGitWt(t, binPath, repo.Root, "wt-test-branch")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		out, err := runGitWt(t, binPath, repo.Root, "__complete", "")
		if err != nil {
			t.Fatalf("__complete failed: %v\noutput: %s", err, out)
		}

		// Should have worktree directory with [worktree: branch=...] description
		// Note: when branch name equals directory name, both entries exist
		lines := strings.Split(out, "\n")
		hasWorktreeDirDesc := false
		for _, line := range lines {
			// Check for [worktree: branch=...] (directory entry) or [branch: worktree=...] (branch entry)
			if strings.Contains(line, "[worktree: branch=") || strings.Contains(line, "[branch: worktree=") {
				hasWorktreeDirDesc = true
				break
			}
		}
		if !hasWorktreeDirDesc {
			t.Errorf("worktree completion should have description with worktree info, got: %s", out)
		}
	})
}

// TestE2E_CreateWorktreeWithStartPoint tests creating a worktree from a specific start-point.
func TestE2E_CreateWorktreeWithStartPoint(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a commit on main
	repo.CreateFile("main-file.txt", "main content")
	repo.Commit("main commit")

	// Create a branch from the first commit
	repo.Git("branch", "old-base", "HEAD~1")

	// Create worktree from old-base start-point
	stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "feature-from-old", "old-base")
	if err != nil {
		t.Fatalf("git-wt with start-point failed: %v\nstderr: %s", err, stderr)
	}

	wtPath := strings.TrimSpace(stdout)

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatalf("worktree was not created at %s", wtPath)
	}

	// Verify the worktree is based on old-base (should NOT have main-file.txt)
	mainFilePath := filepath.Join(wtPath, "main-file.txt")
	if _, err := os.Stat(mainFilePath); !os.IsNotExist(err) {
		t.Error("worktree should NOT have main-file.txt (should be based on old-base, not main)")
	}
}

// TestE2E_CreateWorktreeWithRemoteStartPoint tests creating a worktree from a remote branch.
func TestE2E_CreateWorktreeWithRemoteStartPoint(t *testing.T) {
	binPath := buildBinary(t)

	// Create a "remote" repo
	remoteRepo := testutil.NewTestRepo(t)
	remoteRepo.CreateFile("README.md", "# Remote")
	remoteRepo.Commit("remote initial commit")
	remoteRepo.CreateFile("remote-file.txt", "remote content")
	remoteRepo.Commit("remote second commit")

	// Clone the remote repo
	cloneDir := t.TempDir()
	clonePath := filepath.Join(cloneDir, "clone")
	cmd := exec.Command("git", "clone", remoteRepo.Root, clonePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone failed: %v\noutput: %s", err, out)
	}

	// In the clone, create a new commit on main
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = clonePath
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout failed: %v\noutput: %s", err, out)
	}

	// Create worktree from origin/main~1 (first commit)
	stdout, stderr, err := runGitWtStdout(t, binPath, clonePath, "feature-from-remote", "origin/main~1")
	if err != nil {
		t.Fatalf("git-wt with remote start-point failed: %v\nstderr: %s", err, stderr)
	}

	wtPath := strings.TrimSpace(stdout)

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatalf("worktree was not created at %s", wtPath)
	}

	// Verify the worktree is based on the first commit (should NOT have remote-file.txt)
	remoteFilePath := filepath.Join(wtPath, "remote-file.txt")
	if _, err := os.Stat(remoteFilePath); !os.IsNotExist(err) {
		t.Error("worktree should NOT have remote-file.txt (should be based on origin/main~1)")
	}
}

// TestE2E_TooManyArguments tests that more than 2 arguments without -d/-D flag returns an error.
func TestE2E_TooManyArguments(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Try to pass 3 arguments (should fail)
	_, err := runGitWt(t, binPath, repo.Root, "branch-a", "branch-b", "branch-c")
	if err == nil {
		t.Fatal("command should fail when passing more than 2 arguments without -d/-D flag")
	}
}

// TestE2E_StartPointIgnoredForExistingBranch tests that start-point is ignored when branch already exists.
func TestE2E_StartPointIgnoredForExistingBranch(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a branch
	repo.Git("branch", "existing-branch")

	// Create another commit on main
	repo.CreateFile("new-file.txt", "new content")
	repo.Commit("second commit")

	// Create worktree for existing branch with a start-point (should be ignored)
	stdout, stderr, err := runGitWtStdout(t, binPath, repo.Root, "existing-branch", "main")
	if err != nil {
		t.Fatalf("git-wt failed: %v\nstderr: %s", err, stderr)
	}

	wtPath := strings.TrimSpace(stdout)

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Fatalf("worktree was not created at %s", wtPath)
	}

	// Verify the worktree is based on existing-branch (should NOT have new-file.txt)
	// because start-point is ignored for existing branches
	newFilePath := filepath.Join(wtPath, "new-file.txt")
	if _, err := os.Stat(newFilePath); !os.IsNotExist(err) {
		t.Error("worktree should NOT have new-file.txt (start-point should be ignored for existing branch)")
	}
}

// TestE2E_DeleteMultipleWorktrees tests deleting multiple worktrees in a single command.
func TestE2E_DeleteMultipleWorktrees(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create multiple worktrees first (one by one)
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

	// Delete all worktrees at once
	out, err := runGitWt(t, binPath, repo.Root, "-D", "del-a", "del-b", "del-c")
	if err != nil {
		t.Fatalf("git-wt -D with multiple args failed: %v\noutput: %s", err, out)
	}

	// Verify all worktrees were deleted
	for _, path := range []string{pathA, pathB, pathC} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("worktree should have been deleted at %s", path)
		}
	}

	// Verify output contains deletion messages for all
	if !strings.Contains(out, "del-a") {
		t.Errorf("output should contain deletion message for 'del-a', got: %s", out)
	}
	if !strings.Contains(out, "del-b") {
		t.Errorf("output should contain deletion message for 'del-b', got: %s", out)
	}
	if !strings.Contains(out, "del-c") {
		t.Errorf("output should contain deletion message for 'del-c', got: %s", out)
	}
}

// TestE2E_DeleteMultipleWorktrees_StopsOnError tests that deletion stops on first error.
func TestE2E_DeleteMultipleWorktrees_StopsOnError(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create two worktrees (one by one)
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
	// Should fail on stop-b and NOT delete stop-c
	out, err := runGitWt(t, binPath, repo.Root, "-D", "stop-a", "stop-b", "stop-c")
	if err == nil {
		t.Fatal("command should fail when deleting non-existent worktree")
	}

	// Verify error message mentions stop-b
	if !strings.Contains(out, "stop-b") {
		t.Errorf("error should mention 'stop-b', got: %s", out)
	}

	// Verify stop-c was NOT deleted (execution stopped at stop-b error)
	if _, err := os.Stat(pathC); os.IsNotExist(err) {
		t.Error("stop-c should NOT have been deleted (execution should stop on error)")
	}
}

// TestE2E_DeleteBranchOnly tests deleting a branch that has no associated worktree.
func TestE2E_DeleteBranchOnly(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a branch without a worktree using git branch
	branchName := "branch-only"
	cmd := exec.Command("git", "branch", branchName)
	cmd.Dir = repo.Root
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to create branch: %v", err)
	}

	// Verify branch exists
	cmd = exec.Command("git", "branch", "--list", branchName)
	cmd.Dir = repo.Root
	out, err := cmd.Output()
	if err != nil || !strings.Contains(string(out), branchName) {
		t.Fatalf("branch should exist before deletion")
	}

	// Delete branch using git-wt -D
	wtOut, err := runGitWt(t, binPath, repo.Root, "-D", branchName)
	if err != nil {
		t.Fatalf("failed to delete branch-only: %v, output: %s", err, wtOut)
	}

	// Verify output message
	if !strings.Contains(wtOut, "Deleted branch") || !strings.Contains(wtOut, "no worktree was associated") {
		t.Errorf("output should indicate branch-only deletion, got: %s", wtOut)
	}

	// Verify branch no longer exists
	cmd = exec.Command("git", "branch", "--list", branchName)
	cmd.Dir = repo.Root
	out, err = cmd.Output()
	if err != nil {
		t.Fatalf("failed to list branches: %v", err)
	}
	if strings.Contains(string(out), branchName) {
		t.Error("branch should have been deleted")
	}
}

// TestE2E_DeleteBranchOnly_NotExists tests error when neither worktree nor branch exists.
func TestE2E_DeleteBranchOnly_NotExists(t *testing.T) {
	binPath := buildBinary(t)

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Try to delete non-existent branch
	out, err := runGitWt(t, binPath, repo.Root, "-D", "non-existent-branch")
	if err == nil {
		t.Fatal("should fail when deleting non-existent branch")
	}

	// Verify error message
	if !strings.Contains(out, "no worktree or branch found") {
		t.Errorf("error should mention 'no worktree or branch found', got: %s", out)
	}
}

// TestE2E_DeleteCurrentWorktree tests deleting the worktree you're currently in.
// This tests the fix for issue #58: safely remove current worktree and return to repository root.
func TestE2E_DeleteCurrentWorktree(t *testing.T) {
	binPath := buildBinary(t)

	t.Run("delete_current_worktree_outputs_main_repo_path_with_shell_integration", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree
		out, err := runGitWt(t, binPath, repo.Root, "current-wt")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Delete worktree from WITHIN the worktree itself with GIT_WT_SHELL_INTEGRATION=1
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

		// stdout should contain the main repo path as the last line (for shell integration to cd)
		lines := strings.Split(stdout, "\n")
		lastLine := lines[len(lines)-1]

		if lastLine != repo.Root {
			t.Errorf("last line should be main repo path %q, got %q", repo.Root, lastLine)
		}

		// Verify worktree was deleted
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}
	})

	t.Run("delete_current_worktree_no_path_without_shell_integration", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree
		out, err := runGitWt(t, binPath, repo.Root, "current-wt-no-shell")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Delete worktree WITHOUT GIT_WT_SHELL_INTEGRATION (direct call)
		stdout, stderr, err := runGitWtStdout(t, binPath, wtPath, "-D", "current-wt-no-shell")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\nstderr: %s", err, stderr)
		}

		trimmed := strings.TrimSpace(stdout)
		if trimmed == "" {
			t.Fatal("expected output but got none")
		}

		// stdout should NOT contain the main repo path (no shell integration)
		lines := strings.Split(trimmed, "\n")
		lastLine := lines[len(lines)-1]

		// Last line should be the deletion message, not a directory path
		if !strings.Contains(lastLine, "Deleted") {
			t.Errorf("last line should be a deletion message containing %q, got %q", "Deleted", lastLine)
		}

		// Verify worktree was deleted
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			t.Error("worktree should have been deleted")
		}
	})

	t.Run("delete_other_worktree_does_not_output_path", func(t *testing.T) {
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
		wtPathB := worktreePath(outB)

		// Delete wt-a from wt-b (not deleting current worktree)
		stdout, stderr, err := runGitWtStdout(t, binPath, wtPathB, "-D", "wt-a")
		if err != nil {
			t.Fatalf("git-wt -D failed: %v\nstderr: %s", err, stderr)
		}

		trimmed := strings.TrimSpace(stdout)
		if trimmed == "" {
			t.Fatal("expected output but got none")
		}

		// stdout should NOT contain a directory path (since we're not deleting current worktree)
		lines := strings.Split(trimmed, "\n")
		lastLine := lines[len(lines)-1]

		// Last line should be the deletion message, not a path
		if !strings.Contains(lastLine, "Deleted") {
			t.Errorf("last line should be a deletion message containing %q, got %q", "Deleted", lastLine)
		}
	})

	t.Run("delete_current_worktree_safe_delete_with_shell_integration", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		// Create worktree
		out, err := runGitWt(t, binPath, repo.Root, "safe-del-wt")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}
		wtPath := worktreePath(out)

		// Delete worktree using -d (safe delete) from within the worktree with GIT_WT_SHELL_INTEGRATION=1
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

		// stdout should contain the main repo path as the last line
		lines := strings.Split(stdout, "\n")
		lastLine := lines[len(lines)-1]

		if lastLine != repo.Root {
			t.Errorf("last line should be main repo path %q, got %q", repo.Root, lastLine)
		}
	})
}

// TestE2E_DeleteCurrentWorktree_ShellIntegration tests shell integration for deleting current worktree.
func TestE2E_DeleteCurrentWorktree_ShellIntegration(t *testing.T) {
	binPath := buildBinary(t)

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
			if _, err := exec.LookPath(tt.shell); err != nil {
				t.Skipf("%s not available", tt.shell)
			}

			repo := testutil.NewTestRepo(t)
			repo.CreateFile("README.md", "# Test")
			repo.Commit("initial commit")

			branchName := fmt.Sprintf("del-%s-test", tt.shell)

			// Create worktree
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

			// After deleting current worktree, pwd should be main repo root
			if pwd != repo.Root {
				t.Errorf("pwd should be main repo root %q after deleting current worktree, got: %s", repo.Root, pwd)
			}
		})
	}
}

// TestE2E_ShellIntegration_PowerShell tests the actual shell integration with PowerShell.
func TestE2E_ShellIntegration_PowerShell(t *testing.T) {
	// PowerShell init script uses git.exe which is Windows-specific
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell shell integration test is only supported on Windows")
	}

	// Try pwsh first (cross-platform), then powershell (Windows)
	var pwshPath string
	if p, err := exec.LookPath("pwsh"); err == nil {
		pwshPath = p
	} else if p, err := exec.LookPath("powershell"); err == nil {
		pwshPath = p
	} else {
		t.Skip("PowerShell not available")
	}

	binPath := buildBinary(t)
	// On Windows, binary needs .exe extension
	if runtime.GOOS == "windows" && !strings.HasSuffix(binPath, ".exe") {
		binPath += ".exe"
	}

	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	script := fmt.Sprintf(`
$ErrorActionPreference = "Stop"
Set-Location %q
$env:PATH = %q + [IO.Path]::PathSeparator + $env:PATH
Invoke-Expression (git wt --init powershell | Out-String)

# Test: git wt <branch> should cd to the worktree
git wt shell-pwsh-test
Get-Location | Select-Object -ExpandProperty Path
`, repo.Root, filepath.Dir(binPath))

	cmd := exec.Command(pwshPath, "-NoProfile", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell shell integration failed: %v\noutput: %s", err, out)
	}

	output := strings.TrimSpace(string(out))
	lines := strings.Split(output, "\n")
	pwd := strings.TrimSpace(lines[len(lines)-1])

	if !strings.Contains(pwd, "shell-pwsh-test") {
		t.Errorf("pwd should contain worktree path, got: %s", pwd)
	}
}
