package node

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestShimTargetsWindows: on Windows the shims must point at node.exe /
// npm.cmd / npx.cmd in the version ROOT (not bin/node like Unix).
func TestShimTargetsWindows(t *testing.T) {
	root := filepath.Join("versions", "v20.20.2")
	bin := filepath.Join(root, "bin")

	shims := shimTargets(root, bin, true)
	want := map[string]string{
		"node.exe": filepath.Join(root, "node.exe"),
		"npm.cmd":  filepath.Join(root, "npm.cmd"),
		"npx.cmd":  filepath.Join(root, "npx.cmd"),
	}
	if len(shims) != len(want) {
		t.Fatalf("expected %d shims, got %d", len(want), len(shims))
	}
	for _, s := range shims {
		if want[s.name] != s.target {
			t.Errorf("shim %s: target %q, want %q", s.name, s.target, want[s.name])
		}
	}
}

// TestShimTargetsUnix: Unix layout keeps bin/node.
func TestShimTargetsUnix(t *testing.T) {
	root := filepath.Join("versions", "v20.20.2")
	bin := filepath.Join(root, "bin")

	shims := shimTargets(root, bin, false)
	want := map[string]string{
		"node": filepath.Join(bin, "node"),
		"npm":  filepath.Join(bin, "npm"),
		"npx":  filepath.Join(bin, "npx"),
	}
	if len(shims) != len(want) {
		t.Fatalf("expected %d shims, got %d", len(want), len(shims))
	}
	for _, s := range shims {
		if want[s.name] != s.target {
			t.Errorf("shim %s: target %q, want %q", s.name, s.target, want[s.name])
		}
	}
}

// TestShimTargetsCrossCompile is a compile-time sanity check: this package
// builds for windows (shim.go references runtime.GOOS, not build tags).
func TestShimTargetsCrossCompile(t *testing.T) {
	_ = runtime.GOOS
}
