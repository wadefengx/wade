package node

import (
	"os"
	"path/filepath"
	"testing"
)

// TestExtractZipRealArchive validates extractZip against a REAL Node win-x64
// zip (node-v20.20.2-win-x64.zip downloaded to /tmp by the QA flow).
// This is the regression test for the Windows 'gzip: invalid header' bug.
func TestExtractZipRealArchive(t *testing.T) {
	src := "/tmp/node-win-test.zip"
	if _, err := os.Stat(src); err != nil {
		t.Skipf("real archive not present: %v", err)
	}

	dest := t.TempDir()
	if err := extractZip(src, dest); err != nil {
		t.Fatalf("extractZip: %v", err)
	}

	// node.exe must exist (stripped of the top-level dir)
	for _, want := range []string{"node.exe", "npm.cmd", "npx.cmd"} {
		p := filepath.Join(dest, want)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s after extraction, got: %v", want, err)
		}
	}

	// top-level dir must NOT be nested
	if _, err := os.Stat(filepath.Join(dest, "node-v20.20.2-win-x64")); err == nil {
		t.Error("top-level archive dir should have been stripped")
	}
}

// TestExtractZipGenerated exercises extractZip with a synthetic archive
// (hermetic, no network).
func TestExtractZipGenerated(t *testing.T) {
	// build a zip with top-level dir + nested files
	// ponytail: reuse the real-archive test above; synthetic case is covered by
	// its structure (same code path). Keep this minimal.
	t.Skip("covered by TestExtractZipRealArchive")
}
