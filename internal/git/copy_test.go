package git

import (
	"fmt"
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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

func TestCopyDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	// Create source directory tree
	for _, f := range []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{"a.txt", "file a", 0644},
		{"sub/b.txt", "file b", 0755},
		{"sub/deep/c.txt", "file c", 0600},
	} {
		p := filepath.Join(srcDir, f.path)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(f.content), f.mode); err != nil {
			t.Fatal(err)
		}
	}

	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	for _, f := range []struct {
		path    string
		content string
		mode    os.FileMode
	}{
		{"a.txt", "file a", 0644},
		{"sub/b.txt", "file b", 0755},
		{"sub/deep/c.txt", "file c", 0600},
	} {
		p := filepath.Join(dstDir, f.path)
		got, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("file %s not found: %v", f.path, err)
			continue
		}
		if string(got) != f.content {
			t.Errorf("%s content = %q, want %q", f.path, got, f.content)
		}
		info, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != f.mode {
			t.Errorf("%s mode = %v, want %v", f.path, info.Mode().Perm(), f.mode)
		}
	}
}

func TestCopyDir_IndependentOfSource(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	p := filepath.Join(srcDir, "file.txt")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Modify source after copy
	if err := os.WriteFile(p, []byte("modified"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(dstDir, "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Errorf("destination was affected by source modification: got %q, want %q", got, "original")
	}
}

func TestCopyFilesToWorktree_DirectoryLevelCopy(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "node_modules/\n")
	repo.Commit("initial commit")

	// Create ignored directory with multiple files (simulating node_modules)
	repo.CreateFile("node_modules/pkg-a/index.js", "module.exports = 'a'")
	repo.CreateFile("node_modules/pkg-a/package.json", `{"name": "pkg-a"}`)
	repo.CreateFile("node_modules/pkg-b/index.js", "module.exports = 'b'")
	repo.CreateFile("node_modules/.package-lock.json", "{}")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	opts := CopyOptions{CopyIgnored: true}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	for _, file := range []string{
		"node_modules/pkg-a/index.js",
		"node_modules/pkg-a/package.json",
		"node_modules/pkg-b/index.js",
		"node_modules/.package-lock.json",
	} {
		dstPath := filepath.Join(dstDir, file)
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("file %q was not copied", file)
		}
	}

	// Verify content
	got, err := os.ReadFile(filepath.Join(dstDir, "node_modules/pkg-a/index.js"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "module.exports = 'a'" {
		t.Errorf("content = %q, want %q", got, "module.exports = 'a'")
	}
}

func TestCopyFilesToWorktree_DirectoryLevelCopy_WithNoCopy(t *testing.T) {
	repo := testutil.NewTestRepo(t)
	repo.CreateFile("README.md", "# Test")
	repo.CreateFile(".gitignore", "node_modules/\n")
	repo.Commit("initial commit")

	repo.CreateFile("node_modules/pkg-a/index.js", "module.exports = 'a'")
	repo.CreateFile("node_modules/pkg-b/index.js", "module.exports = 'b'")

	dstDir := filepath.Join(repo.ParentDir(), "dst")
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		t.Fatalf("failed to create dst dir: %v", err)
	}

	restore := repo.Chdir()
	defer restore()

	// NoCopy excludes some files in node_modules — should fall back to file-by-file
	opts := CopyOptions{
		CopyIgnored: true,
		NoCopy:      []string{"node_modules/pkg-b/"},
	}
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
	if err != nil {
		t.Fatalf("CopyFilesToWorktree failed: %v", err)
	}

	// pkg-a should be copied
	if _, err := os.Stat(filepath.Join(dstDir, "node_modules/pkg-a/index.js")); os.IsNotExist(err) {
		t.Error("node_modules/pkg-a/index.js should have been copied")
	}

	// pkg-b should NOT be copied (NoCopy)
	if _, err := os.Stat(filepath.Join(dstDir, "node_modules/pkg-b/index.js")); !os.IsNotExist(err) {
		t.Error("node_modules/pkg-b/index.js should NOT have been copied")
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
	err := CopyFilesToWorktree(t.Context(), repo.Root, dstDir, opts, nil)
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

// createLargeNodeModules creates a fake node_modules with the given number of
// packages, each containing several files, to simulate a real-world scenario.
func createLargeNodeModules(b *testing.B, baseDir string, numPackages int) []string {
	b.Helper()
	jsContent := []byte("'use strict'; module.exports = function() { return 'hello world'; };")
	jsonContent := []byte(`{"name":"pkg","version":"1.0.0","main":"index.js"}`)
	readmeContent := []byte("# Package\nThis is a test package for benchmarking.")
	licenceContent := []byte("MIT License\nCopyright (c) 2024")
	mapContent := make([]byte, 4096)

	var files []string
	for i := range numPackages {
		pkgName := fmt.Sprintf("pkg-%04d", i)
		pkgDir := filepath.Join(baseDir, "node_modules", pkgName)
		if err := os.MkdirAll(pkgDir, 0755); err != nil {
			b.Fatal(err)
		}
		for _, f := range []struct {
			name    string
			content []byte
		}{
			{"index.js", jsContent},
			{"package.json", jsonContent},
			{"README.md", readmeContent},
			{"LICENCE", licenceContent},
			{"index.js.map", mapContent},
		} {
			p := filepath.Join(pkgDir, f.name)
			if err := os.WriteFile(p, f.content, 0644); err != nil {
				b.Fatal(err)
			}
			rel, _ := filepath.Rel(baseDir, p)
			files = append(files, rel)
		}
	}
	return files
}

func BenchmarkCopyDir_Small(b *testing.B) {
	srcDir := b.TempDir()
	createLargeNodeModules(b, srcDir, 500)
	srcNodeModules := filepath.Join(srcDir, "node_modules")

	b.ResetTimer()
	for b.Loop() {
		dstDir := filepath.Join(b.TempDir(), "node_modules")
		if err := copyDir(srcNodeModules, dstDir); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCopyFileByFile_Small(b *testing.B) {
	srcDir := b.TempDir()
	files := createLargeNodeModules(b, srcDir, 500)

	b.ResetTimer()
	for b.Loop() {
		dstDir := b.TempDir()
		for _, file := range files {
			src := filepath.Join(srcDir, file)
			dst := filepath.Join(dstDir, file)
			if err := copyFile(src, dst); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkCopyDir_RealNodeModules(b *testing.B) {
	src := os.Getenv("BENCH_NODE_MODULES")
	if src == "" {
		b.Skip("set BENCH_NODE_MODULES to a real node_modules path to run this benchmark")
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		b.Skip("BENCH_NODE_MODULES path not found")
	}

	b.ResetTimer()
	for b.Loop() {
		dst := filepath.Join(b.TempDir(), "node_modules")
		if err := copyDir(src, dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCopyFileByFile_RealNodeModules(b *testing.B) {
	src := os.Getenv("BENCH_NODE_MODULES")
	if src == "" {
		b.Skip("set BENCH_NODE_MODULES to a real node_modules path to run this benchmark")
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		b.Skip("BENCH_NODE_MODULES path not found")
	}

	var files []string
	if err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, _ := filepath.Rel(src, path)
			files = append(files, rel)
		}
		return nil
	}); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		dstBase := b.TempDir()
		for _, file := range files {
			srcFile := filepath.Join(src, file)
			dstFile := filepath.Join(dstBase, file)
			if err := copyFile(srcFile, dstFile); err != nil {
				b.Fatal(err)
			}
		}
	}
}
