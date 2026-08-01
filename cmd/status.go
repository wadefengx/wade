package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
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

		// Wade itself
		fmt.Printf("  🏄 Wade:        %s\n", displayVersion())

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

		// Shim health: is ~/.wade/shims actually on PATH?
		if nodeVer != "" {
			shimDir, _ := node.ShimDir()
			switch {
			case pathInEnvPath(shimDir):
				// shims on current session PATH — check actual node resolution
				if p := whichNode(); p != "" && !strings.Contains(strings.ToLower(p), "shims") {
					fmt.Println()
					fmt.Printf("  ⚠️  'node' resolves to %s (system), not wade's shim!\n", p)
					fmt.Println("      Ensure ~/.wade/shims is FIRST in your PATH (it must come before Program Files\\nodejs).")
					fmt.Println("      Fix: wade setup --auto, then open a NEW terminal.")
				}
			case userPathHasShims(shimDir):
				fmt.Println()
				fmt.Println("  ⚠️  shims are in your user PATH, but this window has the OLD PATH.")
				fmt.Println("      Close this window and open a NEW cmd/PowerShell — then 'node' will be wade's.")
			default:
				fmt.Println()
				fmt.Println("  ⚠️  ~/.wade/shims is NOT on your PATH — 'node' is the system version, not wade's!")
				fmt.Println("      Fix: wade setup --auto, then open a NEW terminal.")
			}
		}

		fmt.Println()
		fmt.Println("  💡 Try: wade -i | wade go ls | wade python registry ls")
		return nil
	},
}

// displayVersion returns the wade version string, falling back to "dev".
func displayVersion() string {
	if version == "" || version == "dev" {
		return "dev (built from source)"
	}
	return version
}

// userPathHasShims reports whether shimDir is in the persisted USER PATH
// (registry HKCU\Environment). Windows only — on unix returns false so the
// current-session PATH check governs.
func userPathHasShims(shimDir string) bool {
	if runtime.GOOS != "windows" {
		return false
	}
	out, err := exec.Command("reg", "query", `HKCU\Environment`, "/v", "Path").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), strings.ToLower(shimDir))
}

// whichNode returns the path 'node' resolves to, or "" if not found.
func whichNode() string {
	p, err := exec.LookPath("node")
	if err != nil {
		return ""
	}
	return p
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
