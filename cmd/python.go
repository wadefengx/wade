package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/python"
)

var pythonCmd = &cobra.Command{
	Use:   "python",
	Short: "Manage Python versions and pip mirrors",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var pythonInstallCmd = &cobra.Command{
	Use:   "install <version>",
	Short: "Install a Python version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return python.Install(args[0])
	},
}

var pythonUseCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Switch to a Python version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := resolveInstalledPythonVersion(args[0])
		if err != nil {
			return err
		}
		return python.UseVersion(version)
	},
}

var pythonLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List managed and system Python versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		versions, err := python.InstalledVersions()
		if err != nil {
			return err
		}
		current, _ := python.CurrentVersion()
		cfg, _ := config.Load()

		fmt.Println("📦 Managed Python versions:")
		if len(versions) == 0 {
			fmt.Println("  (none — run 'wade python install 3.12')")
		}
		for _, version := range versions {
			markers := ""
			if version == current {
				markers += " (current)"
			}
			if version == cfg.DefaultPythonVersion {
				markers += " (default)"
			}
			fmt.Printf("  %s%s\n", version, markers)
		}

		fmt.Println("🐍 System Python:")
		pythons := python.DetectSystemPython()
		if len(pythons) == 0 {
			fmt.Println("  (no Python found on PATH)")
		}
		for _, detected := range pythons {
			fmt.Printf("  %s\n", detected)
		}
		return nil
	},
}

var pythonLsRemoteCmd = &cobra.Command{
	Use:   "ls-remote",
	Short: "List available Python versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		builds, err := python.FetchRemoteVersions()
		if err != nil {
			return err
		}
		installed, _ := python.InstalledVersions()
		installedSet := make(map[string]bool, len(installed))
		for _, version := range installed {
			installedSet[version] = true
		}
		limit := 20
		if len(builds) < limit {
			limit = len(builds)
		}
		fmt.Printf("Available Python versions (showing %d of %d):\n", limit, len(builds))
		for _, build := range builds[:limit] {
			marker := ""
			if installedSet[build.Version] {
				marker = " ✓"
			}
			fmt.Printf("  %s%s\n", build.Version, marker)
		}
		return nil
	},
}

var pythonDefaultCmd = &cobra.Command{
	Use:   "default <version>",
	Short: "Set the default Python version (and switch to it)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := resolveInstalledPythonVersion(args[0])
		if err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.DefaultPythonVersion = version
		if err := config.Save(cfg); err != nil {
			return err
		}
		if err := python.UseVersion(version); err != nil {
			return err
		}
		fmt.Printf("Default Python version set to %s\n", version)
		return nil
	},
}

var pythonUninstallCmd = &cobra.Command{
	Use:   "uninstall <version>",
	Short: "Remove a Python version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version, err := resolveInstalledPythonVersion(args[0])
		if err != nil {
			return err
		}
		if err := python.Uninstall(version); err != nil {
			return err
		}
		fmt.Printf("Uninstalled Python %s\n", version)
		return nil
	},
}

func resolveInstalledPythonVersion(input string) (string, error) {
	versions, err := python.InstalledVersions()
	if err != nil {
		return "", err
	}
	input = strings.TrimPrefix(input, "v")
	for _, version := range versions {
		if version == input {
			return version, nil
		}
	}
	for _, version := range versions {
		if strings.HasPrefix(version, input+".") {
			return version, nil
		}
	}
	return "", fmt.Errorf("Python %s is not installed — run 'wade python install %s'", input, input)
}

var pythonRegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage pip mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var pythonRegistryLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List pip mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {
		headers := []string{"Name", "URL"}
		var rows [][]string
		for _, mirror := range python.PipPresets() {
			rows = append(rows, []string{mirror.Name, mirror.URL})
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
	rootCmd.AddCommand(pythonCmd)
	pythonCmd.AddCommand(pythonInstallCmd, pythonUseCmd, pythonLsCmd, pythonLsRemoteCmd, pythonDefaultCmd, pythonUninstallCmd, pythonRegistryCmd)
	pythonRegistryCmd.AddCommand(pythonRegistryLsCmd, pythonRegistryUseCmd)
}
