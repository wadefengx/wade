package golang

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wadefengx/wade/internal/config"
)

// TestUseVersionWindowsShims: on Windows the shim files keep the .exe
// extension (shims/go.exe → versions/.../bin/go.exe) — cmd/PowerShell's
// PATHEXT only matches .exe/.cmd/.bat, so extensionless shims are never
// found. (node shims are node.exe for the same reason.)
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

	// Simulate Windows layout: go.exe / gofmt.exe
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
	// On Windows the shim must be go.exe (PATHEXT); on Unix it's `go`.
	want := []string{"go", "gofmt"}
	if runtime.GOOS == "windows" {
		want = []string{"go.exe", "gofmt.exe"}
	}
	for _, w := range want {
		if _, err := os.Stat(filepath.Join(shimDir, w)); err != nil {
			t.Errorf("shim %q not created: %v", w, err)
		}
	}
}
