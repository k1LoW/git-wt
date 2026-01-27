//go:build !darwin

// Default implementation for non-Darwin platforms.
// On Linux (kernel 4.5+), io.Copy internally attempts copy_file_range(2),
// which enables efficient in-kernel copying on supported filesystems
// (Btrfs, XFS, ext4, NFS 4.2+, etc.) without data passing through userspace.

package git

import (
	"io"
	"os"
	"path/filepath"
)

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
