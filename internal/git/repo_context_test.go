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
	if rc.bare {
		t.Error("Bare should be false for normal repository")
	}
	if rc.worktree {
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
	if rc.bare {
		t.Error("Bare should be false for normal repository's worktree")
	}
	if !rc.worktree {
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
	if !rc.bare {
		t.Error("Bare should be true for bare repository")
	}
	if rc.worktree {
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
	if !rc.bare {
		t.Error("Bare should be true for worktree from bare repository")
	}
	if !rc.worktree {
		t.Error("Worktree should be true inside a linked worktree from bare")
	}
}

// TestDetectRepoContext_DotGitBareRepo tests a bare repository where
// git-common-dir ends with ".git" (created via core.bare = true).
// This is the case that filepath.Base heuristic alone cannot detect.
func TestDetectRepoContext_DotGitBareRepo(t *testing.T) {
	bareRepo := testutil.NewDotGitBareTestRepo(t)

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
	if !rc.bare {
		t.Error("bare should be true for dotgit bare repository")
	}
	if rc.worktree {
		t.Error("worktree should be false at dotgit bare repository root")
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

	rc := RepoContext{bare: false, worktree: true, dir: cwd}
	ctx := WithRepoContext(t.Context(), rc)

	got := RepoContextFrom(ctx)
	if got == nil {
		t.Fatal("RepoContextFrom returned nil, expected cached RepoContext")
	}
	if got.bare != rc.bare {
		t.Errorf("bare: got %v, want %v", got.bare, rc.bare)
	}
	if got.worktree != rc.worktree {
		t.Errorf("worktree: got %v, want %v", got.worktree, rc.worktree)
	}
	if got.dir != rc.dir {
		t.Errorf("dir: got %q, want %q", got.dir, rc.dir)
	}
}

func TestRepoContextFrom_CwdMismatch(t *testing.T) {
	rc := RepoContext{bare: false, worktree: false, dir: "/nonexistent/path"}
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

func TestShowPrefix(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile("some/path/file.txt", "content")
	repo.Commit("initial commit")

	t.Run("at_repo_root", func(t *testing.T) {
		restore := repo.Chdir()
		defer restore()

		prefix, err := ShowPrefix(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prefix != "" {
			t.Errorf("ShowPrefix() at root = %q, want empty string", prefix) //nostyle:errorstrings
		}
	})

	t.Run("in_subdirectory", func(t *testing.T) {
		restore := repo.Chdir()
		defer restore()

		subdir := filepath.Join(repo.Root, "some", "path")
		if err := os.Chdir(subdir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		prefix, err := ShowPrefix(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if prefix != "some/path" {
			t.Errorf("ShowPrefix() in subdir = %q, want %q", prefix, "some/path") //nostyle:errorstrings
		}
	})
}

func TestMainRepoRoot(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a subdirectory
	repo.CreateFile("subdir/file.txt", "content")
	repo.Commit("add subdir")

	restore := repo.Chdir()
	defer restore()

	root, err := MainRepoRoot(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root != repo.Root {
		t.Errorf("MainRepoRoot() = %q, want %q", root, repo.Root) //nostyle:errorstrings
	}

	// Test from subdirectory
	subdir := filepath.Join(repo.Root, "subdir")
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("failed to chdir to subdir: %v", err)
	}

	root, err = MainRepoRoot(t.Context())
	if err != nil {
		t.Fatalf("unexpected error from subdir: %v", err)
	}

	if root != repo.Root {
		t.Errorf("MainRepoRoot() from subdir = %q, want %q", root, repo.Root) //nostyle:errorstrings
	}
}

func TestRepoName(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	name, err := RepoName(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The repo is created in a temp directory named "repo"
	if name != "repo" {
		t.Errorf("RepoName() = %q, want %q", name, "repo") //nostyle:errorstrings
	}
}

func TestGitDirs_NormalRepo(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	gitDir, gitCommonDir, err := gitDirs(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Normal repo: both should point to .git, and they should be equal
	if filepath.Base(gitDir) != ".git" {
		t.Errorf("gitDir base should be .git, got %q", filepath.Base(gitDir))
	}
	if filepath.Base(gitCommonDir) != ".git" {
		t.Errorf("gitCommonDir base should be .git, got %q", filepath.Base(gitCommonDir))
	}
	if gitDir != gitCommonDir {
		t.Errorf("gitDir and gitCommonDir should be equal in normal repo\ngitDir:      %s\ngitCommonDir: %s", gitDir, gitCommonDir)
	}
}

func TestGitDirs_NormalWorktree(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	wtPath := filepath.Join(repo.ParentDir(), "wt-feature")
	cmd := exec.Command("git", "-C", repo.Root, "worktree", "add", "-b", "feature", wtPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wtPath) })

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

	gitDir, gitCommonDir, err := gitDirs(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Normal worktree: gitCommonDir base is .git, but gitDir differs
	if filepath.Base(gitCommonDir) != ".git" {
		t.Errorf("gitCommonDir base should be .git, got %q", filepath.Base(gitCommonDir))
	}
	if gitDir == gitCommonDir {
		t.Error("gitDir and gitCommonDir should differ in a linked worktree")
	}
}

func TestGitDirs_BareRepo(t *testing.T) {
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

	gitDir, gitCommonDir, err := gitDirs(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bare repo: base is NOT .git, and gitDir == gitCommonDir
	if filepath.Base(gitCommonDir) == ".git" {
		t.Errorf("gitCommonDir base should NOT be .git for bare repo, got %q", filepath.Base(gitCommonDir))
	}
	if gitDir != gitCommonDir {
		t.Errorf("gitDir and gitCommonDir should be equal in bare repo root\ngitDir:      %s\ngitCommonDir: %s", gitDir, gitCommonDir)
	}
}

func TestGitDirs_WorktreeFromBare(t *testing.T) {
	bareRepo := testutil.NewBareTestRepo(t)

	wtPath := filepath.Join(bareRepo.ParentDir(), "wt-test")
	cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wtPath) })

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

	gitDir, gitCommonDir, err := gitDirs(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Worktree from bare: base is NOT .git, and gitDir differs from gitCommonDir
	if filepath.Base(gitCommonDir) == ".git" {
		t.Errorf("gitCommonDir base should NOT be .git for bare-derived worktree, got %q", filepath.Base(gitCommonDir))
	}
	if gitDir == gitCommonDir {
		t.Error("gitDir and gitCommonDir should differ in a linked worktree from bare")
	}
}

func TestMainRepoRoot_BareRepo(t *testing.T) {
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

	root, err := MainRepoRoot(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bare repo: MainRepoRoot should return the bare repo directory itself
	if root != bareRepo.Root {
		t.Errorf("MainRepoRoot() = %q, want %q", root, bareRepo.Root) //nostyle:errorstrings
	}
}

func TestMainRepoRoot_WorktreeFromBare(t *testing.T) {
	bareRepo := testutil.NewBareTestRepo(t)

	wtPath := filepath.Join(bareRepo.ParentDir(), "wt-test")
	cmd := exec.Command("git", "-C", bareRepo.Root, "worktree", "add", wtPath, "main")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git worktree add failed: %v\noutput: %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wtPath) })

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

	root, err := MainRepoRoot(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// From a bare-derived worktree, MainRepoRoot should return the bare repo directory
	if root != bareRepo.Root {
		t.Errorf("MainRepoRoot() = %q, want %q", root, bareRepo.Root) //nostyle:errorstrings
	}
}

func TestRepoName_BareRepo(t *testing.T) {
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

	name, err := RepoName(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// BareTestRepo creates "repo.git", so RepoName should return "repo.git"
	if name != "repo.git" {
		t.Errorf("RepoName() = %q, want %q", name, "repo.git") //nostyle:errorstrings
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
	cached := RepoContext{bare: true, worktree: true, dir: cwd}
	ctx := WithRepoContext(t.Context(), cached)

	// DetectRepoContext should return the cached value without running git
	rc, err := DetectRepoContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The cached value has Bare=true, but the actual repo is not bare.
	// If caching works, we get the cached value back.
	if !rc.bare {
		t.Error("expected cached Bare=true, got false (cache was not used)")
	}
	if !rc.worktree {
		t.Error("expected cached Worktree=true, got false (cache was not used)")
	}
}
