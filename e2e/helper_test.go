// Package e2e contains end-to-end tests for git-wt.
//
// helper_test.go provides shared test utilities:
//   - buildBinary: builds the git-wt binary for testing
//   - runGitWt: executes git-wt and returns combined output
//   - runGitWtStdout: executes git-wt and returns stdout/stderr separately
//   - worktreePath: extracts worktree path from command output
package e2e

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/exec"
)

func TestMain(m *testing.M) {
	// Prevent the user's global/system git config from leaking into tests.
	// See: https://git-scm.com/docs/git-config#ENVIRONMENT (Git 2.32+)
	os.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	os.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	os.Exit(m.Run())
}

// buildBinary builds git-wt binary for testing and returns the path.
func buildBinary(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "git-wt")

	cmd := exec.Command("go", "build", "-o", binPath, "..")
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

// createBareWorktree creates a worktree from the given directory using git-wt
// and returns the worktree path.
func createBareWorktree(t *testing.T, binPath, dir, branch string) string {
	t.Helper()

	stdout, _, err := runGitWtStdout(t, binPath, dir, branch)
	if err != nil {
		t.Fatalf("failed to create worktree %q: %v\nstdout: %s", branch, err, stdout)
	}
	return worktreePath(stdout)
}

// commitUnmergedChange creates a file and commits it in the given directory,
// producing an unmerged commit relative to the parent branch.
func commitUnmergedChange(t *testing.T, dir string) {
	t.Helper()

	// Ensure git user config is available (bare repo worktrees may not
	// inherit user config when GIT_CONFIG_GLOBAL is /dev/null).
	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git %s failed: %v", strings.Join(args, " "), err)
		}
	}

	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("content"), 0600); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "unmerged commit")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}
}

// assertLastLine checks that the last line of output matches the expected string.
func assertLastLine(t *testing.T, output, expected string) {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	lastLine := lines[len(lines)-1]
	if lastLine != expected {
		t.Errorf("last line should be %q, got %q", expected, lastLine)
	}
}
