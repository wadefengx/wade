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

	// Check for --yes / -y flag
	autoYes, _ := cmd.Flags().GetBool("yes")

	input := func(prompt string) string {
		if autoYes {
			fmt.Print(prompt)
			fmt.Println(" [y]")
			return "y"
		}
		fmt.Print(prompt)
		s, _ := reader.ReadString('\n')
		return strings.TrimSpace(s)
	}

	fmt.Println("🏄  Welcome to Wade!")
	fmt.Println("   Let's get your Node.js environment set up.")
	fmt.Println()

	cfg, _ := config.Load()

	// ── Step 1: Node download mirror ──
	fmt.Println("🌐 Step 1: Node.js download source")
	fmt.Println()
	fmt.Println("   Where should wade download Node.js from?")
	fmt.Println("     1.  Mirror (npmmirror.com)  ← fast in China, default")
	fmt.Println("     2.  Official (nodejs.org)    ← global")
	fmt.Println()

	choice := input("   Choose [1-2]: ")
	switch choice {
	case "2":
		cfg.NodeMirror = "https://nodejs.org/dist/"
		fmt.Println("   ✅ Using official nodejs.org")
	default:
		cfg.NodeMirror = "https://npmmirror.com/mirrors/node/"
		fmt.Println("   ✅ Using npmmirror.com (fast in China)")
	}
	config.Save(cfg)
	fmt.Println()

	// ── Step 2: Node.js version ──
	fmt.Println("📦 Step 2: Node.js version")
	fmt.Println()

	installed, _ := node.InstalledVersions()
	if len(installed) > 0 {
		fmt.Printf("   Already installed: %s\n", strings.Join(installed, ", "))
		fmt.Println()
	}

	recommended := []struct {
		label string
		ver   string
	}{{"v22 (LTS)", "22"}, {"v20 (LTS)", "20"}, {"v18 (LTS)", "18"}}

	fmt.Println("   Recommended LTS versions:")
	for i, r := range recommended {
		installedMark := ""
		for _, v := range installed {
			if strings.HasPrefix(v, "v"+r.ver) {
				installedMark = " ✓ installed"
				break
			}
		}
		fmt.Printf("     %d. %s%s\n", i+1, r.label, installedMark)
	}
	fmt.Println("     4. Skip (keep current)")
	fmt.Println()

	choice = input("   Choose [1-4]: ")
	if idx := parseChoice(choice, 1, 4); idx >= 1 && idx <= 3 {
		ver := recommended[idx-1].ver
		alreadyInstalled := false
		for _, v := range installed {
			if strings.HasPrefix(v, "v"+ver) {
				alreadyInstalled = true
				break
			}
		}
		if !alreadyInstalled {
			fmt.Printf("\n   ⬇️  Installing Node %s...\n", ver)
			resolved, err := node.ResolveVersion(ver, cfg.NodeMirror)
			if err != nil {
				fmt.Printf("   ⚠️  %v\n", err)
			} else if err := node.Install(resolved, cfg.NodeMirror); err != nil {
				fmt.Printf("   ⚠️  %v\n", err)
			} else {
				cfg.DefaultVersion = resolved
				config.Save(cfg)
				node.UseVersion(resolved)
				fmt.Printf("   ✅ Node %s installed and activated\n", resolved)
			}
		} else {
			fmt.Printf("\n   ✅ Node %s already installed\n", ver)
			for _, v := range installed {
				if strings.HasPrefix(v, "v"+ver) {
					node.UseVersion(v)
					break
				}
			}
		}
	}
	fmt.Println()

	// ── Step 3: Registry mirror ──
	fmt.Println("📦 Step 3: Registry mirror (npm install source)")
	fmt.Println()

	fmt.Println("   Testing registry speeds...")
	results := registry.Test(toRegistries(nil))
	if len(results) > 0 && results[0].Error == "" {
		fmt.Printf("   Fastest: %s (%s)\n", results[0].Name, results[0].Latency.Round(1000000))
	}
	fmt.Println()

	allRegs := registry.All(toRegistries(nil))
	fmt.Println("   Choose registry for npm/yarn/pnpm:")
	for i, r := range allRegs {
		if i >= 5 {
			break
		}
		marker := ""
		if r.Name == cfg.CurrentRegistry {
			marker = " ← current"
		}
		fmt.Printf("     %d. %s%s\n", i+1, r.Name, marker)
	}
	fmt.Println()

	choice = input(fmt.Sprintf("   Choose [1-5]: "))
	if idx := parseChoice(choice, 1, 5); idx >= 1 {
		regName := allRegs[idx-1].Name
		if regName != cfg.CurrentRegistry {
			if err := registry.Use(regName); err != nil {
				fmt.Printf("   ⚠️  %v\n", err)
			} else {
				fmt.Printf("   ✅ Switched registry to %s\n", regName)
			}
		}
	}
	fmt.Println()

	// ── Step 4: PATH ──
	fmt.Println("⚙️  Step 4: PATH")
	fmt.Println()

	shimDir, _ := node.ShimDir()
	pathHasShim := strings.Contains(os.Getenv("PATH"), shimDir)

	if !pathHasShim {
		shell := detectShell()
		rcFile := shellConfigPath(shell)
		if rcFile != "" {
			fmt.Printf("   Add ~/.wade/shims to %s?\n", filepath.Base(rcFile))
			choice = input("   [Y/n]: ")
			if choice == "" || strings.ToLower(choice) == "y" {
				appendToFile(rcFile, fmt.Sprintf("\n# wade — Node version manager\nexport PATH=\"%s:$PATH\"\n", shimDir))
				fmt.Printf("   ✅ Added to %s (run 'source %s' to apply)\n", rcFile, rcFile)
			}
		}
	} else {
		fmt.Println("   ✅ PATH already configured")
	}
	fmt.Println()

	// ── Summary ──
	fmt.Println("✨ All done! Here's your setup:")
	fmt.Println()
	cfg, _ = config.Load()
	cur, _ := node.CurrentVersion()
	curReg, curURL, _ := registry.GetCurrent()
	fmt.Printf("   🟢 Node:     %s (default)\n", cur)
	fmt.Printf("   🌐 Node src: %s\n", cfg.NodeMirror)
	fmt.Printf("   📦 Registry: %s → %s\n", curReg, curURL)
	fmt.Printf("   📁 Config:   %s\n", mustConfigPath())
	fmt.Println()
	fmt.Println("   Try: wade status | wade node ls | wade registry test")
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
