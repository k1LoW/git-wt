//go:build linux

// Linux implementation using FICLONE ioctl for CoW reflink on supported
// filesystems (Btrfs, XFS with reflink). Falls back to io.Copy which
// internally uses copy_file_range(2) for efficient in-kernel copying.

package git

import (
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

func copyFile(src, dst string) error {
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

	if err := copyFileClone(src, dst, srcInfo); err == nil {
		return nil
	}

	return copyFileTraditional(src, dst, srcInfo)
}

// copyFileClone attempts a CoW reflink via FICLONE ioctl.
func copyFileClone(src, dst string, srcInfo os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if err := unix.IoctlFileClone(int(out.Fd()), int(in.Fd())); err != nil {
		os.Remove(dst)
		return err
	}

	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}

// copyDir copies a directory tree using file-by-file CoW clone with fallback.
// Linux has no directory-level clonefile equivalent, so we walk and clone each file.
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

func copyFileTraditional(src, dst string, srcInfo os.FileInfo) (err error) {
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
