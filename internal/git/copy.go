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

	// Deduplicate and filter files
	seen := make(map[string]struct{})
	var filtered []string
	for _, file := range files {
		if _, exists := seen[file]; exists {
			continue
		}
		seen[file] = struct{}{}

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

		if noCopyMatcher != nil {
			pathComponents := strings.Split(file, string(filepath.Separator))
			if noCopyMatcher.Match(pathComponents, false) {
				continue
			}
		}

		filtered = append(filtered, file)
	}

	// Group files by top-level directory and attempt directory-level copy
	dirFiles := groupByTopLevelDir(filtered)
	copiedDirs := make(map[string]struct{})

	for dir, dirFileList := range dirFiles {
		srcDir := filepath.Join(srcRoot, dir)
		info, err := os.Stat(srcDir)
		if err != nil || !info.IsDir() {
			continue
		}

		allFilesInDir, err := countFilesInDir(srcDir)
		if err != nil {
			continue
		}

		if len(dirFileList) >= allFilesInDir {
			dstDir := filepath.Join(dstRoot, dir)
			if err := copyDir(srcDir, dstDir); err != nil {
				if warn != nil {
					fmt.Fprintf(warn, "warning: failed to copy directory %s, falling back to file-by-file: %v\n", dir, err)
				}
				continue
			}
			copiedDirs[dir] = struct{}{}
		}
	}

	// Copy remaining files that were not covered by directory-level copy
	for _, file := range filtered {
		topDir := topLevelDir(file)
		if topDir != "" {
			if _, ok := copiedDirs[topDir]; ok {
				continue
			}
		}

		src := filepath.Join(srcRoot, file)
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

// groupByTopLevelDir groups files by their top-level directory.
// Files at the root level are not included.
func groupByTopLevelDir(files []string) map[string][]string {
	groups := make(map[string][]string)
	for _, file := range files {
		dir := topLevelDir(file)
		if dir != "" {
			groups[dir] = append(groups[dir], file)
		}
	}
	return groups
}

// countFilesInDir counts all regular files in a directory recursively.
func countFilesInDir(dir string) (int, error) {
	count := 0
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count, err
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

