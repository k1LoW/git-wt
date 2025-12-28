package git

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyOptions holds the copy configuration.
type CopyOptions struct {
	CopyIgnored   bool
	CopyUntracked bool
	CopyModified  bool
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
		untracked, err := listUntrackedFiles(ctx, srcRoot)
		if err != nil {
			return err
		}
		files = append(files, untracked...)
	}

	if opts.CopyModified {
		modified, err := listModifiedFiles(ctx, srcRoot)
		if err != nil {
			return err
		}
		files = append(files, modified...)
	}

	// Remove duplicates
	seen := make(map[string]struct{})
	for _, file := range files {
		if _, exists := seen[file]; exists {
			continue
		}
		seen[file] = struct{}{}

		src := filepath.Join(srcRoot, file)
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

// listUntrackedFiles returns untracked files (not ignored).
func listUntrackedFiles(ctx context.Context, root string) ([]string, error) {
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

// listModifiedFiles returns tracked files with modifications.
func listModifiedFiles(ctx context.Context, root string) ([]string, error) {
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

func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Skip directories
	if srcInfo.IsDir() {
		return nil
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	// Preserve file permissions
	return os.Chmod(dst, srcInfo.Mode())
}
