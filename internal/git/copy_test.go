package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/k1LoW/git-wt/testutil"
)

func TestCopyFilesToWorktree_Ignored(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n*.log\n")
	repo.Commit("initial commit")

	// Create ignored files
	repo.CreateFile(".env", "SECRET=value")
	repo.CreateFile("app.log", "log content")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	opts := CopyOptions{CopyIgnored: true}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// Check that ignored files were copied
	for _, file := range []string{".env", "app.log"} {
		dstPath := filepath.Join(dstDir, file)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("ignored file %q was not copied", file)
		}
	}

	// Check content
	content, err := os.ReadFile(filepath.Join(dstDir, ".env"))
	if err != nil {
		t.Fatalf("failed to read .env: %v", err)
	}
	if string(content) != "SECRET=value" {
		t.Errorf(".env content = %q, want %q", string(content), "SECRET=value")
	}
}

func TestCopyFilesToWorktree_Untracked(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.Commit("initial commit")

	// Create untracked file (not in .gitignore, not committed)
	repo.CreateFile("untracked.txt", "untracked content")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	opts := CopyOptions{CopyUntracked: true}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// Check that untracked file was copied
	dstPath := filepath.Join(dstDir, "untracked.txt")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("untracked file was not copied")
	}

	// Check content
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read untracked.txt: %v", err)
	}
	if string(content) != "untracked content" {
		t.Errorf("untracked.txt content = %q, want %q", string(content), "untracked content")
	}
}

func TestCopyFilesToWorktree_Modified(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile("tracked.txt", "original content")
	repo.Commit("initial commit")

	// Modify tracked file
	repo.CreateFile("tracked.txt", "modified content")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	opts := CopyOptions{CopyModified: true}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// Check that modified file was copied
	dstPath := filepath.Join(dstDir, "tracked.txt")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("modified file was not copied")
	}

	// Check content (should be modified version)
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read tracked.txt: %v", err)
	}
	if string(content) != "modified content" {
		t.Errorf("tracked.txt content = %q, want %q", string(content), "modified content")
	}
}

func TestCopyFilesToWorktree_NoOptions(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n")
	repo.Commit("initial commit")

	// Create various files
	repo.CreateFile(".env", "SECRET=value")
	repo.CreateFile("untracked.txt", "untracked")
	repo.CreateFile("README.md", "modified")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// No copy options enabled
	opts := CopyOptions{}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// Check that no files were copied
	for _, file := range []string{".env", "untracked.txt", "README.md"} {
		dstPath := filepath.Join(dstDir, file)
		if _, err := os.Stat(dstPath); !os.IsNotExist(err) {
			t.Errorf("file %q should not have been copied", file)
		}
	}
}

func TestCopyFilesToWorktree_Subdirectory(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "config/local.yml\n")
	repo.Commit("initial commit")

	// Create ignored file in subdirectory
	repo.CreateFile("config/local.yml", "local: true")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	opts := CopyOptions{CopyIgnored: true}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// Check that file in subdirectory was copied
	dstPath := filepath.Join(dstDir, "config/local.yml")
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("ignored file in subdirectory was not copied")
	}

	// Check content
	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read config/local.yml: %v", err)
	}
	if string(content) != "local: true" {
		t.Errorf("config/local.yml content = %q, want %q", string(content), "local: true")
	}
}
