//go:build windows

package platform

import (
	"fmt"
	"os"
)

// Symlink creates a symbolic link, falling back to a hard link when Windows
// does not permit unprivileged symbolic-link creation.
func Symlink(oldname, newname string) error {
	if err := os.Symlink(oldname, newname); err == nil {
		return nil
	} else if linkErr := os.Link(oldname, newname); linkErr != nil {
		return fmt.Errorf("create symbolic link (%v), then hard link: %w", err, linkErr)
	}
	return nil
}
