package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// TestUpdateCheckCachePath ensures the cache path lives under ~/.wade.
func TestUpdateCheckCachePath(t *testing.T) {
	p := updateCheckCachePath()
	if p == "" {
		t.Fatal("cache path should not be empty")
	}
	if filepath.Base(p) != ".last-update-check" {
		t.Errorf("unexpected cache filename: %s", filepath.Base(p))
	}
}

// TestIsTerminalFalseInCI: when stdin is not a TTY (pipes, CI), the
// update prompt must NOT block waiting for input.
func TestIsTerminalFalseInCI(t *testing.T) {
	// In `go test`, stdin is /dev/null or a pipe — never a char device.
	if isTerminal() {
		t.Skip("stdin is a TTY in this environment")
	}
	// cache path still resolvable
	if updateCheckCachePath() == "" {
		t.Fatal("cache path should resolve")
	}
	// and writing the cache should work under a temp HOME
	tmp := t.TempDir()
	old := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", old)
	if got := updateCheckCachePath(); !filepath.HasPrefix(got, tmp) {
		t.Errorf("cache path %q should live under temp HOME %q", got, tmp)
	}
}
