package node

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
		err   bool
	}{
		{"18", "18", false},
		{"18.20", "18.20", false},
		{"18.20.0", "18.20.0", false},
		{"v18.20.0", "18.20.0", false},
		{"lts", "lts", false},
		{"latest", "latest", false},
		{"", "", true},
		{"abc", "", true},
		{"18.20.0.1", "", true},
		{"18.x", "", true},
	}

	for _, tt := range tests {
		got, err := ParseVersion(tt.input)
		if tt.err {
			if err == nil {
				t.Errorf("ParseVersion(%q) expected error, got %q", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseVersion(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseVersion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMustAtoi(t *testing.T) {
	if n := mustAtoi("42"); n != 42 {
		t.Errorf("expected 42, got %d", n)
	}
	if n := mustAtoi("0"); n != 0 {
		t.Errorf("expected 0, got %d", n)
	}
}

func TestInstalledVersions_EmptyDir(t *testing.T) {
	// Use a temp home dir so there's no versions dir
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	versions, err := InstalledVersions()
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 0 {
		t.Errorf("expected 0 versions, got %d", len(versions))
	}
}

func TestInstalledVersions_WithDirs(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create version directories
	versionsDir := filepath.Join(tmpHome, ".wade", "versions")
	os.MkdirAll(versionsDir, 0755)
	os.MkdirAll(filepath.Join(versionsDir, "v20.12.0"), 0755)
	os.MkdirAll(filepath.Join(versionsDir, "v18.20.0"), 0755)
	os.MkdirAll(filepath.Join(versionsDir, "v22.4.0"), 0755)

	versions, err := InstalledVersions()
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d: %v", len(versions), versions)
	}
	// Should be sorted descending
	if versions[0] != "v22.4.0" {
		t.Errorf("expected first to be v22.4.0, got %s", versions[0])
	}
	if versions[2] != "v18.20.0" {
		t.Errorf("expected last to be v18.20.0, got %s", versions[2])
	}
}

func TestIsInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	versionsDir := filepath.Join(tmpHome, ".wade", "versions")
	os.MkdirAll(versionsDir, 0755)
	os.MkdirAll(filepath.Join(versionsDir, "v20.12.0"), 0755)

	if !IsInstalled("v20.12.0") {
		t.Error("expected v20.12.0 to be installed")
	}
	if IsInstalled("v18.20.0") {
		t.Error("expected v18.20.0 to NOT be installed")
	}
}

func TestCurrentVersion_NotFound(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	_, err := CurrentVersion()
	if err == nil {
		t.Error("expected error when no current version file exists")
	}
}

func TestCurrentVersion_Found(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Write current file
	dir := filepath.Join(tmpHome, ".wade")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "current"), []byte("v20.12.0\n"), 0644)

	ver, err := CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if ver != "v20.12.0" {
		t.Errorf("expected v20.12.0, got %s", ver)
	}
}
