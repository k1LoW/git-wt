package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/git-wt/testutil"
)

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

func TestGitConfig(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")
	repo.Git("config", "test.key", "test-value")
	repo.Git("config", "--add", "test.multi", "value1")
	repo.Git("config", "--add", "test.multi", "value2")

	restore := repo.Chdir()
	defer restore()

	tests := []struct {
		name    string
		key     string
		want    []string
		wantErr bool
	}{
		{"existing key", "test.key", []string{"test-value"}, false},
		{"non-existing key", "test.nonexistent", nil, false},
		{"multiple values", "test.multi", []string{"value1", "value2"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GitConfig(t.Context(), tt.key)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GitConfig(%q) error = %v, wantErr %v", tt.key, err, tt.wantErr)
			}
			if len(got) != len(tt.want) {
				t.Errorf("GitConfig(%q) = %v, want %v", tt.key, got, tt.want) //nostyle:errorstrings
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GitConfig(%q)[%d] = %q, want %q", tt.key, i, got[i], tt.want[i]) //nostyle:errorstrings
				}
			}
		})
	}
}

func TestRepoRoot(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create a subdirectory
	repo.CreateFile("subdir/file.txt", "content")
	repo.Commit("add subdir")

	restore := repo.Chdir()
	defer restore()

	root, err := RepoRoot(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if root != repo.Root {
		t.Errorf("RepoRoot() = %q, want %q", root, repo.Root) //nostyle:errorstrings
	}

	// Test from subdirectory
	subdir := filepath.Join(repo.Root, "subdir")
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("failed to chdir to subdir: %v", err)
	}

	root, err = RepoRoot(t.Context())
	if err != nil {
		t.Fatalf("unexpected error from subdir: %v", err)
	}

	if root != repo.Root {
		t.Errorf("RepoRoot() from subdir = %q, want %q", root, repo.Root) //nostyle:errorstrings
	}
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

func TestLoadConfig(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	// Test with custom values
	repo.Git("config", "wt.basedir", "../custom-worktrees")
	repo.Git("config", "wt.copyignored", "true")
	repo.Git("config", "wt.copyuntracked", "false")
	repo.Git("config", "wt.copymodified", "true")

	cfg, err := LoadConfig(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BaseDir != "../custom-worktrees" {
		t.Errorf("LoadConfig().BaseDir = %q, want %q", cfg.BaseDir, "../custom-worktrees") //nostyle:errorstrings
	}
	if !cfg.CopyIgnored {
		t.Errorf("LoadConfig().CopyIgnored = %v, want true", cfg.CopyIgnored) //nostyle:errorstrings
	}
	if cfg.CopyUntracked {
		t.Errorf("LoadConfig().CopyUntracked = %v, want false", cfg.CopyUntracked) //nostyle:errorstrings
	}
	if !cfg.CopyModified {
		t.Errorf("LoadConfig().CopyModified = %v, want true", cfg.CopyModified) //nostyle:errorstrings
	}

	// Test with explicit default pattern
	repo.Git("config", "wt.basedir", "../{gitroot}-wt")
	repo.Git("config", "--unset", "wt.copyignored")
	repo.Git("config", "--unset", "wt.copyuntracked")
	repo.Git("config", "--unset", "wt.copymodified")

	cfg, err = LoadConfig(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.BaseDir != "../{gitroot}-wt" {
		t.Errorf("LoadConfig().BaseDir = %q, want %q", cfg.BaseDir, "../{gitroot}-wt") //nostyle:errorstrings
	}
	if cfg.CopyIgnored {
		t.Errorf("LoadConfig().CopyIgnored default = %v, want false", cfg.CopyIgnored) //nostyle:errorstrings
	}
	if cfg.CopyUntracked {
		t.Errorf("LoadConfig().CopyUntracked default = %v, want false", cfg.CopyUntracked) //nostyle:errorstrings
	}
	if cfg.CopyModified {
		t.Errorf("LoadConfig().CopyModified default = %v, want false", cfg.CopyModified) //nostyle:errorstrings
	}
	if cfg.NoCd {
		t.Errorf("LoadConfig().NoCd default = %v, want false", cfg.NoCd) //nostyle:errorstrings
	}

	// Test NoCd setting
	repo.Git("config", "wt.nocd", "true")

	cfg, err = LoadConfig(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.NoCd {
		t.Errorf("LoadConfig().NoCd = %v, want true", cfg.NoCd) //nostyle:errorstrings
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
			got, err := ExpandPath(t.Context(), tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.path, got, tt.want) //nostyle:errorstrings
			}
		})
	}
}

func TestWorktreePathFor(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	restore := repo.Chdir()
	defer restore()

	// Test with default pattern basedir
	path, err := WorktreePathFor(t.Context(), "../{gitroot}-wt", "feature-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: parent_dir/repo-wt/feature-branch
	expectedDir := filepath.Clean(filepath.Join(repo.Root, "../repo-wt/feature-branch"))
	if path != expectedDir {
		t.Errorf("WorktreePathFor(\"feature-branch\") = %q, want %q", path, expectedDir) //nostyle:errorstrings
	}

	// Test with custom basedir
	path, err = WorktreePathFor(t.Context(), "../{gitroot}-worktrees", "feature-branch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDir = filepath.Clean(filepath.Join(repo.Root, "../repo-worktrees/feature-branch"))
	if path != expectedDir {
		t.Errorf("WorktreePathFor(\"feature-branch\") with custom basedir = %q, want %q", path, expectedDir) //nostyle:errorstrings
	}
}

func TestExpandBaseDir(t *testing.T) {
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
		name    string
		baseDir string
		want    string
	}{
		{
			name:    "relative path with gitroot",
			baseDir: "../{gitroot}-wt",
			want:    filepath.Clean(filepath.Join(repo.Root, "../repo-wt")),
		},
		{
			name:    "absolute path",
			baseDir: "/tmp/worktrees",
			want:    "/tmp/worktrees",
		},
		{
			name:    "tilde expansion",
			baseDir: "~/worktrees",
			want:    filepath.Join(homeDir, "worktrees"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandBaseDir(t.Context(), tt.baseDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExpandBaseDir(%q) = %q, want %q", tt.baseDir, got, tt.want) //nostyle:errorstrings
			}
		})
	}
}

func TestIsBaseDirConfigured(t *testing.T) {
	t.Run("not configured", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")

		t.Cleanup(repo.Chdir())
		t.Setenv("HOME", t.TempDir())

		configured, err := IsBaseDirConfigured(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if configured {
			t.Error("IsBaseDirConfigured() = true, want false")
		}
	})

	t.Run("configured", func(t *testing.T) {
		repo := testutil.NewTestRepo(t)
		repo.CreateFile("README.md", "# Test")
		repo.Commit("initial commit")
		repo.Git("config", "wt.basedir", "../custom-wt")

		t.Cleanup(repo.Chdir())
		t.Setenv("HOME", t.TempDir())

		configured, err := IsBaseDirConfigured(t.Context())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !configured {
			t.Error("IsBaseDirConfigured() = false, want true")
		}
	})
}

func TestSetConfig(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	t.Cleanup(repo.Chdir())
	t.Setenv("HOME", t.TempDir())

	if err := SetConfig(t.Context(), "wt.basedir", "../test-wt"); err != nil {
		t.Fatalf("SetConfig() error = %v", err)
	}

	values, err := GitConfig(t.Context(), "wt.basedir")
	if err != nil {
		t.Fatalf("GitConfig() error = %v", err)
	}
	if len(values) != 1 || values[0] != "../test-wt" {
		t.Errorf("GitConfig() = %v, want [../test-wt]", values) //nostyle:errorstrings
	}
}
