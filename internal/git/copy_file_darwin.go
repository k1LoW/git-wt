//go:build darwin

// macOS implementation using clonefile(2) for APFS Copy-on-Write.
// clonefile creates a lightweight clone that shares data blocks until modified,
// making copies nearly instantaneous regardless of file size.
// Falls back to traditional io.Copy when clonefile fails (non-APFS, cross-device, etc.).

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

	if err := unix.Clonefile(src, dst, unix.CLONE_NOFOLLOW); err == nil {
		return os.Chmod(dst, srcInfo.Mode())
	}

	return copyFileTraditional(src, dst, srcInfo)
}

// copyDir clones an entire directory tree using clonefile(2).
// On APFS this is a single syscall that creates a CoW clone of the whole tree.
// Falls back to recursive file-by-file copy on failure.
func copyDir(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	if err := unix.Clonefile(src, dst, unix.CLONE_NOFOLLOW); err == nil {
		return nil
	}

	return copyDirWalk(src, dst)
}

func copyDirWalk(src, dst string) error {
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
