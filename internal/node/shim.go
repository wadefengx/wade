package node

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wadefengx/wade/internal/config"
)

// ShimDir returns the path to ~/.wade/shims/
func ShimDir() (string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "shims"), nil
}

// UseVersion switches the active Node version by updating shim symlinks
func UseVersion(version string) error {
	if !IsInstalled(version) {
		return fmt.Errorf("version %s is not installed — run 'wade node install %s' first",
			version, version)
	}

	dir, err := config.WadeDir()
	if err != nil {
		return err
	}

	shimDir, err := ShimDir()
	if err != nil {
		return err
	}

	// Ensure shims directory exists
	if err := os.MkdirAll(shimDir, 0755); err != nil {
		return fmt.Errorf("create shims dir: %w", err)
	}

	versionBin := filepath.Join(dir, "versions", version, "bin")

	// Create shims for node, npm, npx
	shims := []struct{ name, target string }{
		{"node", filepath.Join(versionBin, "node")},
		{"npm", filepath.Join(versionBin, "npm")},
		{"npx", filepath.Join(versionBin, "npx")},
	}

	for _, s := range shims {
		target := filepath.Join(shimDir, s.name)

		// Remove existing symlink
		os.Remove(target)

		// Create new symlink
		if err := os.Symlink(s.target, target); err != nil {
			return fmt.Errorf("create shim for %s: %w", s.name, err)
		}
	}

	// Also try yarn and pnpm if they exist in this version
	for _, optional := range []string{"yarn", "pnpm"} {
		target := filepath.Join(versionBin, optional)
		if _, err := os.Stat(target); err == nil {
			shim := filepath.Join(shimDir, optional)
			os.Remove(shim)
			os.Symlink(target, shim)
		}
	}

	// Write current version
	currentFile := filepath.Join(dir, "current")
	if err := os.WriteFile(currentFile, []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("write current version: %w", err)
	}

	// Update default if not set
	cfg, _ := config.Load()
	if cfg.DefaultVersion == "" {
		cfg.DefaultVersion = version
		config.Save(cfg)
	}

	return nil
}

// CurrentVersion returns the currently active Node version
func CurrentVersion() (string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(filepath.Join(dir, "current"))
	if err != nil {
		return "", fmt.Errorf("no active version — run 'wade node use <version>' first")
	}

	return strings.TrimSpace(string(data)), nil
}
