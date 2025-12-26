package git

import (
	"testing"

	"github.com/k1LoW/git-wt/testutil"
)

func TestBranchExists(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")
	repo.Git("branch", "feature")

	restore := repo.Chdir()
	defer restore()

	tests := []struct {
		name   string
		branch string
		want   bool
	}{
		{"existing local branch", "feature", true},
		{"main branch", "main", true},
		{"non-existing branch", "no-such-branch", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BranchExists(tt.branch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("BranchExists(%q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}

func TestLocalBranchExists(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")
	repo.Git("branch", "local-feature")

	restore := repo.Chdir()
	defer restore()

	tests := []struct {
		name   string
		branch string
		want   bool
	}{
		{"existing local branch", "local-feature", true},
		{"main branch", "main", true},
		{"non-existing branch", "no-such-branch", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LocalBranchExists(tt.branch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("LocalBranchExists(%q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}

func TestListBranches(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")
	repo.Git("branch", "feature-a")
	repo.Git("branch", "feature-b")

	restore := repo.Chdir()
	defer restore()

	branches, err := ListBranches()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]bool{
		"main":      false,
		"feature-a": false,
		"feature-b": false,
	}

	for _, b := range branches {
		if _, ok := expected[b]; ok {
			expected[b] = true
		}
	}

	for name, found := range expected {
		if !found {
			t.Errorf("expected branch %q not found in list", name)
		}
	}
}

func TestCreateBranch(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	err := CreateBranch("new-branch")
	if err != nil {
		t.Fatalf("CreateBranch failed: %v", err)
	}

	exists, err := LocalBranchExists("new-branch")
	if err != nil {
		t.Fatalf("LocalBranchExists failed: %v", err)
	}
	if !exists {
		t.Error("created branch does not exist")
	}
}

func TestDeleteBranch(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	tests := []struct {
		name   string
		branch string
		force  bool
		setup  func()
	}{
		{
			name:   "safe delete merged branch",
			branch: "merged-branch",
			force:  false,
			setup: func() {
				repo.Git("branch", "merged-branch")
			},
		},
		{
			name:   "force delete unmerged branch",
			branch: "unmerged-branch",
			force:  true,
			setup: func() {
				repo.Git("checkout", "-b", "unmerged-branch")
				repo.CreateFile("new-file.txt", "content")
				repo.Commit("commit on unmerged branch")
				repo.Git("checkout", "main")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := DeleteBranch(tt.branch, tt.force)
			if err != nil {
				t.Fatalf("DeleteBranch failed: %v", err)
			}

			exists, err := LocalBranchExists(tt.branch)
			if err != nil {
				t.Fatalf("LocalBranchExists failed: %v", err)
			}
			if exists {
				t.Error("deleted branch still exists")
			}
		})
	}
}

func TestIsBranchMerged(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create and merge a branch
	repo.Git("checkout", "-b", "merged-branch")
	repo.CreateFile("merged.txt", "merged content")
	repo.Commit("commit on merged branch")
	repo.Git("checkout", "main")
	repo.Git("merge", "merged-branch")

	// Create an unmerged branch
	repo.Git("checkout", "-b", "unmerged-branch")
	repo.CreateFile("unmerged.txt", "unmerged content")
	repo.Commit("commit on unmerged branch")
	repo.Git("checkout", "main")

	restore := repo.Chdir()
	defer restore()

	tests := []struct {
		name   string
		branch string
		want   bool
	}{
		{"merged branch", "merged-branch", true},
		{"unmerged branch", "unmerged-branch", false},
		{"main branch", "main", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsBranchMerged(tt.branch)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("IsBranchMerged(%q) = %v, want %v", tt.branch, got, tt.want)
			}
		})
	}
}

func TestGetDefaultBranch(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	// Without remote, should fallback to checking main/master
	branch, err := GetDefaultBranch()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if branch != "main" {
		t.Errorf("GetDefaultBranch() = %q, want %q", branch, "main")
	}
}
