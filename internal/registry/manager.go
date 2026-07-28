package registry

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/wadefengx/wade/internal/config"
)

// Use switches all package managers to the given registry
func Use(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	reg, ok := Find(name, toConfigRegistries(cfg.Registries))
	if !ok {
		return fmt.Errorf("registry %q not found — use 'wade registry ls' to see available registries", name)
	}

	var errs []string
	for _, pm := range []string{"npm", "yarn", "pnpm"} {
		if !commandExists(pm) {
			continue
		}
		cmd := exec.Command(pm, "config", "set", "registry", reg.URL)
		if out, err := cmd.CombinedOutput(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s (%v)", pm, strings.TrimSpace(string(out)), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to set registry for some package managers:\n%s", strings.Join(errs, "\n"))
	}

	cfg.CurrentRegistry = name
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	return nil
}

// Add adds a custom registry to the config
func Add(name, url string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Validate URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("registry URL must start with http:// or https://")
	}

	// Check duplicate
	for _, r := range cfg.Registries {
		if r.Name == name {
			return fmt.Errorf("registry %q already exists", name)
		}
	}
	// Also check built-in
	if IsBuiltIn(name) {
		return fmt.Errorf("registry %q is a built-in preset and cannot be overridden", name)
	}

	cfg.Registries = append(cfg.Registries, config.Registry{Name: name, URL: url})
	return config.Save(cfg)
}

// Remove deletes a custom registry from the config
func Remove(name string) error {
	if IsBuiltIn(name) {
		return fmt.Errorf("cannot delete built-in registry %q", name)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	found := false
	for i, r := range cfg.Registries {
		if r.Name == name {
			cfg.Registries = append(cfg.Registries[:i], cfg.Registries[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("registry %q not found", name)
	}

	// If deleted registry was current, reset to npm
	if cfg.CurrentRegistry == name {
		cfg.CurrentRegistry = "npm"
	}

	return config.Save(cfg)
}

// toConfigRegistries converts config.Registry slice to internal Registry slice
func toConfigRegistries(cfgRegs []config.Registry) []Registry {
	regs := make([]Registry, len(cfgRegs))
	for i, r := range cfgRegs {
		regs[i] = Registry{
			Name:      r.Name,
			URL:       r.URL,
			IsBuiltIn: false,
		}
	}
	return regs
}

// commandExists checks if a command is available on PATH
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetCurrent returns the currently active registry name and URL
func GetCurrent() (string, string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", "", err
	}

	reg, ok := Find(cfg.CurrentRegistry, toConfigRegistries(cfg.Registries))
	if !ok {
		return cfg.CurrentRegistry, "unknown", nil
	}
	return reg.Name, reg.URL, nil
}

// EnsureWadeDir creates ~/.wade/ directory structure
func EnsureWadeDir() error {
	dir, err := config.WadeDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, 0755)
}
