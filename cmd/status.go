package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/registry"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current environment status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		name, url, _ := registry.GetCurrent()
		cfgPath, _ := config.ConfigPath()
		wadeDir, _ := config.WadeDir()

		fmt.Println("🏄  wade status")
		fmt.Println("─────────────────────")
		fmt.Printf("  📦 Registry:      %s → %s\n", name, url)
		fmt.Printf("  ⚙️  Config:       %s\n", cfgPath)
		fmt.Printf("  📁 Wade dir:      %s\n", wadeDir)
		fmt.Printf("  🌐 Node mirror:   %s\n", cfg.NodeMirror)

		if len(cfg.Registries) > 0 {
			fmt.Printf("  ⭐ Custom regs:   %d\n", len(cfg.Registries))
		}

		if cfg.DefaultVersion != "" {
			fmt.Printf("  🟢 Node ver:      %s (default)\n", cfg.DefaultVersion)
		}

		fmt.Println()
		fmt.Println("  💡 Tip: 'wade registry use taobao' to switch mirrors")
		fmt.Println("     'wade node install 22' to add a Node version")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
