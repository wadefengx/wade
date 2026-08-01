package cmd

import (
	"os"
	"runtime"
	"testing"
)

func TestPathInEnvPath(t *testing.T) {
	old := os.Getenv("PATH")

	// Build a PATH using platform-native separators and paths.
	// On Windows these are C:\... entries joined by ';'; on Unix
	// /usr/local/... joined by ':'. The function splits by
	// os.PathListSeparator, so test values must match that platform.
	var entries []string
	if runtime.GOOS == "windows" {
		entries = []string{`C:\Program Files\nodejs`, `C:\Users\wade\.wade\shims`, `C:\Program Files\Git\bin`}
	} else {
		entries = []string{"/usr/local/node", "/Users/wade/.wade/shims", "/usr/bin"}
	}
	os.Setenv("PATH", joinPathList(entries))
	defer os.Setenv("PATH", old)

	for _, dir := range entries {
		if !pathInEnvPath(dir) {
			t.Errorf("pathInEnvPath(%q) = false, want true", dir)
		}
		if runtime.GOOS == "windows" {
			// case-insensitive match on Windows paths
			if !pathInEnvPath(`c:\users\wade\.wade\shims`) {
				t.Error("expected case-insensitive match")
			}
		}
	}
	if pathInEnvPath(entries[0] + "-extra") {
		t.Error("prefix match should not count")
	}
}

// TestNormalizePath: old wade versions wrote doubled backslashes into the
// registry (C:\\Users\\...). normalizePath collapses them so pathInEnvPath /
// userPathHasShims match correctly.
func TestNormalizePath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only path normalization")
	}
	if got := normalizePath(`C:\\Users\\wade\\.wade\\shims`); got != `C:\Users\wade\.wade\shims` {
		t.Errorf("normalizePath = %q, want single backslashes", got)
	}
}

func joinPathList(entries []string) string {
	out := ""
	for i, e := range entries {
		if i > 0 {
			out += string(os.PathListSeparator)
		}
		out += e
	}
	return out
}

// TestPathInEnvPathRuntime is a compile guard: node.go uses runtime.GOOS,
// which must build on windows too.
func TestPathInEnvPathRuntime(t *testing.T) {
	_ = runtime.GOOS
}
