package node

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/platform"
)

// ShimDir returns the path to ~/.wade/shims/
func ShimDir() (string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "shims"), nil
}

// shimTargets returns the shim name → target path pairs for a Node version.
// Windows: node.exe/npm.cmd/npx.cmd at version root. Unix: bin/node etc.
func shimTargets(versionRoot, versionBin string, isWindows bool) []struct{ name, target string } {
	if isWindows {
		return []struct{ name, target string }{
			{"node.exe", filepath.Join(versionRoot, "node.exe")},
			{"npm.cmd", filepath.Join(versionRoot, "npm.cmd")},
			{"npx.cmd", filepath.Join(versionRoot, "npx.cmd")},
		}
	}
	return []struct{ name, target string }{
		{"node", filepath.Join(versionBin, "node")},
		{"npm", filepath.Join(versionBin, "npm")},
		{"npx", filepath.Join(versionBin, "npx")},
	}
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

	// Windows layout: node.exe at version root, npm/npx as .cmd wrappers.
	// Unix layout: bin/node, bin/npm, bin/npx.
	versionBin := filepath.Join(dir, "versions", version, "bin")
	versionRoot := filepath.Join(dir, "versions", version)

	shims := shimTargets(versionRoot, versionBin, runtime.GOOS == "windows")

	for _, s := range shims {
		target := filepath.Join(shimDir, s.name)

		// Remove existing symlink
		if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove existing shim for %s: %w", s.name, err)
		}

		// Create new symlink
		if err := platform.Symlink(s.target, target); err != nil {
			return fmt.Errorf("create shim for %s: %w", s.name, err)
		}
	}

	// Also try yarn and pnpm if they exist in this version
	optionalNames := []string{"yarn", "pnpm"}
	if runtime.GOOS == "windows" {
		optionalNames = []string{"yarn.cmd", "pnpm.cmd"}
	}
	for _, optional := range optionalNames {
		target := filepath.Join(versionRoot, optional)
		if _, err := os.Stat(target); err != nil {
			target = filepath.Join(versionBin, optional)
		}
		if _, err := os.Stat(target); err == nil {
			shim := filepath.Join(shimDir, optional)
			if err := os.Remove(shim); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove existing shim for %s: %w", optional, err)
			}
			if err := platform.Symlink(target, shim); err != nil {
				return fmt.Errorf("create shim for %s: %w", optional, err)
			}
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
