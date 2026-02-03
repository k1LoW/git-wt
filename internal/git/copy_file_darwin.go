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

	// Skip directories
	if srcInfo.IsDir() {
		return nil
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Try clonefile first (APFS Copy-on-Write)
	if err := unix.Clonefile(src, dst, unix.CLONE_NOFOLLOW); err == nil {
		// clonefile preserves most permissions but strips setuid/setgid bits,
		// so chmod is needed to restore the original mode completely.
		return os.Chmod(dst, srcInfo.Mode())
	}

	// Fallback to traditional copy (non-APFS, cross-device, etc.)
	return copyFileTraditional(src, dst, srcInfo)
}

func copyFileTraditional(src, dst string, srcInfo os.FileInfo) error {
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
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Preserve file timestamps
	return os.Chtimes(dst, srcInfo.ModTime(), srcInfo.ModTime())
}
