package node

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShimDir(t *testing.T) {
	dir, err := ShimDir()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".wade", "shims")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestUseVersion_NotInstalled(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	err := UseVersion("v99.99.99")
	if err == nil {
		t.Error("expected error for uninstalled version")
	}
}

func TestUseVersion_CreatesShims(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create a fake version installation
	versionsDir := filepath.Join(tmpHome, ".wade", "versions", "v20.12.0", "bin")
	os.MkdirAll(versionsDir, 0755)

	// Create fake node/npm/npx binaries
	for _, name := range []string{"node", "npm", "npx"} {
		os.WriteFile(filepath.Join(versionsDir, name), []byte("fake"), 0755)
	}

	if err := UseVersion("v20.12.0"); err != nil {
		t.Fatal(err)
	}

	// Check shims were created
	shimDir := filepath.Join(tmpHome, ".wade", "shims")
	for _, name := range []string{"node", "npm", "npx"} {
		target := filepath.Join(shimDir, name)
		info, err := os.Lstat(target)
		if err != nil {
			t.Fatalf("shim %s not created: %v", name, err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("expected %s to be a symlink", target)
		}
		// Check the symlink target
		dest, _ := os.Readlink(target)
		expected := filepath.Join(versionsDir, name)
		if dest != expected {
			t.Errorf("shim %s → %s, want → %s", name, dest, expected)
		}
	}

	// Check current file was written
	currentFile := filepath.Join(tmpHome, ".wade", "current")
	data, _ := os.ReadFile(currentFile)
	if string(data) != "v20.12.0\n" {
		t.Errorf("expected current file to contain 'v20.12.0\\n', got %q", string(data))
	}
}

func TestUseVersion_OverwritesOldShims(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create two versions
	for _, ver := range []string{"v18.20.0", "v20.12.0"} {
		binDir := filepath.Join(tmpHome, ".wade", "versions", ver, "bin")
		os.MkdirAll(binDir, 0755)
		for _, name := range []string{"node", "npm", "npx"} {
			os.WriteFile(filepath.Join(binDir, name), []byte("fake"), 0755)
		}
	}

	// Switch to v18
	UseVersion("v18.20.0")

	// Switch to v20
	if err := UseVersion("v20.12.0"); err != nil {
		t.Fatal(err)
	}

	// Verify shim points to v20
	shimDir := filepath.Join(tmpHome, ".wade", "shims")
	dest, _ := os.Readlink(filepath.Join(shimDir, "node"))
	expected := filepath.Join(tmpHome, ".wade", "versions", "v20.12.0", "bin", "node")
	if dest != expected {
		t.Errorf("shim points to %s, want %s", dest, expected)
	}
}

func TestUseVersion_SetsDefaultIfEmpty(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	// Create a fake version
	binDir := filepath.Join(tmpHome, ".wade", "versions", "v20.12.0", "bin")
	os.MkdirAll(binDir, 0755)
	for _, name := range []string{"node", "npm", "npx"} {
		os.WriteFile(filepath.Join(binDir, name), []byte("fake"), 0755)
	}

	UseVersion("v20.12.0")

	// Check config was saved with default
	cfgFile := filepath.Join(tmpHome, ".wade", "config.toml")
	data, _ := os.ReadFile(cfgFile)
	if !strings.Contains(string(data), "default_version = 'v20.12.0'") {
		t.Errorf("expected default_version in config, got:\n%s", string(data))
	}
}
