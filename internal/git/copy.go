package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// CopyOptions holds the copy configuration.
type CopyOptions struct {
	CopyIgnored   bool
	CopyUntracked bool
	CopyModified  bool
	NoCopy        []string
	Copy          []string
	Symlink       []string // Patterns for directories to symlink instead of copy (gitignore syntax)
	ExcludeDirs   []string // Directories to exclude from copying (absolute paths)
}

// CopyFilesToWorktree copies files to the new worktree based on options.
// If w is non-nil, warnings about files that fail to copy are written to it.
func CopyFilesToWorktree(ctx context.Context, srcRoot, dstRoot string, opts CopyOptions, warn io.Writer) error {
	var files []string

	if opts.CopyIgnored {
		ignored, err := listIgnoredFiles(ctx, srcRoot)
		if err != nil {
			return err
		}
		files = append(files, ignored...)
	}

	if opts.CopyUntracked {
		untracked, err := ListUntrackedFiles(ctx, srcRoot)
		if err != nil {
			return err
		}
		files = append(files, untracked...)
	}

	if opts.CopyModified {
		modified, err := ListModifiedFiles(ctx, srcRoot)
		if err != nil {
			return err
		}
		files = append(files, modified...)
	}

	// Add files matching Copy patterns (from ignored files)
	if len(opts.Copy) > 0 {
		copyFiles, err := listFilesMatchingCopyPatterns(ctx, srcRoot, opts.Copy)
		if err != nil {
			return err
		}
		files = append(files, copyFiles...)
	}

	// Build NoCopy matcher using gitignore patterns
	var noCopyMatcher gitignore.Matcher
	if len(opts.NoCopy) > 0 {
		var patterns []gitignore.Pattern
		for _, p := range opts.NoCopy {
			patterns = append(patterns, gitignore.ParsePattern(p, nil))
		}
		noCopyMatcher = gitignore.NewMatcher(patterns)
	}

	// Build Symlink matcher using gitignore patterns
	var symlinkMatcher gitignore.Matcher
	if len(opts.Symlink) > 0 {
		var patterns []gitignore.Pattern
		for _, p := range opts.Symlink {
			patterns = append(patterns, gitignore.ParsePattern(p, nil))
		}
		symlinkMatcher = gitignore.NewMatcher(patterns)
	}

	// Create symlinks for matching top-level directories before file-by-file copy
	symlinkedDirs := make(map[string]struct{})
	if symlinkMatcher != nil {
		dirs := collectTopLevelDirs(files)
		for _, dir := range dirs {
			pathComponents := []string{dir}
			if !symlinkMatcher.Match(pathComponents, true) {
				continue
			}

			srcDir := filepath.Join(srcRoot, dir)
			info, err := os.Lstat(srcDir)
			if err != nil || !info.IsDir() {
				continue
			}

			dstDir := filepath.Join(dstRoot, dir)
			if err := os.MkdirAll(filepath.Dir(dstDir), 0755); err != nil {
				if warn != nil {
					fmt.Fprintf(warn, "warning: failed to create parent for symlink %s: %v\n", dir, err)
				}
				continue
			}
			if err := os.Symlink(srcDir, dstDir); err != nil {
				if warn != nil {
					fmt.Fprintf(warn, "warning: failed to symlink %s: %v\n", dir, err)
				}
				continue
			}
			symlinkedDirs[dir] = struct{}{}
		}
	}

	// Deduplicate and filter files
	seen := make(map[string]struct{})
	for _, file := range files {
		if _, exists := seen[file]; exists {
			continue
		}
		seen[file] = struct{}{}

		// Skip files inside symlinked directories
		topDir := topLevelDir(file)
		if topDir != "" {
			if _, ok := symlinkedDirs[topDir]; ok {
				continue
			}
		}

		// Skip files inside ExcludeDirs
		src := filepath.Join(srcRoot, file)
		shouldSkip := false
		for _, excludeDir := range opts.ExcludeDirs {
			rel, err := filepath.Rel(excludeDir, src)
			if err == nil && !strings.HasPrefix(rel, "..") {
				shouldSkip = true
				break
			}
		}
		if shouldSkip {
			continue
		}

		// Skip files matching NoCopy patterns
		if noCopyMatcher != nil {
			pathComponents := strings.Split(file, string(filepath.Separator))
			if noCopyMatcher.Match(pathComponents, false) {
				continue
			}
		}

		dst := filepath.Join(dstRoot, file)

		if err := copyFile(src, dst); err != nil {
			if warn != nil {
				fmt.Fprintf(warn, "warning: failed to copy %s: %v\n", file, err)
			}
			continue
		}
	}

	return nil
}

// topLevelDir returns the first path component if the file is inside a directory,
// or empty string if the file is at the root level.
func topLevelDir(file string) string {
	dir := filepath.Dir(file)
	if dir == "." {
		return ""
	}
	parts := strings.SplitN(dir, string(filepath.Separator), 2)
	return parts[0]
}

// collectTopLevelDirs returns unique top-level directory names from a file list.
func collectTopLevelDirs(files []string) []string {
	seen := make(map[string]struct{})
	var dirs []string
	for _, file := range files {
		dir := topLevelDir(file)
		if dir == "" {
			continue
		}
		if _, exists := seen[dir]; !exists {
			seen[dir] = struct{}{}
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

// listIgnoredFiles returns files ignored by .gitignore.
func listIgnoredFiles(ctx context.Context, root string) ([]string, error) {
	cmd, err := gitCommand(ctx, "ls-files", "--others", "--ignored", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFileList(string(out)), nil
}

// ListUntrackedFiles returns untracked files (not ignored).
func ListUntrackedFiles(ctx context.Context, root string) ([]string, error) {
	cmd, err := gitCommand(ctx, "ls-files", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFileList(string(out)), nil
}

// ListModifiedFiles returns tracked files with modifications.
func ListModifiedFiles(ctx context.Context, root string) ([]string, error) {
	cmd, err := gitCommand(ctx, "ls-files", "--modified")
	if err != nil {
		return nil, err
	}
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseFileList(string(out)), nil
}

func parseFileList(out string) []string {
	var files []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Skip .git directory
		if strings.HasPrefix(line, ".git/") {
			continue
		}
		files = append(files, line)
	}
	return files
}

// listFilesMatchingCopyPatterns returns ignored and untracked files that match the given patterns.
func listFilesMatchingCopyPatterns(ctx context.Context, root string, patterns []string) ([]string, error) {
	// Get ignored files
	ignored, err := listIgnoredFiles(ctx, root)
	if err != nil {
		return nil, err
	}

	// Get untracked files
	untracked, err := ListUntrackedFiles(ctx, root)
	if err != nil {
		return nil, err
	}

	// Combine both lists
	allFiles := append(ignored, untracked...)

	// Build matcher from patterns
	var matcherPatterns []gitignore.Pattern
	for _, p := range patterns {
		matcherPatterns = append(matcherPatterns, gitignore.ParsePattern(p, nil))
	}
	matcher := gitignore.NewMatcher(matcherPatterns)

	// Filter files matching patterns
	var result []string
	for _, file := range allFiles {
		pathComponents := strings.Split(file, string(filepath.Separator))
		if matcher.Match(pathComponents, false) {
			result = append(result, file)
		}
	}

	return result, nil
}

