package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/node"
	"github.com/wadefengx/wade/internal/registry"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage Node.js versions",
	Long:  `Install, switch, list, and manage Node.js versions.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var nodeInstallCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a Node.js version",
	Long:  `Download and install a Node.js version. Supports partial versions like "18", "18.20".`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Resolve version
		resolved, err := node.ResolveVersion(raw, cfg.NodeMirror)
		if err != nil {
			return fmt.Errorf("resolve version: %w", err)
		}

		fmt.Printf("📥 Resolved %s → %s\n", raw, resolved)
		return node.Install(resolved, cfg.NodeMirror)
	},
}

var nodeUseCmd = &cobra.Command{
	Use:   "use [version]",
	Short: "Switch to a Node.js version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var version string
		usingProjectVersion := false
		if len(args) == 1 {
			version = args[0]
		} else {
			var err error
			version, err = node.FindProjectVersion()
			if err != nil {
				if errors.Is(err, node.ErrProjectVersionNotFound) {
					return cobra.ExactArgs(1)(cmd, args)
				}
				return err
			}
			usingProjectVersion = true
		}

		// Try to resolve partial version against installed versions
		installed, _ := node.InstalledVersions()
		matched := ""
		for _, v := range installed {
			if strings.HasPrefix(v, "v"+strings.TrimPrefix(version, "v")) {
				matched = v
				break
			}
		}
		if matched == "" {
			// Exact match
			matched = version
			if !strings.HasPrefix(matched, "v") {
				matched = "v" + matched
			}
		}

		if usingProjectVersion {
			fmt.Printf("Using project version %s (.wade-version)\n", matched)
		}
		if err := node.UseVersion(matched); err != nil {
			return err
		}

		fmt.Printf("🟢 Now using %s\n", matched)
		return nil
	},
}

var nodeLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List installed Node.js versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		current, _ := node.CurrentVersion()

		versions, err := node.InstalledVersions()
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			fmt.Println("📭 No Node.js versions installed.")
			fmt.Println("Run 'wade node install <version>' to install one.")
			return nil
		}

		fmt.Println("📦 Installed Node versions:")
		for _, v := range versions {
			markers := ""
			if v == current {
				markers += " (current)"
			}
			if cfg.DefaultVersion == v {
				markers += " (default)"
			}
			fmt.Printf("  %s%s\n", v, markers)
		}
		return nil
	},
}

var nodeLsRemoteCmd = &cobra.Command{
	Use:   "ls-remote",
	Short: "List available Node.js versions from mirror",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		versions, err := node.FetchRemoteVersions(cfg.NodeMirror)
		if err != nil {
			return fmt.Errorf("fetch versions: %w", err)
		}

		installed, _ := node.InstalledVersions()
		installedSet := make(map[string]bool)
		for _, v := range installed {
			installedSet[v] = true
		}

		// Show last 20 versions
		limit := 20
		if len(versions) < limit {
			limit = len(versions)
		}

		fmt.Printf("Available versions (showing %d of %d):\n", limit, len(versions))
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

var nodeDefaultCmd = &cobra.Command{
	Use:   "default <version>",
	Short: "Set the default Node.js version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}

		if !node.IsInstalled(version) {
			return fmt.Errorf("version %s is not installed", version)
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.DefaultVersion = version
		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Default Node version set to %s\n", version)
		return nil
	},
}

var nodeUninstallCmd = &cobra.Command{
	Use:   "uninstall <version>",
	Short: "Remove a Node.js version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]
		if !strings.HasPrefix(version, "v") {
			version = "v" + version
		}

		if err := node.Uninstall(version); err != nil {
			return err
		}

		fmt.Printf("Uninstalled %s\n", version)
		return nil
	},
}

var nodeCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Print the currently active Node.js version",
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := node.CurrentVersion()
		if err != nil {
			fmt.Println("no active version")
			return nil
		}
		fmt.Println(version)
		return nil
	},
}

var nodeMirrorCmd = &cobra.Command{
	Use:   "mirror",
	Short: "Show or set Node.js download mirror",
	Long:  `Manage where wade downloads Node.js binaries from.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return mirrorUse(args[0])
		}
		cfg, _ := config.Load()
		label := findMirrorLabel(cfg.NodeMirror)
		fmt.Printf("🌐 Node download source: %s\n   %s\n", label, cfg.NodeMirror)
		return nil
	},
}

var nodeMirrorLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List available Node mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		headers := []string{"Name", "URL", "Status"}
		var rows [][]string
		for _, m := range builtinMirrors {
			status := ""
			if strings.HasPrefix(cfg.NodeMirror, m.URL) || cfg.NodeMirror == m.URL {
				status = "current"
			}
			if m.Name == "official" {
				status += " global"
			} else if status == "" {
				status = "cn"
			}
			rows = append(rows, []string{m.Name, m.URL, strings.TrimSpace(status)})
		}
		renderTable(headers, rows)
		return nil
	},
}

var nodeMirrorUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch Node download mirror",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return mirrorUse(args[0])
	},
}

var nodeMirrorTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test latency of Node mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("⚡ Testing Node mirror latency...")
		// Only test node mirrors, not npm registries
		mirrorRegs := make([]registry.Registry, len(builtinMirrors))
		for i, m := range builtinMirrors {
			mirrorRegs[i] = registry.Registry{Name: m.Name, URL: m.URL, IsBuiltIn: true}
		}
		results := registry.TestMirrors(mirrorRegs)
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

type mirror struct {
	Name string
	URL  string
}

var builtinMirrors = []mirror{
	{Name: "official", URL: "https://nodejs.org/dist/"},
	{Name: "npmmirror", URL: "https://npmmirror.com/mirrors/node/"},
	{Name: "tsinghua", URL: "https://mirrors.tuna.tsinghua.edu.cn/nodejs-release/"},
	{Name: "ustc", URL: "https://mirrors.ustc.edu.cn/node/"},
	{Name: "huawei", URL: "https://mirrors.huaweicloud.com/nodejs/"},
	{Name: "aliyun", URL: "https://mirrors.aliyun.com/nodejs-release/"},
	{Name: "tencent", URL: "https://mirrors.tencent.com/nodejs-release/"},
}

func findMirrorLabel(url string) string {
	for _, m := range builtinMirrors {
		if strings.HasPrefix(url, m.URL) || url == m.URL {
			return m.Name
		}
	}
	return "custom"
}

func mirrorUse(name string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var found *mirror
	for _, m := range builtinMirrors {
		if m.Name == name {
			found = &m
			break
		}
	}
	if found == nil {
		return fmt.Errorf("unknown mirror: %s — use 'wade node mirror ls' to see available mirrors", name)
	}

	cfg.NodeMirror = found.URL
	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Printf("🌐 Switched Node mirror to %s (%s)\n", found.Name, found.URL)
	return nil
}

// toRegistryMirrors converts built-in mirrors to registry.Registry for the test function
func toRegistryMirrors() []registry.Registry {
	regs := make([]registry.Registry, len(builtinMirrors))
	for i, m := range builtinMirrors {
		regs[i] = registry.Registry{Name: m.Name, URL: m.URL, IsBuiltIn: true}
	}
	return regs
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeInstallCmd)
	nodeCmd.AddCommand(nodeUseCmd)
	nodeCmd.AddCommand(nodeLsCmd)
	nodeCmd.AddCommand(nodeLsRemoteCmd)
	nodeCmd.AddCommand(nodeDefaultCmd)
	nodeCmd.AddCommand(nodeUninstallCmd)
	nodeCmd.AddCommand(nodeCurrentCmd)
	nodeCmd.AddCommand(nodeMirrorCmd)
	nodeMirrorCmd.AddCommand(nodeMirrorLsCmd)
	nodeMirrorCmd.AddCommand(nodeMirrorUseCmd)
	nodeMirrorCmd.AddCommand(nodeMirrorTestCmd)
}
