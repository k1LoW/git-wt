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
	// Disable GPG signing for tests
	r.Git("config", "commit.gpgsign", "false")
	// Set default branch name to main (for DefaultBranch detection in local repos)
	r.Git("config", "init.defaultBranch", "main")
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

// BareTestRepo represents a temporary bare git repository for testing.
// A bare repository has no working tree — it contains only the .git internals
// (objects, refs, HEAD, etc.) directly in its root directory.
// `git worktree list --porcelain` outputs a "bare" line for the first entry
// of a bare repository, which is how git-wt detects bare repos.
type BareTestRepo struct {
	t    testing.TB
	Root string // Path to the bare repository (e.g., /tmp/xxx/repo.git)
}

// NewBareTestRepo creates a new temporary bare git repository with an initial commit.
//
// Because bare repositories have no working tree, we cannot commit directly.
// Instead, this helper:
//   1. Creates a bare repo with `git init --bare`
//   2. Creates a temporary normal repo, makes an initial commit
//   3. Pushes the commit to the bare repo so it has at least one ref
//      (required for `git worktree add` to work)
//   4. Cleans up the temporary normal repo (the bare repo persists for the test)
//
// Cleanup of the bare repo is automatically registered via t.Cleanup().
func NewBareTestRepo(t testing.TB) *BareTestRepo { //nostyle:repetition
	t.Helper()

	// Create a parent temp directory to contain both the bare repo and potential worktrees
	parentDir, err := os.MkdirTemp("", "git-wt-bare-test-*")
	if err != nil {
		t.Fatalf("failed to create temp parent dir: %v", err)
	}

	// Resolve symlinks (macOS /var -> /private/var issue)
	parentDir, err = filepath.EvalSymlinks(parentDir)
	if err != nil {
		os.RemoveAll(parentDir)
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	bareDir := filepath.Join(parentDir, "repo.git")

	t.Cleanup(func() {
		os.RemoveAll(parentDir)
	})

	// 1. Initialize bare repository
	cmd := newGitCmd(t, "", "init", "--bare", bareDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init --bare failed: %v\noutput: %s", err, out)
	}

	// Configure bare repo to accept pushes
	cmd = newGitCmd(t, bareDir, "config", "receive.denyCurrentBranch", "ignore")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config failed: %v\noutput: %s", err, out)
	}

	// 2. Create a temporary normal repo to make the initial commit
	tmpCloneDir := filepath.Join(parentDir, "tmp-clone")
	cmd = newGitCmd(t, "", "clone", bareDir, tmpCloneDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone failed: %v\noutput: %s", err, out)
	}

	// Configure the clone
	for _, args := range [][]string{
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test User"},
		{"config", "commit.gpgsign", "false"},
	} {
		cmd = newGitCmd(t, tmpCloneDir, args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s failed: %v\noutput: %s", strings.Join(args, " "), err, out)
		}
	}

	// 3. Create initial commit and push to bare repo
	readmePath := filepath.Join(tmpCloneDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0600); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}

	for _, args := range [][]string{
		{"checkout", "-b", "main"},
		{"add", "-A"},
		{"commit", "-m", "initial commit"},
		{"push", "origin", "main"},
	} {
		cmd = newGitCmd(t, tmpCloneDir, args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s failed: %v\noutput: %s", strings.Join(args, " "), err, out)
		}
	}

	// Set HEAD to point to main branch in the bare repo
	cmd = newGitCmd(t, bareDir, "symbolic-ref", "HEAD", "refs/heads/main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git symbolic-ref failed: %v\noutput: %s", err, out)
	}

	// 4. Remove the temporary clone (no longer needed)
	os.RemoveAll(tmpCloneDir)

	return &BareTestRepo{
		t:    t,
		Root: bareDir,
	}
}

// Git executes a git command in the bare repository and returns stdout.
// It calls t.Fatal on error.
func (r *BareTestRepo) Git(args ...string) string {
	r.t.Helper()
	cmd := newGitCmd(r.t, r.Root, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %s failed: %v\noutput: %s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// ParentDir returns the parent directory of the bare repository.
// This is useful for worktree operations that create directories outside the repo.
func (r *BareTestRepo) ParentDir() string {
	return filepath.Dir(r.Root)
}

// newGitCmd creates a git command with the given arguments.
// If dir is non-empty, the command runs in that directory.
func newGitCmd(t testing.TB, dir string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd
}
