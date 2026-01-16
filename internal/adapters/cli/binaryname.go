package cli

// binaryname.go provides runtime resolution of the actual binary name as invoked
// by the user, enabling CLI output that reflects the user's invocation context
// (e.g., symlinks, renamed binaries).

import (
	"os"
	"path/filepath"
)

// DefaultBinaryName is the fallback name when os.Args[0] is unavailable or invalid.
const DefaultBinaryName = "vdash"

// BinaryName returns the name of the binary as invoked by the user.
// Uses the basename of os.Args[0], falling back to DefaultBinaryName if unavailable.
func BinaryName() string {
	if len(os.Args) == 0 {
		return DefaultBinaryName
	}
	return binaryNameFrom(os.Args[0])
}

// binaryNameFrom extracts the binary name from the given arg0 string.
// This function is separated for testability without mocking os.Args.
func binaryNameFrom(arg0 string) string {
	if arg0 == "" {
		return DefaultBinaryName
	}

	base := filepath.Base(arg0)

	// Handle edge cases: "." (current dir) and "/" (root path on Unix)
	if base == "." || base == "/" {
		return DefaultBinaryName
	}

	return base
}
