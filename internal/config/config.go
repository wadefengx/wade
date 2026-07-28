package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Registry represents a npm/yarn/pnpm registry
type Registry struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

// Config holds all wade configuration
type Config struct {
	DefaultVersion   string     `toml:"default_version"`
	NodeMirror       string     `toml:"node_mirror"`
	CurrentRegistry  string     `toml:"current_registry"`
	Registries       []Registry `toml:"registries"`
	GoMirror         string     `toml:"go_mirror"`
	DefaultGoVersion string     `toml:"default_go_version"`
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		DefaultVersion:  "",
		NodeMirror:      "https://npmmirror.com/mirrors/node/",
		CurrentRegistry: "npm",
		Registries:      nil,
	}
}

// WadeDir returns ~/.wade/
func WadeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".wade"), nil
}

// ConfigPath returns ~/.wade/config.toml
func ConfigPath() (string, error) {
	dir, err := WadeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load reads config from ~/.wade/config.toml, returns defaults if not found
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Fill defaults for zero-value fields
	if cfg.NodeMirror == "" {
		cfg.NodeMirror = "https://npmmirror.com/mirrors/node/"
	}
	if cfg.CurrentRegistry == "" {
		cfg.CurrentRegistry = "npm"
	}

	return &cfg, nil
}

// Save writes config to ~/.wade/config.toml, creating dirs as needed
func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
