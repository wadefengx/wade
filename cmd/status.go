package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	golang "github.com/wadefengx/wade/internal/go"
	"github.com/wadefengx/wade/internal/node"
	"github.com/wadefengx/wade/internal/python"
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

		regName, regURL, _ := registry.GetCurrent()
		cfgPath, _ := config.ConfigPath()
		wadeDir, _ := config.WadeDir()

		fmt.Println("🏄  wade status")
		fmt.Println("─────────────────────")

		// Node
		nodeVer, _ := node.CurrentVersion()
		if nodeVer != "" {
			fmt.Printf("  🟢 Node:        %s", nodeVer)
			if cfg.DefaultVersion == nodeVer {
				fmt.Print(" (default)")
			}
			fmt.Println()
		}
		fmt.Printf("  📦 Registry:    %s → %s\n", regName, regURL)

		// Go
		goVer, _ := golang.CurrentVersion()
		if goVer != "" {
			fmt.Printf("  🔵 Go:          %s", goVer)
			if cfg.DefaultGoVersion == goVer {
				fmt.Print(" (default)")
			}
			fmt.Println()
		} else {
			// Show system Go if not managed by wade
			if sysGo := python.DetectSystemGo(); sysGo != "" {
				fmt.Printf("  🔵 Go:          %s (system)\n", sysGo)
			}
		}
		if cfg.GoMirror != "" {
			label := "custom"
			for _, m := range python.GoMirrorPresets() {
				if strings.HasPrefix(cfg.GoMirror, m.URL) {
					label = m.Name
					break
				}
			}
			fmt.Printf("  🌐 Go mirror:   %s\n", label)
		}

		// Python
		pythons := python.DetectSystemPython()
		if len(pythons) > 0 {
			var versions []string
			for _, p := range pythons {
				parts := strings.SplitN(p, " (system:", 2)
				if len(parts) == 2 {
					versions = append(versions, strings.TrimSuffix(strings.TrimSpace(parts[1]), ")"))
				}
			}
			if len(versions) > 0 {
				fmt.Printf("  🐍 Python:      %s\n", strings.Join(versions, ", "))
			}
		}

		// Config
		fmt.Printf("  ⚙️  Config:      %s\n", cfgPath)
		fmt.Printf("  📁 Wade dir:    %s\n", wadeDir)

		fmt.Println()
		fmt.Println("  💡 Try: wade -i | wade go ls | wade python registry ls")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}