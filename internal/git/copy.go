package git

import (
	"context"
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
func CopyFilesToWorktree(ctx context.Context, srcRoot, dstRoot string, opts CopyOptions) error {
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
			// Parse pattern from git root (empty domain means root)
			patterns = append(patterns, gitignore.ParsePattern(p, nil))
		}
		noCopyMatcher = gitignore.NewMatcher(patterns)
	}

	// Remove duplicates
	seen := make(map[string]struct{})
	for _, file := range files {
		if _, exists := seen[file]; exists {
			continue
		}
		seen[file] = struct{}{}

		// Skip files inside ExcludeDirs
		src := filepath.Join(srcRoot, file)
		shouldSkip := false
		for _, excludeDir := range opts.ExcludeDirs {
			// Check if src is inside excludeDir
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
			// Split file path into components for gitignore matching
			pathComponents := strings.Split(file, string(filepath.Separator))
			isDir := false // files from git ls-files are always files
			if noCopyMatcher.Match(pathComponents, isDir) {
				continue
			}
		}

		dst := filepath.Join(dstRoot, file)

		if err := copyFile(src, dst); err != nil {
			// Skip files that fail to copy (e.g., permission issues)
			continue
		}
	}

	return nil
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

