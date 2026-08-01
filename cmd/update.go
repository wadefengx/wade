package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
		return runUpdate()
	},
}

// runUpdate performs the self-update: check latest, download, verify, replace.
func runUpdate() error {
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
	assetBase := fmt.Sprintf("wade-%s", platform)
	archiveExtension := ".tar.gz"
	if runtime.GOOS == "windows" {
		archiveExtension = ".zip"
	}

	releaseURL := fmt.Sprintf(
		"https://github.com/%s/releases/download/%s/",
		repo, latest,
	)
	url := releaseURL + assetBase + archiveExtension
	checksumURL := releaseURL + assetBase + ".sha256"

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
	tmpFile, err := os.CreateTemp("", "wade-update-*"+archiveExtension)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return err
	}
	if err := tmpFile.Chmod(0755); err != nil {
		return fmt.Errorf("set download permissions: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close download: %w", err)
	}

	if err := verifyChecksum(tmpFile.Name(), checksumURL); err != nil {
		return fmt.Errorf("verify downloaded update: %w", err)
	}

	// Find current binary location
	currentBin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find current binary: %w", err)
	}
	currentBin, err = filepath.EvalSymlinks(currentBin)
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}

	extractedBin, err := os.CreateTemp(filepath.Dir(currentBin), ".wade-update-*")
	if err != nil {
		return fmt.Errorf("create extracted binary: %w", err)
	}
	extractedPath := extractedBin.Name()
	if err := extractedBin.Close(); err != nil {
		os.Remove(extractedPath)
		return fmt.Errorf("close extracted binary: %w", err)
	}
	defer os.Remove(extractedPath)

	if err := extractBinary(tmpFile.Name(), extractedPath); err != nil {
		return fmt.Errorf("extract downloaded update: %w", err)
	}

	// Replace
	backup := currentBin + ".old"
	if err := os.Rename(currentBin, backup); err != nil {
		return fmt.Errorf("backup current binary: %w", err)
	}

	if err := os.Rename(extractedPath, currentBin); err != nil {
		// Restore backup
		os.Rename(backup, currentBin)
		return fmt.Errorf("install new binary: %w", err)
	}

	os.Remove(backup)
	fmt.Printf("Updated to %s\n", latest)
	return nil
}

// getLatestVersion resolves the latest release tag via HTTP redirect
// (https://github.com/REPO/releases/latest → 302 → .../releases/tag/vX.Y.Z).
// Uses the redirect Location instead of the API to avoid rate limits.
func getLatestVersion(repo string) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/releases/latest", repo)

	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // don't follow; grab Location
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	// Location looks like: https://github.com/wadefengx/wade/releases/tag/v0.3.3
	idx := strings.LastIndex(loc, "/tag/")
	if idx == -1 {
		return "", fmt.Errorf("could not find tag in redirect Location %q", loc)
	}
	tag := loc[idx+len("/tag/"):]
	if tag == "" {
		return "", fmt.Errorf("empty tag in redirect Location %q", loc)
	}
	return tag, nil
}

func extractBinary(archivePath, destPath string) error {
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractTarGzBinary(archivePath, destPath)
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZipBinary(archivePath, destPath)
	default:
		return fmt.Errorf("unsupported update archive format: %s", archivePath)
	}
}

func extractTarGzBinary(archivePath, destPath string) error {
	archive, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer archive.Close()

	gz, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("open gzip archive: %w", err)
	}
	defer gz.Close()

	tarReader := tar.NewReader(gz)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg || !isWadeBinary(header.Name) {
			continue
		}
		return writeExtractedBinary(destPath, tarReader)
	}
	return fmt.Errorf("wade binary not found in archive")
}

func extractZipBinary(archivePath, destPath string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}
	defer archive.Close()

	for _, file := range archive.File {
		if file.FileInfo().IsDir() || !isWadeBinary(file.Name) {
			continue
		}
		content, err := file.Open()
		if err != nil {
			return fmt.Errorf("open archived binary: %w", err)
		}
		err = writeExtractedBinary(destPath, content)
		content.Close()
		return err
	}
	return fmt.Errorf("wade binary not found in archive")
}

func isWadeBinary(name string) bool {
	base := filepath.Base(name)
	return base == "wade" || base == "wade.exe"
}

func writeExtractedBinary(destPath string, source io.Reader) error {
	dest, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open extracted binary: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return fmt.Errorf("write extracted binary: %w", err)
	}
	if err := dest.Chmod(0755); err != nil {
		return fmt.Errorf("set extracted binary permissions: %w", err)
	}
	return nil
}

func verifyChecksum(archivePath, checksumURL string) error {
	resp, err := http.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("download checksum: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download checksum failed: HTTP %d", resp.StatusCode)
	}

	checksumFile, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read checksum: %w", err)
	}
	fields := strings.Fields(string(checksumFile))
	if len(fields) == 0 {
		return fmt.Errorf("invalid checksum file")
	}

	archive, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open downloaded archive: %w", err)
	}
	defer archive.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, archive); err != nil {
		return fmt.Errorf("hash downloaded archive: %w", err)
	}

	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, fields[0]) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", fields[0], actual)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
}
