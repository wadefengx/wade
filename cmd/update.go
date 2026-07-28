package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var version string

// SetVersion sets the build version (called from main.go)
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("wade %s\n", version)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update wade to the latest version",
	Long: `Check for a newer version and self-update.

Uses GitHub Releases to find the latest version and downloads the
appropriate binary for the current platform.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		const repo = "wadefengx/wade"

		// Check latest version from GitHub API
		latest, err := getLatestVersion(repo)
		if err != nil {
			return fmt.Errorf("check latest version: %w", err)
		}

		if latest == version {
			fmt.Printf("Already up-to-date (wade %s)\n", version)
			return nil
		}

		fmt.Printf("Updating wade %s → %s\n", version, latest)

		// Build download URL
		platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
		if runtime.GOOS == "windows" {
			platform += ".zip"
		} else {
			platform += ".tar.gz"
		}

		url := fmt.Sprintf(
			"https://github.com/%s/releases/download/%s/wade-%s",
			repo, latest, platform,
		)

		// Download
		fmt.Printf("Downloading %s...\n", url)
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("download: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
		}

		// Write to temp file
		tmpFile, err := os.CreateTemp("", "wade-update-*")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())

		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			return err
		}
		tmpFile.Chmod(0755)
		tmpFile.Close()

		// Find current binary location
		currentBin, err := os.Executable()
		if err != nil {
			return fmt.Errorf("find current binary: %w", err)
		}
		currentBin, err = filepath.EvalSymlinks(currentBin)
		if err != nil {
			return fmt.Errorf("resolve binary path: %w", err)
		}

		// Replace
		backup := currentBin + ".old"
		if err := os.Rename(currentBin, backup); err != nil {
			return fmt.Errorf("backup current binary: %w", err)
		}

		if err := os.Rename(tmpFile.Name(), currentBin); err != nil {
			// Restore backup
			os.Rename(backup, currentBin)
			return fmt.Errorf("install new binary: %w", err)
		}

		os.Remove(backup)
		fmt.Printf("Updated to %s\n", latest)
		return nil
	},
}

func getLatestVersion(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Simple JSON tag_name extraction (no dependency needed)
	text := string(body)
	tagStart := strings.Index(text, `"tag_name":"`)
	if tagStart == -1 {
		return "", fmt.Errorf("could not find tag_name in response")
	}
	tagStart += len(`"tag_name":"`)
	tagEnd := strings.Index(text[tagStart:], `"`)
	if tagEnd == -1 {
		return "", fmt.Errorf("could not parse tag_name")
	}
	tag := text[tagStart : tagStart+tagEnd]

	return tag, nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
}
