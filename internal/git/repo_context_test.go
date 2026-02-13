package git

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/exec"
	"github.com/k1LoW/git-wt/testutil"
)

func TestDetectRepoContext_NormalRepo(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	rc, err := DetectRepoContext(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc.Bare {
		t.Error("Bare should be false for normal repository")
	}
	if rc.Worktree {
		t.Error("Worktree should be false for main working tree")
	}
}

func TestDetectRepoContext_NormalWorktree(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a linked worktree from the normal repo
	wtPath := filepath.Join(repo.ParentDir(), "wt-feature")
	cmd := exec.Command("git", "-C", repo.Root, "worktree", "add", "-b", "feature", wtPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wtPath) })

	// Change to the worktree directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(wtPath); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore cwd: %v", err)
		}
	}()

	rc, err := DetectRepoContext(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rc.Bare {
		t.Error("Bare should be false for normal repository's worktree")
	}
	if !rc.Worktree {
		t.Error("Worktree should be true for linked worktree")
	}
}

func TestDetectRepoContext_BareRepo(t *testing.T) {
	bareRepo := testutil.NewBareTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(bareRepo.Root); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore cwd: %v", err)
		}
	}()

	rc, err := DetectRepoContext(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rc.Bare {
		t.Error("Bare should be true for bare repository")
	}
	if rc.Worktree {
		t.Error("Worktree should be false at bare repository root")
	}
}

func TestDetectRepoContext_WorktreeFromBare(t *testing.T) {
	bareRepo := testutil.NewBareTestRepo(t)

	// Create a worktree from the bare repo
	wtPath := filepath.Join(bareRepo.ParentDir(), "wt-test")
	cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wtPath) })

	// Change to the worktree directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(wtPath); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore cwd: %v", err)
		}
	}()

	rc, err := DetectRepoContext(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rc.Bare {
		t.Error("Bare should be true for worktree from bare repository")
	}
	if !rc.Worktree {
		t.Error("Worktree should be true inside a linked worktree from bare")
	}
}

func TestIsBareRepository_NormalRepo(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	isBare, err := IsBareRepository(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isBare {
		t.Error("normal repository should not be detected as bare")
	}
}

func TestIsBareRepository_BareRepo(t *testing.T) {
	bareRepo := testutil.NewBareTestRepo(t)

	// Change to the bare repo directory to run git commands there
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(bareRepo.Root); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore cwd: %v", err)
		}
	}()

	isBare, err := IsBareRepository(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isBare {
		t.Error("bare repository should be detected as bare")
	}
}

func TestIsBareRepository_WorktreeFromBare(t *testing.T) {
	bareRepo := testutil.NewBareTestRepo(t)

	// Create a worktree from the bare repo
	wtPath := filepath.Join(bareRepo.ParentDir(), "wt-test")
	cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wtPath) })

	// Change to the worktree directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(wtPath); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore cwd: %v", err)
		}
	}()

	// Even from a worktree, IsBareRepository should detect the parent bare repo
	isBare, err := IsBareRepository(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isBare {
		t.Error("worktree from bare repository should be detected as bare")
	}
}

func TestAssertNotBareRepository_NormalRepo(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	err := AssertNotBareRepository(t.Context())
	if err != nil {
		t.Errorf("expected nil from AssertNotBareRepository for normal repo, got: %v", err)
	}
}

func TestAssertNotBareRepository_BareRepo(t *testing.T) {
	bareRepo := testutil.NewBareTestRepo(t)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}
	if err := os.Chdir(bareRepo.Root); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore cwd: %v", err)
		}
	}()

	err = AssertNotBareRepository(t.Context())
	if err == nil {
		t.Fatal("AssertNotBareRepository should return error for bare repo")
	}
	if !errors.Is(err, ErrBareRepository) {
		t.Errorf("expected ErrBareRepository, got: %v", err)
	}
}
