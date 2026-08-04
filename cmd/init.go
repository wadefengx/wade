package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
	golang "github.com/wadefengx/wade/internal/go"
	"github.com/wadefengx/wade/internal/node"
	"github.com/wadefengx/wade/internal/python"
	"github.com/wadefengx/wade/internal/registry"
)

func runInteractiveWizard(cmd *cobra.Command, args []string) error {
	autoYes, _ := cmd.Flags().GetBool("yes")
	cfg, _ := config.Load()

	fmt.Println("◇  Welcome to Wade!")
	fmt.Println()

	// ── Step 0: Choose runtimes ──
	runtimes := []string{}
	if autoYes {
		runtimes = []string{"Node.js", "Go", "Python"}
	} else {
		prompt := &survey.Select{
			Message: "Which runtimes to configure?",
			Options: []string{"Node.js", "Go", "Python", "All of the above"},
		}
		survey.AskOne(prompt, &runtimes)
	}

	hasNode := contains(runtimes, "Node.js") || contains(runtimes, "All of the above")
	hasGo := contains(runtimes, "Go") || contains(runtimes, "All of the above")
	hasPython := contains(runtimes, "Python") || contains(runtimes, "All of the above")
	fmt.Println()

	// ── Node.js ──
	if hasNode {
		// Node mirror
		if autoYes {
			cfg.NodeMirror = "https://npmmirror.com/mirrors/node/"
		} else {
			mirrorOpt := ""
			prompt := &survey.Select{
				Message: "Node.js download source:",
				Options: []string{"mirror (npmmirror.com) — fast in China, recommended", "official (nodejs.org)"},
			}
			survey.AskOne(prompt, &mirrorOpt)
			if strings.Contains(mirrorOpt, "official") {
				cfg.NodeMirror = "https://nodejs.org/dist/"
			} else {
				cfg.NodeMirror = "https://npmmirror.com/mirrors/node/"
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
			// Check if any LTS is installed
			allInstalled := true
			for _, ver := range []string{"22", "20", "18"} {
				if !alreadyInstalled(installed, ver) {
					allInstalled = false
					break
				}
			}
			if !allInstalled || len(installed) == 0 {
				opts := []string{"v22 (LTS) — recommended", "v20 (LTS)", "v18 (LTS)", "Skip"}
				verOpt := ""
				prompt := &survey.Select{Message: "Node.js version:", Options: opts}
				survey.AskOne(prompt, &verOpt)

				ver := ""
				switch {
				case strings.HasPrefix(verOpt, "v22"):
					ver = "22"
				case strings.HasPrefix(verOpt, "v20"):
					ver = "20"
				case strings.HasPrefix(verOpt, "v18"):
					ver = "18"
				}
				if ver != "" && !alreadyInstalled(installed, ver) {
					if resolved, err := node.ResolveVersion(ver, cfg.NodeMirror); err == nil {
						fmt.Printf("◇  Installing Node %s...\n", ver)
						if err := node.Install(resolved, cfg.NodeMirror); err == nil {
							cfg.DefaultVersion = resolved
							config.Save(cfg)
							node.UseVersion(resolved)
						}
					}
				}
			}
		}
		fmt.Println()

		// Registry mirror
		if autoYes {
			cfg.CurrentRegistry = "taobao"
		} else {
			allRegs := registry.All(toRegistries(nil))
			opts := make([]string, len(allRegs))
			for i, r := range allRegs {
				opts[i] = fmt.Sprintf("%s (%s)", r.Name, r.URL)
			}
			regOpt := ""
			prompt := &survey.Select{Message: "Registry mirror (npm/yarn/pnpm):", Options: opts}
			survey.AskOne(prompt, &regOpt)
			for _, r := range allRegs {
				if strings.HasPrefix(regOpt, r.Name+" ") {
					cfg.CurrentRegistry = r.Name
					break
				}
			}
		}
		config.Save(cfg)
		registry.Use(cfg.CurrentRegistry)
		fmt.Println()
	}

	// ── Go ──
	if hasGo {
		if autoYes {
			cfg.GoMirror = "https://golang.google.cn/dl/"
			config.Save(cfg)
			python.UseGoProxy("goproxy.cn")
			fmt.Println("   ✓ Go mirror:  google-cn")
			fmt.Println("   ✓ Go proxy:   goproxy.cn")
		} else {
			// Go mirror
			mirrors := python.GoMirrorPresets()
			opts := make([]string, len(mirrors))
			for i, m := range mirrors {
				label := m.Name
				if m.Name == "google-cn" {
					label += " — recommended in China"
				}
				opts[i] = label
			}
			mirrorOpt := ""
			prompt := &survey.Select{Message: "Go download source:", Options: opts}
			survey.AskOne(prompt, &mirrorOpt)
			for _, m := range mirrors {
				if strings.HasPrefix(mirrorOpt, m.Name) {
					cfg.GoMirror = m.URL
					break
				}
			}
			config.Save(cfg)

			// Go proxy
			proxies := python.GoProxyPresets()
			popts := make([]string, len(proxies))
			for i, p := range proxies {
				label := p.Name
				if p.Name == "goproxy.cn" {
					label += " — recommended in China"
				}
				popts[i] = label
			}
			proxyOpt := ""
			pprompt := &survey.Select{Message: "Go proxy (GOPROXY):", Options: popts}
			survey.AskOne(pprompt, &proxyOpt)
			for _, p := range proxies {
				if strings.HasPrefix(proxyOpt, p.Name) {
					python.UseGoProxy(p.Name)
					break
				}
			}
		}
		fmt.Println()
	}

	// ── Python ──
	if hasPython {
		if autoYes {
			python.UsePipMirror("tsinghua")
			fmt.Println("   ✓ pip mirror: tsinghua")
		} else {
			mirrors := python.PipPresets()
			opts := make([]string, len(mirrors))
			for i, m := range mirrors {
				label := m.Name
				if m.Name == "tsinghua" {
					label += " — recommended in China"
				}
				opts[i] = label
			}
			mirrorOpt := ""
			prompt := &survey.Select{Message: "Python pip registry:", Options: opts}
			survey.AskOne(prompt, &mirrorOpt)
			for _, m := range mirrors {
				if strings.HasPrefix(mirrorOpt, m.Name) {
					python.UsePipMirror(m.Name)
					break
				}
			}
		}
		fmt.Println()
	}

	// ── PATH ──
	shimDir, _ := node.ShimDir()
	if runtime.GOOS == "windows" {
		// Windows: shims go to the FRONT of the user PATH env var (registry).
		// Shell rc files are useless here — cmd doesn't read PowerShell profile,
		// and bash-style 'export PATH' lines don't work in PowerShell either.
		if !strings.Contains(os.Getenv("PATH"), shimDir) {
			addPath := true
			if !autoYes {
				prompt := &survey.Confirm{
					Message: "Add ~/.wade/shims to the FRONT of your user PATH (works in cmd + PowerShell)?",
					Default: true,
				}
				survey.AskOne(prompt, &addPath)
			}
			if addPath {
				if err := runSetup(); err != nil {
					return fmt.Errorf("PATH setup: %w", err)
				}
			}
		}
	} else if !strings.Contains(os.Getenv("PATH"), shimDir) {
		shell := detectShell()
		rcFile := shellConfigPath(shell)
		if rcFile != "" {
			addPath := true
			if !autoYes {
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("Add ~/.wade/shims to %s?", filepath.Base(rcFile)),
					Default: true,
				}
				survey.AskOne(prompt, &addPath)
			}
			if addPath {
				appendToFile(rcFile, fmt.Sprintf("\n# wade — runtime manager\nexport PATH=\"%s:$PATH\"\n", shimDir))
				fmt.Printf("✔  Added to %s\n", rcFile)
			}
		}
	}
	fmt.Println()

	// ── Auto PATH setup (init -y): make it truly one-command install ──
	if autoYes {
		fmt.Println("── Setting up PATH ──")
		if err := runSetup(); err != nil {
			fmt.Printf("⚠️  PATH setup skipped: %v\n", err)
		}
		fmt.Println()
	}

	// ── Summary ──
	fmt.Println("✔  Setup complete!")
	fmt.Println()
	cfg, _ = config.Load()
	cur, _ := node.CurrentVersion()
	curReg, curURL, _ := registry.GetCurrent()
	fmt.Printf("   Node:      %s\n", cur)
	fmt.Printf("   Registry:  %s → %s\n", curReg, curURL)
	if hasGo {
		goVer, _ := golang.CurrentVersion()
		if goVer == "" {
			goVer = "(system — not managed by wade)"
		}
		goMirror := cfg.GoMirror
		for _, m := range python.GoMirrorPresets() {
			if strings.HasPrefix(cfg.GoMirror, m.URL) {
				goMirror = m.Name
				break
			}
		}
		fmt.Printf("   Go:        %s\n", goVer)
		fmt.Printf("   Go mirror: %s\n", goMirror)
	}
	if hasPython {
		fmt.Printf("   Python:    (system — detected, not installed by wade)\n")
	}
	fmt.Println()
	fmt.Println("   Quick: wade status | wade node ls | wade go ls | wade python ls")
	fmt.Println()
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func alreadyInstalled(versions []string, prefix string) bool {
	for _, v := range versions {
		if strings.HasPrefix(v, "v"+prefix) {
			return true
		}
	}
	return false
}

var _ = os.Stdout
