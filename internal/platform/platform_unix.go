//go:build !windows

package platform

import "os"

// Symlink creates a symbolic link from newname to oldname.
func Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}
