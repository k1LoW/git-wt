package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
`, repoRoot, pathDir) //nostyle:funcfmt
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
`, repoRoot, pathDir) //nostyle:funcfmt
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
`, repoRoot, pathDir) //nostyle:funcfmt
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
		"git wt [branch|worktree]",
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
`, repo.Root, filepath.Dir(binPath)) //nostyle:funcfmt

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
`, repo.Root, filepath.Dir(binPath)) //nostyle:funcfmt

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
`, repo.Root, filepath.Dir(binPath)) //nostyle:funcfmt

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
`, repo.Root, filepath.Dir(binPath)) //nostyle:funcfmt

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
