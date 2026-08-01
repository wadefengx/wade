package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wade",
	Short: "All-in-one Node.js version & npm registry manager",
	Long: `Wade manages Node.js versions and npm/yarn/pnpm registries.
Single binary, installed once, no Node.js dependency.`,

	// Check for updates before every command (oh-my-zsh style),
	// cached for 24h so it doesn't slow down interactive use.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if cmd.Name() == "update" || cmd.Name() == "version" {
			return // skip inside update/version itself
		}
		checkForUpdate(false)
	},

	Run: func(cmd *cobra.Command, args []string) {
		// wade -i shortcut
		if skipAll, _ := cmd.Flags().GetBool("init"); skipAll {
			runInteractiveWizard(cmd, args)
			os.Exit(0)
		}
		// wade -u shortcut
		if doUpdate, _ := cmd.Flags().GetBool("update"); doUpdate {
			if err := runUpdate(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			return
		}
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("init", "i", false, "Run interactive setup wizard (alias: wade init)")
	rootCmd.Flags().BoolP("update", "u", false, "Update wade to the latest version (alias: wade update)")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Interactive setup wizard for wade",
		RunE:  runInteractiveWizard,
	}
	initCmd.Flags().BoolP("yes", "y", false, "Skip prompts, use defaults")
	rootCmd.AddCommand(initCmd)
}

// checkForUpdate queries GitHub for the latest release (cached 24h).
// When interactive and a newer version exists, asks the user whether to update.
func checkForUpdate(force bool) {
	if version == "" || version == "dev" {
		return // dev builds skip the check
	}

	// 24h cache so every-command checks don't hit the network each time
	cacheFile := updateCheckCachePath()
	if !force {
		if info, err := os.Stat(cacheFile); err == nil {
			if time.Since(info.ModTime()) < 24*time.Hour {
				return
			}
		}
	}

	client := newHTTPClient(3 * time.Second)
	resp, err := client.Get("https://api.github.com/repos/wadefengx/wade/releases/latest")
	if err != nil {
		return // network failure — stay silent
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}

	body := make([]byte, 4096)
	n, _ := resp.Body.Read(body)
	text := string(body[:n])
	tagStart := strings.Index(text, `"tag_name":"`)
	if tagStart == -1 {
		return
	}
	tagStart += len(`"tag_name":"`)
	tagEnd := strings.Index(text[tagStart:], `"`)
	if tagEnd == -1 {
		return
	}
	latest := text[tagStart : tagStart+tagEnd]

	// Record the check regardless of outcome
	os.MkdirAll(filepath.Dir(cacheFile), 0755)
	os.WriteFile(cacheFile, []byte(latest), 0644)

	if latest == version {
		return
	}

	fmt.Printf("\n✨ New version available: wade %s → %s\n", version, latest)
	if !isTerminal() {
		fmt.Println("   Run 'wade update' to upgrade.")
		return
	}

	fmt.Print("   Update now? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "y" || answer == "yes" {
		fmt.Println()
		if err := runUpdate(); err != nil {
			fmt.Println(err)
		}
	}
}

// updateCheckCachePath returns ~/.wade/.last-update-check
func updateCheckCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".wade", ".last-update-check")
}

func isTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
