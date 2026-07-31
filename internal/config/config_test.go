package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.NodeMirror != "https://npmmirror.com/mirrors/node/" {
		t.Errorf("unexpected NodeMirror: %s", cfg.NodeMirror)
	}
	if cfg.CurrentRegistry != "npm" {
		t.Errorf("unexpected CurrentRegistry: %s", cfg.CurrentRegistry)
	}
	if cfg.DefaultVersion != "" {
		t.Errorf("expected empty DefaultVersion, got %s", cfg.DefaultVersion)
	}
}

func TestWadeDir(t *testing.T) {
	dir, err := WadeDir()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".wade")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".wade", "config.toml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Use a temp dir so we don't clobber the real config
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	cfg := &Config{
		DefaultVersion:   "v20.12.0",
		NodeMirror:       "https://npmmirror.com/mirrors/node/",
		CurrentRegistry:  "taobao",
		DefaultGoVersion: "go1.23.4",
		GoMirror:         "https://goproxy.cn",
	}

	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if loaded.DefaultVersion != "v20.12.0" {
		t.Errorf("expected v20.12.0, got %s", loaded.DefaultVersion)
	}
	if loaded.CurrentRegistry != "taobao" {
		t.Errorf("expected taobao, got %s", loaded.CurrentRegistry)
	}
	if loaded.DefaultGoVersion != "go1.23.4" {
		t.Errorf("expected go1.23.4, got %s", loaded.DefaultGoVersion)
	}
	if loaded.GoMirror != "https://goproxy.cn" {
		t.Errorf("expected goproxy.cn, got %s", loaded.GoMirror)
	}
}

func TestLoadNonExistentReturnsDefaults(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.NodeMirror != "https://npmmirror.com/mirrors/node/" {
		t.Errorf("expected default NodeMirror, got %s", cfg.NodeMirror)
	}
	if cfg.CurrentRegistry != "npm" {
		t.Errorf("expected default registry 'npm', got %s", cfg.CurrentRegistry)
	}
}

func TestLoadWithRegistries(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	cfg := &Config{
		Registries: []Registry{
			{Name: "myreg", URL: "https://myreg.com/"},
		},
	}
	Save(cfg)

	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Registries) != 1 {
		t.Fatalf("expected 1 registry, got %d", len(loaded.Registries))
	}
	if loaded.Registries[0].Name != "myreg" {
		t.Errorf("expected name 'myreg', got %s", loaded.Registries[0].Name)
	}
	if loaded.Registries[0].URL != "https://myreg.com/" {
		t.Errorf("expected URL 'https://myreg.com/', got %s", loaded.Registries[0].URL)
	}
}
