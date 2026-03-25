//go:build !darwin && !linux

// Default implementation for platforms without specialised CoW support.

package git

import (
	"io"
	"os"
	"path/filepath"
)

func copyFile(src, dst string) (err error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return nil
	}

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
	defer func() {
		if cerr := out.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}

// copyDir copies a directory tree using file-by-file copy.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		return copyFile(path, target)
	})
}
