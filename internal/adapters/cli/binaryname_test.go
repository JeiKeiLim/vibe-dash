package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBinaryNameFrom(t *testing.T) {
	tests := []struct {
		name        string
		arg0        string
		expected    string
		windowsOnly bool
	}{
		{"normal invocation", "vdash", "vdash", false},
		{"full path unix", "/usr/local/bin/vdash", "vdash", false},
		{"full path windows", "C:\\Program Files\\vdash.exe", "vdash.exe", true},
		{"full path with spaces", "/path/to/my tool", "my tool", false},
		{"renamed binary", "my-custom-name", "my-custom-name", false},
		{"symlink name", "v", "v", false},
		{"empty string", "", "vdash", false},
		{"dot only", ".", "vdash", false},
		{"root path unix", "/", "vdash", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip Windows-specific tests on non-Windows platforms
			// filepath.Base behaves differently per platform for backslash paths
			if tt.windowsOnly && runtime.GOOS != "windows" {
				t.Skip("Windows-specific test, skipping on non-Windows platform")
			}
			got := binaryNameFrom(tt.arg0)
			if got != tt.expected {
				t.Errorf("binaryNameFrom(%q) = %q, want %q", tt.arg0, got, tt.expected)
			}
		})
	}
}

func TestBinaryName(t *testing.T) {
	// BinaryName() uses os.Args[0], which in test context is the test binary path.
	// We verify it returns the base name of whatever invoked this test.
	got := BinaryName()
	if got == "" {
		t.Error("BinaryName() returned empty string")
	}

	// The result should match filepath.Base of os.Args[0]
	// This validates BinaryName() actually uses os.Args[0] correctly
	expected := filepath.Base(os.Args[0])
	if got != expected {
		t.Errorf("BinaryName() = %q, want %q (filepath.Base of os.Args[0])", got, expected)
	}
}

func TestBinaryName_EmptyArgsDocumentation(t *testing.T) {
	// Note: The len(os.Args) == 0 branch in BinaryName() cannot be tested without
	// modifying the global os.Args slice, which would affect other tests and is
	// considered unsafe in concurrent test environments.
	//
	// The binaryNameFrom() function is tested separately to cover the actual
	// name resolution logic, while BinaryName() is an integration point that
	// depends on runtime state (os.Args).
	//
	// This test documents that the edge case exists and why it's not directly tested.
	t.Log("Empty os.Args handling is documented but not directly testable without global state modification")
}
