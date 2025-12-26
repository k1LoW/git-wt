package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/git-wt/testutil"
)

func TestGetConfig(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")
	repo.Git("config", "test.key", "test-value")

	restore := repo.Chdir()
	defer restore()

	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{"existing key", "test.key", "test-value", false},
		{"non-existing key", "test.nonexistent", "", false},
		{"user.email", "user.email", "test@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConfig(tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetConfig(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("GetConfig(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestGetRepoRoot(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a subdirectory
	repo.CreateFile("subdir/file.txt", "content")
	repo.Commit("add subdir")

	restore := repo.Chdir()
	defer restore()

	root, err := GetRepoRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root != repo.Root {
		t.Errorf("GetRepoRoot() = %q, want %q", root, repo.Root)
	}

	// Test from subdirectory
	subdir := filepath.Join(repo.Root, "subdir")
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("failed to chdir to subdir: %v", err)
	}

	root, err = GetRepoRoot()
	if err != nil {
		t.Fatalf("unexpected error from subdir: %v", err)
	}

	if root != repo.Root {
		t.Errorf("GetRepoRoot() from subdir = %q, want %q", root, repo.Root)
	}
}

func TestGetRepoName(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	name, err := GetRepoName()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The repo is created in a temp directory named "repo"
	if name != "repo" {
		t.Errorf("GetRepoName() = %q, want %q", name, "repo")
	}
}

func TestGetBaseDir(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	// Test with custom value (to avoid being affected by global config)
	repo.Git("config", "wt.basedir", "../custom-worktrees")

	got, err := GetBaseDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "../custom-worktrees" {
		t.Errorf("GetBaseDir() = %q, want %q", got, "../custom-worktrees")
	}

	// Test default value by unsetting local config
	// Note: This may still be affected by global config, so we set an explicit value instead
	repo.Git("config", "wt.basedir", "../{gitroot}-wt")

	got, err = GetBaseDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "../{gitroot}-wt" {
		t.Errorf("GetBaseDir() with default pattern = %q, want %q", got, "../{gitroot}-wt")
	}
}

func TestExpandPath(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "relative path",
			path: "../sibling",
			want: filepath.Clean(filepath.Join(repo.Root, "../sibling")),
		},
		{
			name: "absolute path",
			path: "/absolute/path",
			want: "/absolute/path",
		},
		{
			name: "tilde expansion",
			path: "~/projects",
			want: filepath.Join(homeDir, "projects"),
		},
		{
			name: "tilde only",
			path: "~",
			want: homeDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetWorktreePath(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	// Set explicit basedir to avoid being affected by global config
	repo.Git("config", "wt.basedir", "../{gitroot}-wt")

	// Test with default pattern basedir
	path, err := GetWorktreePath("feature-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: parent_dir/repo-wt/feature-branch
	expectedDir := filepath.Clean(filepath.Join(repo.Root, "../repo-wt/feature-branch"))
	if path != expectedDir {
		t.Errorf("GetWorktreePath(\"feature-branch\") = %q, want %q", path, expectedDir)
	}

	// Test with custom basedir
	repo.Git("config", "wt.basedir", "../{gitroot}-worktrees")

	path, err = GetWorktreePath("feature-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDir = filepath.Clean(filepath.Join(repo.Root, "../repo-worktrees/feature-branch"))
	if path != expectedDir {
		t.Errorf("GetWorktreePath(\"feature-branch\") with custom basedir = %q, want %q", path, expectedDir)
	}
}
