//go:build darwin

package guidedmigration

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// IsLocalPath reports whether path is backed by a filesystem macOS identifies
// as local. When the destination has not been created yet, its nearest existing
// ancestor determines the filesystem.
func IsLocalPath(path string) (bool, error) {
	existing := path
	for {
		_, err := os.Stat(existing)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return false, err
		}
		parent := filepath.Dir(existing)
		if parent == existing {
			return false, err
		}
		existing = parent
	}

	var stat unix.Statfs_t
	if err := unix.Statfs(existing, &stat); err != nil {
		return false, err
	}
	return stat.Flags&unix.MNT_LOCAL != 0, nil
}
