package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/wadefengx/wade/internal/config"
)

var (
	setupAuto   bool
	setupDryRun bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "One-command setup: create directories, configure PATH, get ready",
	Long: `Interactive setup that:
1. Creates ~/.wade/ directory structure
2. Adds ~/.wade/shims to your shell PATH
3. Verifies everything is working

Run 'wade setup --auto' to skip prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interactive := !setupAuto

		fmt.Println("🏄 Wade Setup")
		fmt.Println("─────────────")
		fmt.Println()

		// Step 1: Create Wade directories
		wadeDir, _ := config.WadeDir()
		shimDir, _ := filepath.Abs(filepath.Join(wadeDir, "shims"))

		fmt.Printf("📁 Creating %s ... ", wadeDir)
		if err := os.MkdirAll(shimDir, 0755); err != nil {
			return fmt.Errorf("create directory: %w", err)
		}
		fmt.Println("✓")

		// Step 2: Create default config if not exists
		cfgPath, _ := config.ConfigPath()
		if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
			fmt.Printf("📄 Creating %s ... ", cfgPath)
			if err := config.Save(config.DefaultConfig()); err != nil {
				return fmt.Errorf("create config: %w", err)
			}
			fmt.Println("✓")
		} else {
			fmt.Println("📄 Config already exists ✓")
		}

		// Step 3: Detect shell
		shell := detectShell()

		// Windows: add shims to the USER PATH env var (registry-backed) —
		// works for cmd AND PowerShell, unlike shell profile files.
		if runtime.GOOS == "windows" {
			shimAbs, _ := filepath.Abs(shimDir)
			psCmd := fmt.Sprintf(
				`$p=[Environment]::GetEnvironmentVariable('Path','User'); if($p -notlike '*%s*'){ [Environment]::SetEnvironmentVariable('Path', $p+';%s','User'); Write-Output 'added' } else { Write-Output 'exists' }`,
				strings.ReplaceAll(shimAbs, `\`, `\\`),
				strings.ReplaceAll(shimAbs, `\`, `\\`),
			)
			out, err := exec.Command("powershell", "-NoProfile", "-Command", psCmd).CombinedOutput()
			if err == nil && strings.Contains(string(out), "added") {
				fmt.Printf("✅ Added %s to user PATH (cmd + PowerShell)\n", shimAbs)
			} else if err == nil && strings.Contains(string(out), "exists") {
				fmt.Printf("✅ %s already in user PATH\n", shimAbs)
			} else {
				fmt.Println("⚠️  Could not update user PATH automatically.")
				fmt.Printf("   Add %s to your PATH manually (System Settings → Environment Variables).\n", shimAbs)
			}
			fmt.Println()
			fmt.Println("⚠️  PATH updated — open a NEW cmd/PowerShell window, then run:")
			fmt.Println("   wade node use <version>")
			fmt.Println()
			return nil
		}
		fmt.Printf("🐚 Detected shell: %s\n", shell)

		// Step 4: Add to PATH
		shellRc := shellConfigPath(shell)
		if shellRc == "" {
			fmt.Println()
			fmt.Println("⚠️  Could not detect shell config file. Please add this line manually:")
			fmt.Println()
			fmt.Printf("   export PATH=\"%s:$PATH\"\n", shimDir)
			fmt.Println()
			return nil
		}

		needAdd := true
		if data, err := os.ReadFile(shellRc); err == nil {
			if strings.Contains(string(data), ".wade/shims") {
				fmt.Printf("✅ PATH already configured in %s\n", shellRc)
				needAdd = false
			}
		}

		if needAdd {
			pathLine := fmt.Sprintf("\n# wade - Node version manager\nexport PATH=\"%s:$PATH\"\n", shimDir)

			if interactive && !setupDryRun {
				fmt.Println()
				fmt.Printf("Add this to %s?\n", shellRc)
				fmt.Printf("   %s\n", strings.TrimSpace(pathLine))
				fmt.Print("   [Y/n]: ")

				reader := bufio.NewReader(os.Stdin)
				response, _ := reader.ReadString('\n')
				response = strings.TrimSpace(strings.ToLower(response))
				if response == "n" || response == "no" {
					fmt.Println("   Skipped. You can add it later manually.")
				} else {
					appendToFile(shellRc, pathLine)
				}
			} else {
				// Auto mode: add it
				appendToFile(shellRc, pathLine)
			}
		}

		// Step 5: Verify
		fmt.Println()
		fmt.Println("✅ Setup complete!")
		fmt.Println()
		fmt.Println("Next steps:")
		if needAdd {
			fmt.Printf("   1. Run: source %s\n", shellRc)
		}
		fmt.Println("   2. Try: wade version")
		fmt.Println("   3. Try: wade node install 20")
		fmt.Println()
		return nil
	},
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			return "powershell"
		}
		return "unknown"
	}
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	return filepath.Base(shell)
}

func shellConfigPath(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		// Check for .zshrc, fallback to .zprofile
		zshrc := filepath.Join(home, ".zshrc")
		if _, err := os.Stat(zshrc); err == nil {
			return zshrc
		}
		return zshrc // will be created if needed
	case "bash":
		return filepath.Join(home, ".bashrc")
	case "fish":
		return filepath.Join(home, ".config", "fish", "config.fish")
	case "powershell":
		return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	default:
		return ""
	}
}

func appendToFile(path, content string) error {
	fmt.Printf("   ✏️  Writing to %s ... ", filepath.Base(path))

	// Read existing
	existing, _ := os.ReadFile(path)

	// Append
	if !strings.HasSuffix(string(existing), "\n") && len(existing) > 0 {
		content = "\n" + content
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	fmt.Println("✓")
	return nil
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolVar(&setupAuto, "auto", false, "Skip all prompts, auto-configure")
	setupCmd.Flags().BoolVar(&setupDryRun, "dry-run", false, "Show what would be done without doing it")
}
