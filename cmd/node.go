package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/node"
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

		fmt.Printf("Resolved %s → %s\n", raw, resolved)
		return node.Install(resolved, cfg.NodeMirror)
	},
}

var nodeUseCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Switch to a Node.js version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]

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

		if err := node.UseVersion(matched); err != nil {
			return err
		}

		fmt.Printf("Now using %s\n", matched)
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
			fmt.Println("No Node.js versions installed.")
			fmt.Println("Run 'wade node install <version>' to install one.")
			return nil
		}

		fmt.Println("Installed Node versions:")
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

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(nodeInstallCmd)
	nodeCmd.AddCommand(nodeUseCmd)
	nodeCmd.AddCommand(nodeLsCmd)
	nodeCmd.AddCommand(nodeLsRemoteCmd)
	nodeCmd.AddCommand(nodeDefaultCmd)
	nodeCmd.AddCommand(nodeUninstallCmd)
	nodeCmd.AddCommand(nodeCurrentCmd)
}
