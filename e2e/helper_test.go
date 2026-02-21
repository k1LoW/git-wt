// Package e2e contains end-to-end tests for git-wt.
//
// helper_test.go provides shared test utilities:
//   - buildBinary: builds the git-wt binary for testing
//   - runGitWt: executes git-wt and returns combined output
//   - runGitWtStdout: executes git-wt and returns stdout/stderr separately
//   - runGitWtWithStderr: executes git-wt with isolated HOME and returns stdout/stderr separately
//   - runGitWtWithShellIntegration: executes git-wt with GIT_WT_SHELL_INTEGRATION=1
//   - worktreePath: extracts worktree path from command output
//   - addRawWorktreeFromBare: creates a worktree via raw git command
//   - assertWorktreeExists: asserts that a worktree directory exists
//   - assertWorktreeDeleted: asserts that a worktree directory has been removed
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

// runGitWtWithStderr runs git-wt with an isolated HOME directory and returns stdout, stderr, and error separately.
// This is useful for tests that need to avoid interference from user-level git config.
func runGitWtWithStderr(t *testing.T, binPath, dir string, args ...string) (string, string, error) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
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

// addRawWorktreeFromBare creates a worktree from a bare repository using a raw
// git worktree add command and registers cleanup to remove it.
func addRawWorktreeFromBare(t *testing.T, root, wtPath, branch string) {
	t.Helper()
	cmd := exec.Command("git", "-C", root, "worktree", "add", wtPath, branch)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wtPath) })
}

// assertWorktreeExists asserts that path is non-empty and the directory exists.
func assertWorktreeExists(t *testing.T, path string) {
	t.Helper()
	if path == "" {
		t.Fatal("expected worktree path in stdout, got empty")
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("worktree directory should exist at %s", path)
	}
}

// assertWorktreeDeleted asserts that the worktree directory no longer exists.
func assertWorktreeDeleted(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("worktree directory should have been deleted: %s", path)
	}
}

// runGitWtWithShellIntegration runs git-wt with GIT_WT_SHELL_INTEGRATION=1
// and returns stdout, stderr, and the error.
func runGitWtWithShellIntegration(t *testing.T, binPath, dir string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binPath, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_WT_SHELL_INTEGRATION=1")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err := cmd.Run()
	return stdoutBuf.String(), stderrBuf.String(), err
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
