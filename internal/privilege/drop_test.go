package privilege

import (
	"os"
	"testing"
)

func TestDropPrivilegesNonRoot(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping non-root test when running as root")
	}

	err := DropPrivileges(1000, 1000)
	if err != nil {
		t.Fatalf("expected no error when running as non-root user, got: %v", err)
	}
}
