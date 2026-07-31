package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	golang "github.com/wadefengx/wade/internal/go"
	"github.com/wadefengx/wade/internal/python"
	"github.com/wadefengx/wade/internal/registry"
)

// ── Go commands ──

var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Manage Go versions, mirrors, and proxies",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var goInstallCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a Go version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver := args[0]
		if !strings.HasPrefix(ver, "go") {
			ver = "go" + ver
		}
		cfg, _ := config.Load()
		mirror := cfg.GoMirror
		if mirror == "" {
			mirror = "https://go.dev/dl/"
		}
		return golang.Install(ver, mirror)
	},
}

var goUseCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Switch to a Go version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver := args[0]
		if !strings.HasPrefix(ver, "go") {
			ver = "go" + ver
		}
		return golang.UseVersion(ver)
	},
}

var goLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List installed Go versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		versions, _ := golang.InstalledVersions()
		current, _ := golang.CurrentVersion()
		cfg, _ := config.Load()

		fmt.Println("📦 Installed Go versions:")
		for _, v := range versions {
			markers := ""
			if v == current {
				markers += " (current)"
			}
			if cfg.DefaultGoVersion == v {
				markers += " (default)"
			}
			fmt.Printf("  %s%s\n", v, markers)
		}
		if len(versions) == 0 {
			fmt.Println("  (none — run 'wade go install 1.23')")
		}
		return nil
	},
}

var goLsRemoteCmd = &cobra.Command{
	Use:   "ls-remote",
	Short: "List available Go versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		mirror := cfg.GoMirror
		if mirror == "" {
			mirror = "https://go.dev/dl/"
		}
		versions, err := golang.FetchRemoteVersions(mirror)
		if err != nil {
			return err
		}
		installed, _ := golang.InstalledVersions()
		installedSet := make(map[string]bool)
		for _, v := range installed {
			installedSet[v] = true
		}

		limit := 20
		if len(versions) < limit {
			limit = len(versions)
		}
		fmt.Printf("Available Go versions (showing %d of %d):\n", limit, len(versions))
		for i := 0; i < limit; i++ {
			v := versions[i]
			marker := ""
			if installedSet[v] {
				marker = " ✓"
			}
			fmt.Printf("  %s%s\n", v, marker)
		}
		return nil
	},
}

var goMirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Show or set Go download mirror",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		cur := cfg.GoMirror
		if cur == "" {
			cur = "https://go.dev/dl/ (default)"
		}
		fmt.Printf("🌐 Go download source: %s\n", cur)
		return nil
	},
}

var goMirrorLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Go download mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		headers := []string{"Name", "URL", "Status"}
		var rows [][]string
		for _, m := range python.GoMirrorPresets() {
			status := ""
			if strings.HasPrefix(cfg.GoMirror, m.URL) {
				status = "current"
			}
			rows = append(rows, []string{m.Name, m.URL, status})
		}
		renderTable(headers, rows)
		return nil
	},
}

var goMirrorUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch Go download mirror",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m, ok := python.FindGoMirror(args[0])
		if !ok {
			return fmt.Errorf("unknown mirror: %s — use 'wade go mirror ls' to see available", args[0])
		}
		cfg, _ := config.Load()
		cfg.GoMirror = m.URL
		config.Save(cfg)
		fmt.Printf("🌐 Switched Go mirror to %s (%s)\n", m.Name, m.URL)
		return nil
	},
}

var goMirrorTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test Go mirror latency",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("⚡ Testing Go mirror latency...")
		mirrors := python.GoMirrorPresets()
		regs := make([]registry.Registry, len(mirrors))
		for i, m := range mirrors {
			regs[i] = registry.Registry{Name: m.Name, URL: m.URL, IsBuiltIn: true}
		}
		results := registry.TestMirrors(regs)
		headers := []string{"Mirror", "URL", "Latency"}
		var rows [][]string
		for _, r := range results {
			lat := r.Latency.String()
			if r.Error != "" {
				lat = r.Error
			}
			rows = append(rows, []string{r.Name, r.URL, lat})
		}
		renderTable(headers, rows)
		return nil
	},
}

var goProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "Manage Go module proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		return nil
	},
}

var goProxyLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List Go proxies",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, p := range python.GoProxyPresets() {
			fmt.Printf("  %s  %s\n", p.Name, p.URL)
		}
		return nil
	},
}

var goProxyUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch Go proxy",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return python.UseGoProxy(args[0])
	},
}

// ── Python commands ──

var pythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Manage Python versions and pip mirrors",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var pythonLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "Show Python versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		pythons := python.DetectSystemPython()
		fmt.Println("🐍 Detected Python:")
		for _, p := range pythons {
			fmt.Printf("  %s\n", p)
		}
		if len(pythons) == 0 {
			fmt.Println("  (no Python found on PATH)")
		}
		return nil
	},
}

var pythonRegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage pip mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		return nil
	},
}

var pythonRegistryLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List pip mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		headers := []string{"Name", "URL"}
		var rows [][]string
		for _, m := range python.PipPresets() {
			rows = append(rows, []string{m.Name, m.URL})
		}
		renderTable(headers, rows)
		return nil
	},
}

var pythonRegistryUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch pip to a mirror",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return python.UsePipMirror(args[0])
	},
}

func init() {
	// Go commands
	rootCmd.AddCommand(goCmd)
	goCmd.AddCommand(goInstallCmd)
	goCmd.AddCommand(goUseCmd)
	goCmd.AddCommand(goLsCmd)
	goCmd.AddCommand(goLsRemoteCmd)
	goCmd.AddCommand(goMirrorCmd)
	goCmd.AddCommand(goProxyCmd)
	goMirrorCmd.AddCommand(goMirrorLsCmd)
	goMirrorCmd.AddCommand(goMirrorUseCmd)
	goMirrorCmd.AddCommand(goMirrorTestCmd)
	goProxyCmd.AddCommand(goProxyLsCmd)
	goProxyCmd.AddCommand(goProxyUseCmd)

	// Python commands
	rootCmd.AddCommand(pythonCmd)
	pythonCmd.AddCommand(pythonLsCmd)
	pythonCmd.AddCommand(pythonRegistryCmd)
	pythonRegistryCmd.AddCommand(pythonRegistryLsCmd)
	pythonRegistryCmd.AddCommand(pythonRegistryUseCmd)
}
