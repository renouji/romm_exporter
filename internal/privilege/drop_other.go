//go:build !linux

package privilege

import (
	"log"
	"os"
)

// DropPrivileges is a stub for non-Linux OSes where syscall.Setuid is unavailable or different.
func DropPrivileges(targetUID, targetGID int) error {
	if os.Getuid() == 0 && (targetUID != 0 || targetGID != 0) {
		log.Printf("[WARN] PUID/PGID privilege drop is only supported on Linux via syscall. Running without dropping privileges.")
	}
	return nil
}
