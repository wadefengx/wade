package golang

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wadefengx/wade/internal/config"
)

// TestUseVersionWindowsShims: on Windows the shim must resolve go.exe —
// UseVersion looks up go.exe/gofmt.exe and creates extensionless shims
// (`go` → shims/go pointing at go.exe) so cmd's PATHEXT finds them.
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
	for _, want := range []string{"go", "gofmt"} {
		if _, err := os.Stat(filepath.Join(shimDir, want)); err != nil {
			t.Errorf("shim %q not created: %v", want, err)
		}
	}
}
