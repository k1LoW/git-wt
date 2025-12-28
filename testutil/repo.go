package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/k1LoW/exec"
)

// TestRepo represents a temporary git repository for testing.
type TestRepo struct {
	t    testing.TB
	Root string
}

// NewTestRepo creates a new temporary git repository.
// Cleanup is automatically registered via t.Cleanup().
func NewTestRepo(t testing.TB) *TestRepo { //nostyle:repetition
	t.Helper()

	// Create a parent temp directory to contain both the repo and potential worktrees
	parentDir, err := os.MkdirTemp("", "git-wt-test-*")
	if err != nil {
		t.Fatalf("failed to create temp parent dir: %v", err)
	}

	// Resolve symlinks (macOS /var -> /private/var issue)
	parentDir, err = filepath.EvalSymlinks(parentDir)
	if err != nil {
		os.RemoveAll(parentDir)
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	repoDir := filepath.Join(parentDir, "repo")
	if err := os.Mkdir(repoDir, 0755); err != nil {
		os.RemoveAll(parentDir)
		t.Fatalf("failed to create repo dir: %v", err)
	}

	r := &TestRepo{
		t:    t,
		Root: repoDir,
	}

	t.Cleanup(func() {
		os.RemoveAll(parentDir)
	})

	// Initialize git repository
	r.Git("init")
	r.Git("config", "user.email", "test@example.com")
	r.Git("config", "user.name", "Test User")
	// Set default branch name to main
	r.Git("checkout", "-b", "main")

	return r
}

// Git executes a git command in the repository and returns stdout.
// It calls t.Fatal on error.
func (r *TestRepo) Git(args ...string) string {
	r.t.Helper()
	out, err := r.GitE(args...)
	if err != nil {
		r.t.Fatalf("git %s failed: %v\noutput: %s", strings.Join(args, " "), err, out)
	}
	return out
}

// GitE executes a git command and returns stdout and error.
// Use this for testing error cases.
func (r *TestRepo) GitE(args ...string) (string, error) {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Root
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// CreateFile creates a file with the given content.
// It calls t.Fatal on error.
func (r *TestRepo) CreateFile(path, content string) {
	r.t.Helper()
	fullPath := filepath.Join(r.Root, path)

	// Create parent directories if needed
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		r.t.Fatalf("failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
		r.t.Fatalf("failed to create file %s: %v", path, err)
	}
}

// Commit stages all changes and commits with the given message.
// It calls t.Fatal on error.
func (r *TestRepo) Commit(message string) {
	r.t.Helper()
	r.Git("add", "-A")
	r.Git("commit", "-m", message)
}

// Chdir changes the current working directory to the repository root.
// Returns a function to restore the original directory.
// It calls t.Fatal on error.
func (r *TestRepo) Chdir() func() {
	r.t.Helper()
	origDir, err := os.Getwd()
	if err != nil {
		r.t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(r.Root); err != nil {
		r.t.Fatalf("failed to chdir to %s: %v", r.Root, err)
	}

	return func() {
		if err := os.Chdir(origDir); err != nil {
			r.t.Fatalf("failed to restore directory to %s: %v", origDir, err)
		}
	}
}

// Path returns the absolute path to a file in the repository.
func (r *TestRepo) Path(relPath string) string {
	return filepath.Join(r.Root, relPath)
}

// ParentDir returns the parent directory of the repository.
// This is useful for worktree operations that create directories outside the repo.
func (r *TestRepo) ParentDir() string {
	return filepath.Dir(r.Root)
}
