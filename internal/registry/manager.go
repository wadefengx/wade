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
	var successCount int
	for _, pm := range []string{"npm", "yarn", "pnpm"} {
		if !commandExists(pm) {
			continue
		}
		cmd := exec.Command(pm, "config", "set", "registry", reg.URL)
		if out, err := cmd.CombinedOutput(); err != nil {
			// pnpm v11 requires Node >= 22, so it might fail on older Node.
			// Report the error but don't block the whole operation.
			errs = append(errs, fmt.Sprintf("  ⚠ %s: %s", pm, strings.TrimSpace(firstLine(string(out)))))
		} else {
			successCount++
		}
	}

	if successCount == 0 && len(errs) > 0 {
		return fmt.Errorf("failed to set registry for any package manager:\n%s", strings.Join(errs, "\n"))
	}

	// Save config even if some PMs failed (e.g., pnpm on older Node)
	cfg.CurrentRegistry = name
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Report warnings for failed PMs
	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Warnings:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
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

// firstLine returns the first line of a string, or the whole string if single line
func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
