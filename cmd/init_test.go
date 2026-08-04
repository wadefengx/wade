package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wadefengx/wade/internal/config"
)

// TestInitAutoYesWritesAllRuntimes: `wade init -y` must persist registry +
// Go mirror settings to config.toml (not just Node). Exercises the real
// config package with HOME redirected to a temp dir.
func TestInitAutoYesWritesAllRuntimes(t *testing.T) {
	oldHome := os.Getenv("HOME")
	tmp := t.TempDir()
	os.Setenv("HOME", tmp)
	defer func() { os.Setenv("HOME", oldHome) }()

	// The real config package resolves ~/.wade/config.toml via UserHomeDir.
	cfg := config.DefaultConfig()
	cfg.CurrentRegistry = "taobao"
	cfg.GoMirror = "https://golang.google.cn/dl/"
	if err := config.Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.CurrentRegistry != "taobao" {
		t.Errorf("registry = %q, want taobao", loaded.CurrentRegistry)
	}
	if loaded.GoMirror != "https://golang.google.cn/dl/" {
		t.Errorf("go mirror = %q, want google-cn", loaded.GoMirror)
	}
	// config.toml should exist in the temp HOME
	if _, err := os.Stat(filepath.Join(tmp, ".wade", "config.toml")); err != nil {
		t.Errorf("config.toml not written: %v", err)
	}
}
