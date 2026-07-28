package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/registry"
)

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Manage npm/yarn/pnpm registries",
	Long:  `Switch, add, delete, list, and test npm/yarn/pnpm registries.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var registryLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all registries",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		all := registry.All(toRegistries(cfg.Registries))
		headers := []string{"Registry", "URL", "Status"}
		var rows [][]string
		for _, r := range all {
			status := ""
			if r.Name == cfg.CurrentRegistry {
				status = "current"
			}
			if !r.IsBuiltIn {
				if status != "" {
					status += ", custom"
				} else {
					status = "custom"
				}
			}
			rows = append(rows, []string{r.Name, r.URL, status})
		}
		renderTable(headers, rows)
		return nil
	},
}

var registryUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch all package managers to a registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := registry.Use(name); err != nil {
			return err
		}
		cfg, _ := config.Load()
		r, _ := registry.Find(cfg.CurrentRegistry, toRegistries(cfg.Registries))
		if r != nil {
			fmt.Printf("Switched to %s (%s)\n", r.Name, r.URL)
		}
		return nil
	},
}

var registryAddCmd = &cobra.Command{
	Use:   "add <name> <url>",
	Short: "Add a custom registry",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, url := args[0], args[1]
		if err := registry.Add(name, url); err != nil {
			return err
		}
		fmt.Printf("Added custom registry %q → %s\n", name, url)
		return nil
	},
}

var registryDelCmd = &cobra.Command{
	Use:   "del <name>",
	Short: "Delete a custom registry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := registry.Remove(name); err != nil {
			return err
		}
		fmt.Printf("Deleted custom registry %q\n", name)
		return nil
	},
}

var registryTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test latency of all registries",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		fmt.Println("Testing registry latency...")
		results := registry.Test(toRegistries(cfg.Registries))

		headers := []string{"Registry", "URL", "Latency"}
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

func init() {
	rootCmd.AddCommand(registryCmd)
	registryCmd.AddCommand(registryLsCmd)
	registryCmd.AddCommand(registryUseCmd)
	registryCmd.AddCommand(registryAddCmd)
	registryCmd.AddCommand(registryDelCmd)
	registryCmd.AddCommand(registryTestCmd)
}

// toRegistries converts config.Registry slice to internal Registry slice
func toRegistries(cfgRegs []config.Registry) []registry.Registry {
	regs := make([]registry.Registry, len(cfgRegs))
	for i, r := range cfgRegs {
		regs[i] = registry.Registry{
			Name:      r.Name,
			URL:       r.URL,
			IsBuiltIn: false,
		}
	}
	return regs
}

// renderTable prints a simple aligned table to stdout
func renderTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Print(h)
		if i < len(headers)-1 {
			fmt.Print(strings.Repeat(" ", widths[i]-len(h)+2))
		}
	}
	fmt.Println()

	// Print separator
	for i, w := range widths {
		fmt.Print(strings.Repeat("─", w))
		if i < len(widths)-1 {
			fmt.Print("  ")
		}
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Print(cell)
			if i < len(row)-1 {
				fmt.Print(strings.Repeat(" ", widths[i]-len(cell)+2))
			}
		}
		fmt.Println()
	}
}

var _ = os.Stdout // ensure os imported
