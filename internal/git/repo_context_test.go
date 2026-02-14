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

func TestWithRepoContext_CacheHit(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	rc := RepoContext{Bare: false, Worktree: true, Dir: cwd}
	ctx := WithRepoContext(t.Context(), rc)

	got := RepoContextFrom(ctx)
	if got == nil {
		t.Fatal("RepoContextFrom returned nil, expected cached RepoContext")
	}
	if got.Bare != rc.Bare {
		t.Errorf("bare: got %v, want %v", got.Bare, rc.Bare)
	}
	if got.Worktree != rc.Worktree {
		t.Errorf("worktree: got %v, want %v", got.Worktree, rc.Worktree)
	}
	if got.Dir != rc.Dir {
		t.Errorf("dir: got %q, want %q", got.Dir, rc.Dir)
	}
}

func TestRepoContextFrom_CwdMismatch(t *testing.T) {
	rc := RepoContext{Bare: false, Worktree: false, Dir: "/nonexistent/path"}
	ctx := WithRepoContext(t.Context(), rc)

	got := RepoContextFrom(ctx)
	if got != nil {
		t.Errorf("repoContextFrom should return nil when cwd differs from Dir, got %+v", got)
	}
}

func TestRepoContextFrom_NoContext(t *testing.T) {
	got := RepoContextFrom(t.Context())
	if got != nil {
		t.Errorf("repoContextFrom should return nil for plain context, got %+v", got)
	}
}

func TestDetectRepoContext_UsesCache(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Pre-populate context with a known RepoContext
	cached := RepoContext{Bare: true, Worktree: true, Dir: cwd}
	ctx := WithRepoContext(t.Context(), cached)

	// DetectRepoContext should return the cached value without running git
	rc, err := DetectRepoContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The cached value has Bare=true, but the actual repo is not bare.
	// If caching works, we get the cached value back.
	if !rc.Bare {
		t.Error("expected cached Bare=true, got false (cache was not used)")
	}
	if !rc.Worktree {
		t.Error("expected cached Worktree=true, got false (cache was not used)")
	}
}
