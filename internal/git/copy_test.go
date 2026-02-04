package git

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestCopyFilesToWorktree_NoCopy(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n*.log\nvendor/\nconfig/local.yml\n")
	repo.Commit("initial commit")

	// Create ignored files
	repo.CreateFile(".env", "SECRET=value")
	repo.CreateFile("app.log", "log content")
	repo.CreateFile("vendor/github.com/foo/bar.go", "package foo")
	repo.CreateFile("config/local.yml", "local: true")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Copy ignored files but exclude *.log and vendor/ (gitignore pattern)
	opts := CopyOptions{
		CopyIgnored: true,
		NoCopy:      []string{"*.log", "vendor/"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// .env should be copied
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); os.IsNotExist(err) {
		t.Error(".env should have been copied")
	}

	// config/local.yml should be copied (not in NoCopy)
	if _, err := os.Stat(filepath.Join(dstDir, "config/local.yml")); os.IsNotExist(err) {
		t.Error("config/local.yml should have been copied")
	}

	// app.log should NOT be copied (matches *.log)
	if _, err := os.Stat(filepath.Join(dstDir, "app.log")); !os.IsNotExist(err) {
		t.Error("app.log should NOT have been copied")
	}

	// vendor/github.com/foo/bar.go should NOT be copied (matches vendor/)
	if _, err := os.Stat(filepath.Join(dstDir, "vendor/github.com/foo/bar.go")); !os.IsNotExist(err) {
		t.Error("vendor/github.com/foo/bar.go should NOT have been copied")
	}
}

func TestCopyFilesToWorktree_NoCopy_GitignorePatterns(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "*.secret\nbuild/\ntemp/\n")
	repo.Commit("initial commit")

	// Create ignored files with various patterns
	repo.CreateFile("api.secret", "api key")
	repo.CreateFile("db.secret", "db password")
	repo.CreateFile("build/output.js", "compiled")
	repo.CreateFile("build/nested/file.js", "nested compiled")
	repo.CreateFile("temp/cache.txt", "cache")
	repo.CreateFile("src/temp/data.txt", "not excluded")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Exclude only build/ directory using gitignore pattern
	opts := CopyOptions{
		CopyIgnored: true,
		NoCopy:      []string{"build/"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// *.secret files should be copied (not in NoCopy)
	for _, file := range []string{"api.secret", "db.secret"} {
		if _, err := os.Stat(filepath.Join(dstDir, file)); os.IsNotExist(err) {
			t.Errorf("%s should have been copied", file)
		}
	}

	// temp/ files should be copied (not in NoCopy)
	if _, err := os.Stat(filepath.Join(dstDir, "temp/cache.txt")); os.IsNotExist(err) {
		t.Error("temp/cache.txt should have been copied")
	}

	// build/ files should NOT be copied
	for _, file := range []string{"build/output.js", "build/nested/file.js"} {
		if _, err := os.Stat(filepath.Join(dstDir, file)); !os.IsNotExist(err) {
			t.Errorf("%s should NOT have been copied", file)
		}
	}
}

func TestCopyFilesToWorktree_Copy(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "*.code-workspace\n.vscode/\n.env\n")
	repo.Commit("initial commit")

	// Create ignored files
	repo.CreateFile("project.code-workspace", `{"folders": []}`)
	repo.CreateFile(".vscode/settings.json", `{"editor.tabSize": 2}`)
	repo.CreateFile(".env", "SECRET=value")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Copy only specific ignored files using Copy patterns (CopyIgnored=false)
	opts := CopyOptions{
		CopyIgnored: false,
		Copy:        []string{"*.code-workspace"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// project.code-workspace should be copied (matches Copy pattern)
	if _, err := os.Stat(filepath.Join(dstDir, "project.code-workspace")); os.IsNotExist(err) {
		t.Error("project.code-workspace should have been copied")
	}

	// .vscode/settings.json should NOT be copied (not in Copy patterns)
	if _, err := os.Stat(filepath.Join(dstDir, ".vscode/settings.json")); !os.IsNotExist(err) {
		t.Error(".vscode/settings.json should NOT have been copied")
	}

	// .env should NOT be copied (not in Copy patterns)
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); !os.IsNotExist(err) {
		t.Error(".env should NOT have been copied")
	}
}

func TestCopyFilesToWorktree_Copy_WithNoCopy(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "*.code-workspace\n.env\n")
	repo.Commit("initial commit")

	// Create ignored files
	repo.CreateFile("project.code-workspace", `{"folders": []}`)
	repo.CreateFile("other.code-workspace", `{"folders": []}`)
	repo.CreateFile(".env", "SECRET=value")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Copy *.code-workspace but exclude other.code-workspace via NoCopy
	// NoCopy should take precedence over Copy
	opts := CopyOptions{
		CopyIgnored: false,
		Copy:        []string{"*.code-workspace"},
		NoCopy:      []string{"other.code-workspace"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// project.code-workspace should be copied
	if _, err := os.Stat(filepath.Join(dstDir, "project.code-workspace")); os.IsNotExist(err) {
		t.Error("project.code-workspace should have been copied")
	}

	// other.code-workspace should NOT be copied (NoCopy takes precedence)
	if _, err := os.Stat(filepath.Join(dstDir, "other.code-workspace")); !os.IsNotExist(err) {
		t.Error("other.code-workspace should NOT have been copied (NoCopy takes precedence)")
	}

	// .env should NOT be copied
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); !os.IsNotExist(err) {
		t.Error(".env should NOT have been copied")
	}
}

func TestCopyFilesToWorktree_Copy_MultiplePatterns(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "*.code-workspace\n.vscode/\n.idea/\n.env\n")
	repo.Commit("initial commit")

	// Create ignored files
	repo.CreateFile("project.code-workspace", `{"folders": []}`)
	repo.CreateFile(".vscode/settings.json", `{"editor.tabSize": 2}`)
	repo.CreateFile(".idea/workspace.xml", "<project/>")
	repo.CreateFile(".env", "SECRET=value")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Copy multiple patterns
	opts := CopyOptions{
		CopyIgnored: false,
		Copy:        []string{"*.code-workspace", ".vscode/"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// project.code-workspace should be copied
	if _, err := os.Stat(filepath.Join(dstDir, "project.code-workspace")); os.IsNotExist(err) {
		t.Error("project.code-workspace should have been copied")
	}

	// .vscode/settings.json should be copied
	if _, err := os.Stat(filepath.Join(dstDir, ".vscode/settings.json")); os.IsNotExist(err) {
		t.Error(".vscode/settings.json should have been copied")
	}

	// .idea/workspace.xml should NOT be copied (not in Copy patterns)
	if _, err := os.Stat(filepath.Join(dstDir, ".idea/workspace.xml")); !os.IsNotExist(err) {
		t.Error(".idea/workspace.xml should NOT have been copied")
	}

	// .env should NOT be copied
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); !os.IsNotExist(err) {
		t.Error(".env should NOT have been copied")
	}
}

func TestCopyFilesToWorktree_Copy_WithCopyIgnored(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "*.code-workspace\n.env\n")
	repo.Commit("initial commit")

	// Create ignored files
	repo.CreateFile("project.code-workspace", `{"folders": []}`)
	repo.CreateFile(".env", "SECRET=value")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Both CopyIgnored and Copy are set
	opts := CopyOptions{
		CopyIgnored: true,
		Copy:        []string{"*.code-workspace"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// Both files should be copied (CopyIgnored copies all, Copy adds workspace)
	if _, err := os.Stat(filepath.Join(dstDir, "project.code-workspace")); os.IsNotExist(err) {
		t.Error("project.code-workspace should have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); os.IsNotExist(err) {
		t.Error(".env should have been copied")
	}
}

// TestCopyFilesToWorktree_Copy_MatchesUntrackedFiles confirms that
// the Copy option matches both ignored AND untracked files.
func TestCopyFilesToWorktree_Copy_MatchesUntrackedFiles(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n") // Only .env is ignored
	repo.Commit("initial commit")

	// Create an untracked file (not in .gitignore, not committed)
	repo.CreateFile("untracked.txt", "untracked content")
	// Create an ignored file for comparison
	repo.CreateFile(".env", "SECRET=value")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Copy using a pattern that matches the untracked file
	opts := CopyOptions{
		CopyIgnored: false,
		Copy:        []string{"untracked.txt"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// untracked.txt SHOULD be copied because Copy matches untracked files too
	if _, err := os.Stat(filepath.Join(dstDir, "untracked.txt")); os.IsNotExist(err) {
		t.Error("untracked.txt should have been copied")
	}

	// .env should NOT be copied (doesn't match the Copy pattern)
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); !os.IsNotExist(err) {
		t.Error(".env should NOT have been copied (doesn't match Copy pattern)")
	}
}

func TestCopyFile_PreservesTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")

	if err := os.WriteFile(srcPath, []byte("content"), 0600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	// Set an old modification time on the source file
	oldTime := time.Date(2020, 6, 15, 10, 30, 0, 0, time.UTC)
	if err := os.Chtimes(srcPath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set source file time: %v", err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("failed to stat destination file: %v", err)
	}

	if !dstInfo.ModTime().Equal(oldTime) {
		t.Errorf("destination mtime = %v, want %v", dstInfo.ModTime(), oldTime)
	}
}

func TestCopyFilesToWorktree_ExcludeDirs(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", ".env\n.worktrees/\n")
	repo.Commit("initial commit")

	// Create ignored files
	repo.CreateFile(".env", "SECRET=value")
	// Create files in the directory that should be excluded (simulating worktrees basedir)
	repo.CreateFile(".worktrees/existing-wt/README.md", "# Existing worktree")
	repo.CreateFile(".worktrees/existing-wt/.env", "WT_SECRET=value")
	repo.CreateFile(".worktrees/.gitignore", "*\n")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// Copy ignored files but exclude .worktrees directory
	opts := CopyOptions{
		CopyIgnored: true,
		ExcludeDirs: []string{filepath.Join(repo.Root, ".worktrees")},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// .env should be copied (not in ExcludeDirs)
	if _, err := os.Stat(filepath.Join(dstDir, ".env")); os.IsNotExist(err) {
		t.Error(".env should have been copied")
	}

	// Files inside .worktrees should NOT be copied (in ExcludeDirs)
	if _, err := os.Stat(filepath.Join(dstDir, ".worktrees/existing-wt/README.md")); !os.IsNotExist(err) {
		t.Error(".worktrees/existing-wt/README.md should NOT have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, ".worktrees/existing-wt/.env")); !os.IsNotExist(err) {
		t.Error(".worktrees/existing-wt/.env should NOT have been copied")
	}
	if _, err := os.Stat(filepath.Join(dstDir, ".worktrees/.gitignore")); !os.IsNotExist(err) {
		t.Error(".worktrees/.gitignore should NOT have been copied")
	}
}
