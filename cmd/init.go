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
	"github.com/wadefengx/wade/internal/python"
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

	cfg, _ := config.Load()

	fmt.Println("◇  Welcome to Wade!")
	fmt.Println()

	// ── Step 0: Choose runtimes ──
	fmt.Println("◇  Which runtimes to configure?")
	fmt.Println("│  1. Node.js  (recommended for frontend)")
	fmt.Println("│  2. Go       (recommended for backend/CLI)")
	fmt.Println("│  3. Python   (recommended for data/scripts)")
	fmt.Println("│  4. All of the above")
	choice := input("└  1-4 › ")

	hasNode := choice == "1" || choice == "4"
	hasGo := choice == "2" || choice == "4"
	hasPython := choice == "3" || choice == "4"

	// Determine defaults
	if autoYes {
		hasNode = true
		hasGo = true
		hasPython = true
	}

	fmt.Println()

	// ── Node.js configuration ──
	if hasNode {
		// Node mirror
		cfg.NodeMirror = "https://npmmirror.com/mirrors/node/"
		if !autoYes {
			fmt.Println("◇  Node.js download source:")
			fmt.Println("│  1. Mirror (npmmirror.com) — fast in China, recommended")
			fmt.Println("│  2. Official (nodejs.org)")
			c := input("└  1-2 › ")
			if c == "2" {
				cfg.NodeMirror = "https://nodejs.org/dist/"
			}
		}
		config.Save(cfg)

		// Node version
		installed, _ := node.InstalledVersions()
		if autoYes {
			// Auto-install latest LTS (try 22, then 20)
			for _, ver := range []string{"22", "20", "18"} {
				if !alreadyInstalled(installed, ver) {
					if resolved, err := node.ResolveVersion(ver, cfg.NodeMirror); err == nil {
						fmt.Printf("◇  Installing Node %s...\n", ver)
						if err := node.Install(resolved, cfg.NodeMirror); err == nil {
							cfg.DefaultVersion = resolved
							config.Save(cfg)
							node.UseVersion(resolved)
						}
					}
					break
				}
			}
		} else {
			fmt.Println("◇  Node.js version:")
			recs := []struct{ label, ver string }{{"v22 (LTS) — recommended", "22"}, {"v20 (LTS)", "20"}, {"v18 (LTS)", "18"}}
			for i, r := range recs {
				m := ""
				if alreadyInstalled(installed, r.ver) {
					m = " [installed]"
				}
				fmt.Printf("│  %d. %s%s\n", i+1, r.label, m)
			}
			fmt.Println("│  4. Skip")
			c := input("└  1-4 › ")
			if idx, _ := strconv.Atoi(c); idx >= 1 && idx <= 3 {
				ver := recs[idx-1].ver
				if !alreadyInstalled(installed, ver) {
					if resolved, err := node.ResolveVersion(ver, cfg.NodeMirror); err == nil {
						fmt.Printf("◇  Installing Node %s...\n", ver)
						if err := node.Install(resolved, cfg.NodeMirror); err == nil {
							cfg.DefaultVersion = resolved
							config.Save(cfg)
							node.UseVersion(resolved)
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
		}
		fmt.Println()

		// Registry mirror
		cfg.CurrentRegistry = "taobao"
		if !autoYes {
			results := registry.Test(toRegistries(nil))
			fastest := ""
			if len(results) > 0 && results[0].Error == "" {
				fastest = results[0].Name
			}
			allRegs := registry.All(toRegistries(nil))
			fmt.Println("◇  Registry mirror (npm/yarn/pnpm):")
			for i, r := range allRegs {
				if i >= 5 {
					break
				}
				s := ""
				if r.Name == fastest {
					s = " ★ fastest"
				}
				fmt.Printf("│  %d. %s%s\n", i+1, r.Name, s)
			}
			c := input("└  1-5 › ")
			if idx, _ := strconv.Atoi(c); idx >= 1 && idx <= 5 {
				cfg.CurrentRegistry = allRegs[idx-1].Name
			}
		}
		config.Save(cfg)
		registry.Use(cfg.CurrentRegistry)
		fmt.Println()
	}

	// ── Go configuration ──
	if hasGo {
		cfg.GoMirror = "https://golang.google.cn/dl/"
		config.Save(cfg)
		if !autoYes {
			fmt.Println("◇  Go download source:")
			for i, m := range python.GoMirrorPresets() {
				mark := ""
				if m.Name == "google-cn" {
					mark = " — recommended in China"
				}
				fmt.Printf("│  %d. %s%s\n", i+1, m.Name, mark)
			}
			c := input("└  1-4 › ")
			if idx, _ := strconv.Atoi(c); idx >= 1 && idx <= 4 {
				cfg.GoMirror = python.GoMirrorPresets()[idx-1].URL
				config.Save(cfg)
			}
		}
		fmt.Println("🌐 Go mirror: google-cn ✓")

		// Go proxy
		if !autoYes {
			fmt.Println("◇  Go proxy (GOPROXY):")
			for i, p := range python.GoProxyPresets() {
				mark := ""
				if p.Name == "goproxy.cn" {
					mark = " — recommended in China"
				}
				fmt.Printf("│  %d. %s%s\n", i+1, p.Name, mark)
			}
			c := input("└  1-3 › ")
			if idx, _ := strconv.Atoi(c); idx >= 1 && idx <= 3 {
				python.UseGoProxy(python.GoProxyPresets()[idx-1].Name)
			}
		} else {
			python.UseGoProxy("goproxy.cn")
		}
		fmt.Println("🌐 Go proxy: goproxy.cn ✓")
		fmt.Println()
	}

	// ── Python configuration ──
	if hasPython {
		if !autoYes {
			fmt.Println("◇  Python pip registry:")
			for i, m := range python.PipPresets() {
				mark := ""
				if m.Name == "tsinghua" {
					mark = " — recommended in China"
				}
				fmt.Printf("│  %d. %s%s\n", i+1, m.Name, mark)
			}
			c := input("└  1-6 › ")
			if idx, _ := strconv.Atoi(c); idx >= 1 && idx <= 6 {
				python.UsePipMirror(python.PipPresets()[idx-1].Name)
			}
		} else {
			python.UsePipMirror("tsinghua")
		}
		fmt.Println("🐍 pip registry: tsinghua ✓")
		fmt.Println()
	}

	// ── PATH ──
	shimDir, _ := node.ShimDir()
	if !strings.Contains(os.Getenv("PATH"), shimDir) {
		shell := detectShell()
		rcFile := shellConfigPath(shell)
		if rcFile != "" {
			fmt.Printf("◇  Add ~/.wade/shims to %s?\n", filepath.Base(rcFile))
			choice = input("└  Y/n › ")
			if choice == "" || strings.ToLower(choice) == "y" {
				appendToFile(rcFile, fmt.Sprintf("\n# wade — runtime manager\nexport PATH=\"%s:$PATH\"\n", shimDir))
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
	fmt.Printf("   Go mirror: %s\n", cfg.GoMirror)
	fmt.Println("   Go proxy:  goproxy.cn")
	fmt.Println("   pip:       tsinghua")
	fmt.Println()
	fmt.Println("   Quick: wade status | wade node ls | wade go ls | wade python ls")
	fmt.Println()
	return nil
}

func alreadyInstalled(versions []string, prefix string) bool {
	for _, v := range versions {
		if strings.HasPrefix(v, "v"+prefix) {
			return true
		}
	}
	return false
}