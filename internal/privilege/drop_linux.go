//go:build linux

package privilege

import (
	"fmt"
	"os"
	"syscall"
)

// DropPrivileges drops process root privileges to target PUID and PGID if currently running as root.
func DropPrivileges(targetUID, targetGID int) error {
	currentUID := os.Getuid()
	if currentUID != 0 {
		// Not running as root, no privilege drop needed or possible.
		return nil
	}

	if targetUID == 0 && targetGID == 0 {
		// User explicitly requested to run as root.
		return nil
	}

	// Drop auxiliary groups first
	if err := syscall.Setgroups([]int{targetGID}); err != nil {
		return fmt.Errorf("failed to setgroups to GID %d: %w", targetGID, err)
	}

	// Drop GID before UID
	if err := syscall.Setgid(targetGID); err != nil {
		return fmt.Errorf("failed to setgid to GID %d: %w", targetGID, err)
	}

	// Drop UID
	if err := syscall.Setuid(targetUID); err != nil {
		return fmt.Errorf("failed to setuid to UID %d: %w", targetUID, err)
	}

	return nil
}
