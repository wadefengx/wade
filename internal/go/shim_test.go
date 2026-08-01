package golang

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/wadefengx/wade/internal/config"
)

// TestUseVersionWindowsShims: on Windows the shim is a go.cmd WRAPPER
// (not a hardlink!). Go 1.21+ binaries are trimmed and infer GOROOT from
// their own executable path — a hardlink shim breaks that ("go binary is
// trimmed and GOROOT is not set"). The .cmd wrapper calls the real binary
// path and sets GOROOT explicitly.
func TestUseVersionWindowsShims(t *testing.T) {
	// Redirect HOME so WadeDir() resolves into the temp dir, not ~/.wade
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", t.TempDir())
	defer func() { os.Setenv("HOME", oldHome) }()

	dir, _ := config.WadeDir()
	version := "go1.23.12"
	versionDir := filepath.Join(dir, "go", "versions", version)
	binDir := filepath.Join(versionDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Simulate layout: go.exe / gofmt.exe on Windows, go/gofmt on Unix
	if runtime.GOOS == "windows" {
		for _, f := range []string{"go.exe", "gofmt.exe"} {
			if err := os.WriteFile(filepath.Join(binDir, f), []byte("bin"), 0755); err != nil {
				t.Fatal(err)
			}
		}
	} else {
		for _, f := range []string{"go", "gofmt"} {
			if err := os.WriteFile(filepath.Join(binDir, f), []byte("bin"), 0755); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := UseVersion(version); err != nil {
		t.Fatalf("UseVersion: %v", err)
	}

	shimDir := filepath.Join(dir, "shims")
	// Windows: go.cmd wrapper (contains GOROOT + real path); Unix: symlink `go`.
	want := []string{"go", "gofmt"}
	if runtime.GOOS == "windows" {
		want = []string{"go.cmd", "gofmt.cmd"}
	}
	for _, w := range want {
		shim := filepath.Join(shimDir, w)
		if _, err := os.Stat(shim); err != nil {
			t.Errorf("shim %q not created: %v", w, err)
		}
		if runtime.GOOS == "windows" {
			data, _ := os.ReadFile(shim)
			if !strings.Contains(string(data), "GOROOT") || !strings.Contains(string(data), "go.exe") {
				t.Errorf("shim %q should be a GOROOT-setting wrapper, got:\n%s", w, data)
			}
		}
	}
}
