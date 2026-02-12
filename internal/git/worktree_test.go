package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/git-wt/testutil"
)

func TestListWorktrees(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	// Initially, only the main worktree should exist
	worktrees, err := ListWorktrees(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(worktrees) != 1 {
		t.Errorf("expected 1 worktree, got %d", len(worktrees))
	}

	if worktrees[0].Branch != "main" {
		t.Errorf("expected branch 'main', got %q", worktrees[0].Branch)
	}

	if worktrees[0].Path != repo.Root {
		t.Errorf("expected path %q, got %q", repo.Root, worktrees[0].Path)
	}
}

func TestListWorktrees_BareRepo(t *testing.T) {
	repo := testutil.NewBareTestRepo(t)

	t.Cleanup(repo.Chdir())

	worktrees, err := ListWorktrees(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(worktrees))
	}

	wt := worktrees[0]

	if !wt.Bare {
		t.Error("expected Bare = true")
	}
	if wt.Path != repo.Root {
		t.Errorf("expected path %q, got %q", repo.Root, wt.Path)
	}
	if wt.Branch == "" {
		t.Error("bare worktree Branch should not be empty")
	}
	if wt.Head == "" {
		t.Error("bare worktree Head should not be empty")
	}
}

func TestListWorktrees_Multiple(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a worktree
	wtPath := filepath.Join(repo.ParentDir(), "worktree-feature")
	repo.Git("worktree", "add", "-b", "feature", wtPath)

	restore := repo.Chdir()
	defer restore()

	worktrees, err := ListWorktrees(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(worktrees) != 2 {
		t.Errorf("expected 2 worktrees, got %d", len(worktrees))
	}

	// Check that both worktrees are present
	branches := make(map[string]string)
	for _, wt := range worktrees {
		branches[wt.Branch] = wt.Path
	}

	if _, ok := branches["main"]; !ok {
		t.Error("main worktree not found")
	}
	if _, ok := branches["feature"]; !ok {
		t.Error("feature worktree not found")
	}
}

func TestCurrentWorktree(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	path, err := CurrentWorktree(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != repo.Root {
		t.Errorf("CurrentWorktree() = %q, want %q", path, repo.Root) //nostyle:errorstrings
	}
}

func TestCurrentWorktree_BareRepo(t *testing.T) {
	repo := testutil.NewBareTestRepo(t)

	t.Cleanup(repo.Chdir())

	path, err := CurrentWorktree(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != repo.Root {
		t.Errorf("CurrentWorktree() = %q, want %q", path, repo.Root) //nostyle:errorstrings
	}
}

func TestFindWorktreeByBranchOrDir_BareRepo(t *testing.T) {
	repo := testutil.NewBareTestRepo(t)

	t.Cleanup(repo.Chdir())

	// The bare repo has HEAD pointing to "main", but FindWorktreeByBranchOrDir
	// should NOT match the bare entry — bare entries are not switchable worktrees.
	wt, err := FindWorktreeByBranchOrDir(t.Context(), "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wt != nil {
		t.Errorf("expected nil for bare-only repo, got path=%q", wt.Path)
	}
}

func TestFindWorktreeByBranch_BareRepo(t *testing.T) {
	repo := testutil.NewBareTestRepo(t)

	t.Cleanup(repo.Chdir())

	// The bare repo has HEAD pointing to "main", but FindWorktreeByBranch
	// should NOT match the bare entry — bare entries are not switchable worktrees.
	wt, err := FindWorktreeByBranch(t.Context(), "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wt != nil {
		t.Errorf("expected nil for bare-only repo, got path=%q", wt.Path)
	}
}

func TestFindWorktreeByBranch(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a worktree
	wtPath := filepath.Join(repo.ParentDir(), "worktree-feature")
	repo.Git("worktree", "add", "-b", "feature", wtPath)

	restore := repo.Chdir()
	defer restore()

	tests := []struct {
		name     string
		branch   string
		wantNil  bool
		wantPath string
	}{
		{"existing main branch", "main", false, repo.Root},
		{"existing feature branch", "feature", false, wtPath},
		{"non-existing branch", "no-such-branch", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wt, err := FindWorktreeByBranch(t.Context(), tt.branch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil {
				if wt != nil {
					t.Errorf("expected nil, got worktree with path %q", wt.Path)
				}
				return
			}

			if wt == nil {
				t.Fatal("expected worktree, got nil")
			}

			if wt.Path != tt.wantPath {
				t.Errorf("worktree path = %q, want %q", wt.Path, tt.wantPath)
			}
		})
	}
}

func TestAddWorktree(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")
	repo.Git("branch", "existing-branch")

	restore := repo.Chdir()
	defer restore()

	wtPath := filepath.Join(repo.ParentDir(), "worktree-existing")
	err := AddWorktree(t.Context(), wtPath, "existing-branch", CopyOptions{})
	if err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// Verify basedir files were created
	baseDir := filepath.Dir(wtPath)
	if _, err := os.Stat(filepath.Join(baseDir, ".gitignore")); os.IsNotExist(err) {
		t.Error(".gitignore was not created in basedir")
	}
	if _, err := os.Stat(filepath.Join(baseDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md was not created in basedir")
	}

	// Verify it appears in worktree list
	wt, err := FindWorktreeByBranch(t.Context(), "existing-branch")
	if err != nil {
		t.Fatalf("FindWorktreeByBranch failed: %v", err)
	}
	if wt == nil {
		t.Error("worktree not found after creation")
	}
}

func TestAddWorktreeWithNewBranch(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	wtPath := filepath.Join(repo.ParentDir(), "worktree-new")
	err := AddWorktreeWithNewBranch(t.Context(), wtPath, "new-branch", "", CopyOptions{})
	if err != nil {
		t.Fatalf("AddWorktreeWithNewBranch failed: %v", err)
	}

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// Verify basedir files were created
	baseDir := filepath.Dir(wtPath)
	if _, err := os.Stat(filepath.Join(baseDir, ".gitignore")); os.IsNotExist(err) {
		t.Error(".gitignore was not created in basedir")
	}
	if _, err := os.Stat(filepath.Join(baseDir, "README.md")); os.IsNotExist(err) {
		t.Error("README.md was not created in basedir")
	}

	// Verify branch was created
	exists, err := LocalBranchExists(t.Context(), "new-branch")
	if err != nil {
		t.Fatalf("LocalBranchExists failed: %v", err)
	}
	if !exists {
		t.Error("branch was not created")
	}

	// Verify it appears in worktree list
	wt, err := FindWorktreeByBranch(t.Context(), "new-branch")
	if err != nil {
		t.Fatalf("FindWorktreeByBranch failed: %v", err)
	}
	if wt == nil {
		t.Error("worktree not found after creation")
	}
}

func TestAddWorktree_BareRepo(t *testing.T) {
	repo := testutil.NewBareTestRepo(t)

	t.Cleanup(repo.Chdir())

	// Create a branch in the bare repo
	repo.Git("branch", "existing-branch")

	wtPath := filepath.Join(repo.ParentDir(), "worktree-existing")
	err := AddWorktree(t.Context(), wtPath, "existing-branch", CopyOptions{})
	if err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// Verify it appears in worktree list
	wt, err := FindWorktreeByBranch(t.Context(), "existing-branch")
	if err != nil {
		t.Fatalf("FindWorktreeByBranch failed: %v", err)
	}
	if wt == nil {
		t.Error("worktree not found after creation")
	}
}

func TestAddWorktreeWithNewBranch_BareRepo(t *testing.T) {
	repo := testutil.NewBareTestRepo(t)

	t.Cleanup(repo.Chdir())

	wtPath := filepath.Join(repo.ParentDir(), "worktree-new")
	err := AddWorktreeWithNewBranch(t.Context(), wtPath, "new-branch", "", CopyOptions{})
	if err != nil {
		t.Fatalf("AddWorktreeWithNewBranch failed: %v", err)
	}

	// Verify worktree was created
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// Verify branch was created
	exists, err := LocalBranchExists(t.Context(), "new-branch")
	if err != nil {
		t.Fatalf("LocalBranchExists failed: %v", err)
	}
	if !exists {
		t.Error("branch was not created")
	}

	// Verify it appears in worktree list
	wt, err := FindWorktreeByBranch(t.Context(), "new-branch")
	if err != nil {
		t.Fatalf("FindWorktreeByBranch failed: %v", err)
	}
	if wt == nil {
		t.Error("worktree not found after creation")
	}
}

func TestFindWorktreeByBranchOrDir_BareRepo_DotPath(t *testing.T) {
	repo := testutil.NewBareTestRepo(t)

	t.Cleanup(repo.Chdir())

	// Create a linked worktree from the bare repo
	wtPath := filepath.Join(repo.ParentDir(), "wt-feature")
	err := AddWorktreeWithNewBranch(t.Context(), wtPath, "feature", "", CopyOptions{})
	if err != nil {
		t.Fatalf("AddWorktreeWithNewBranch failed: %v", err)
	}

	// Change into the linked worktree directory
	if err := os.Chdir(wtPath); err != nil {
		t.Fatalf("failed to chdir to worktree: %v", err)
	}

	// FindWorktreeByBranchOrDir with "." should find the current worktree
	wt, err := FindWorktreeByBranchOrDir(t.Context(), ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wt == nil {
		t.Fatal("expected to find worktree for \".\", got nil")
	}
	if wt.Branch != "feature" {
		t.Errorf("expected branch 'feature', got %q", wt.Branch)
	}
}

func TestRemoveWorktree(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a worktree
	wtPath := filepath.Join(repo.ParentDir(), "worktree-to-remove")
	repo.Git("worktree", "add", "-b", "to-remove", wtPath)

	restore := repo.Chdir()
	defer restore()

	// Verify worktree exists
	wt, err := FindWorktreeByBranch(t.Context(), "to-remove")
	if err != nil {
		t.Fatalf("FindWorktreeByBranch failed: %v", err)
	}
	if wt == nil {
		t.Fatal("worktree should exist before removal")
	}

	// Remove worktree
	err = RemoveWorktree(t.Context(), wtPath, false)
	if err != nil {
		t.Fatalf("RemoveWorktree failed: %v", err)
	}

	// Verify worktree no longer exists
	wt, err = FindWorktreeByBranch(t.Context(), "to-remove")
	if err != nil {
		t.Fatalf("FindWorktreeByBranch failed: %v", err)
	}
	if wt != nil {
		t.Error("worktree should not exist after removal")
	}
}

func TestRemoveWorktree_Force(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a worktree
	wtPath := filepath.Join(repo.ParentDir(), "worktree-dirty")
	repo.Git("worktree", "add", "-b", "dirty", wtPath)

	// Make the worktree dirty (untracked file)
	if err := os.WriteFile(filepath.Join(wtPath, "dirty.txt"), []byte("dirty"), 0600); err != nil {
		t.Fatalf("failed to create dirty file: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Force remove worktree
	err := RemoveWorktree(t.Context(), wtPath, true)
	if err != nil {
		t.Fatalf("RemoveWorktree(force) failed: %v", err)
	}

	// Verify worktree no longer exists
	wt, err := FindWorktreeByBranch(t.Context(), "dirty")
	if err != nil {
		t.Fatalf("FindWorktreeByBranch failed: %v", err)
	}
	if wt != nil {
		t.Error("worktree should not exist after force removal")
	}
}
