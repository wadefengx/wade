package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/node"
	"github.com/wadefengx/wade/internal/registry"
)

func runInteractiveWizard(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	autoYes, _ := cmd.Flags().GetBool("yes")

	input := func(prefix string) string {
		fmt.Print(prefix)
		if autoYes {
			fmt.Println()
			return ""
		}
		s, _ := reader.ReadString('\n')
		return strings.TrimSpace(s)
	}

	fmt.Println("◇  Welcome to Wade!")
	fmt.Println()

	cfg, _ := config.Load()

	// ── Step 1: Node mirror ──
	fmt.Println("◇  Node.js download source:")
	fmt.Println("│  1. Mirror (npmmirror.com) — fast in China, recommended")
	fmt.Println("│  2. Official (nodejs.org)")
	choice := input("└  1-2 › ")

	if choice == "2" {
		cfg.NodeMirror = "https://nodejs.org/dist/"
	} else {
		cfg.NodeMirror = "https://npmmirror.com/mirrors/node/"
	}
	config.Save(cfg)
	fmt.Println()

	// ── Step 2: Node version ──
	fmt.Println("◇  Node.js version:")
	recommended := []struct {
		label string
		ver   string
	}{{"v22 (LTS) — recommended", "22"}, {"v20 (LTS)", "20"}, {"v18 (LTS)", "18"}}
	installed, _ := node.InstalledVersions()

	for i, r := range recommended {
		marker := ""
		for _, v := range installed {
			if strings.HasPrefix(v, "v"+r.ver) {
				marker = " [installed]"
				break
			}
		}
		fmt.Printf("│  %d. %s%s\n", i+1, r.label, marker)
	}
	fmt.Println("│  4. Skip")
	choice = input("└  1-4 › ")

	if idx := parseChoice(choice, 1, 4); idx >= 1 && idx <= 3 {
		ver := recommended[idx-1].ver
		already := false
		for _, v := range installed {
			if strings.HasPrefix(v, "v"+ver) {
				already = true
				break
			}
		}
		if !already {
			fmt.Println()
			fmt.Printf("◇  Installing Node %s...\n", ver)
			resolved, err := node.ResolveVersion(ver, cfg.NodeMirror)
			if err == nil {
				if err := node.Install(resolved, cfg.NodeMirror); err == nil {
					cfg.DefaultVersion = resolved
					config.Save(cfg)
					node.UseVersion(resolved)
					fmt.Printf("✔  Node %s installed\n", resolved)
				}
			}
		} else {
			for _, v := range installed {
				if strings.HasPrefix(v, "v"+ver) {
					node.UseVersion(v)
					break
				}
			}
		}
	}
	fmt.Println()

	// ── Step 3: Registry ──
	fmt.Println("◇  Registry mirror:")
	results := registry.Test(toRegistries(nil))
	fastest := ""
	if len(results) > 0 && results[0].Error == "" {
		fastest = results[0].Name
	}

	allRegs := registry.All(toRegistries(nil))
	for i, r := range allRegs {
		if i >= 5 {
			break
		}
		speed := ""
		if r.Name == fastest {
			speed = " ★ fastest"
		}
		marker := ""
		if r.Name == cfg.CurrentRegistry {
			marker = " [current]"
		}
		fmt.Printf("│  %d. %s%s%s\n", i+1, r.Name, speed, marker)
	}
	choice = input("└  1-5 › ")

	if idx := parseChoice(choice, 1, 5); idx >= 1 {
		regName := allRegs[idx-1].Name
		if regName != cfg.CurrentRegistry {
			if err := registry.Use(regName); err != nil {
				fmt.Printf("⚠  %v\n", err)
			}
		}
	}
	fmt.Println()

	// ── Step 4: PATH ──
	shimDir, _ := node.ShimDir()
	if !strings.Contains(os.Getenv("PATH"), shimDir) {
		shell := detectShell()
		rcFile := shellConfigPath(shell)
		if rcFile != "" {
			fmt.Printf("◇  Add ~/.wade/shims to %s?\n", filepath.Base(rcFile))
			choice = input("└  Y/n › ")
			if choice == "" || strings.ToLower(choice) == "y" {
				appendToFile(rcFile, fmt.Sprintf("\n# wade — Node version manager\nexport PATH=\"%s:$PATH\"\n", shimDir))
				fmt.Printf("✔  Added to %s\n", rcFile)
			}
		}
	}
	fmt.Println()

	// ── Summary ──
	fmt.Println("✔  Setup complete!")
	fmt.Println()
	cfg, _ = config.Load()
	cur, _ := node.CurrentVersion()
	curReg, curURL, _ := registry.GetCurrent()
	fmt.Printf("   Node:      %s\n", cur)
	fmt.Printf("   Registry:  %s → %s\n", curReg, curURL)
	mirrorLabel := "npmmirror.com"
	if strings.Contains(cfg.NodeMirror, "nodejs.org") {
		mirrorLabel = "nodejs.org"
	}
	fmt.Printf("   Mirror:    %s\n", mirrorLabel)
	fmt.Println()
	fmt.Println("   Quick start: wade status | wade node ls | wade registry test")
	fmt.Println()
	return nil
}

func mustConfigPath() string {
	p, _ := config.ConfigPath()
	return p
}

func parseChoice(s string, min, max int) int {
	n, err := strconv.Atoi(s)
	if err != nil || n < min || n > max {
		return 0
	}
	return n
}
