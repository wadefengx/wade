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

	// Download — dual-channel: API asset endpoint first (api.github.com +
	// release-assets CDN reachable from CN), github.com direct fallback.
	fmt.Printf("Downloading %s...\n", url)
	dlClient := newHTTPClient(60 * time.Second)

	resp, err := downloadAsset(dlClient, repo, latest, assetBase+archiveExtension)
	if err != nil {
		// Fallback: github.com direct
		resp, err = dlClient.Get(url)
	}
	if err != nil {
		return fmt.Errorf("download: %w\n       Tip: if the download hangs/times out (common in CN networks), set a proxy and retry:\n       set HTTP_PROXY=http://127.0.0.1:7890 && set HTTPS_PROXY=http://127.0.0.1:7890", err)
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

// getLatestVersion resolves the latest release tag.
// Dual-channel: GitHub API first (reachable from CN networks), HTTP redirect
// (https://github.com/REPO/releases/latest → 302 → .../releases/tag/vX.Y.Z)
// as fallback for networks where api.github.com is blocked/rate-limited.
func getLatestVersion(repo string) (string, error) {
	// Channel 1: API (api.github.com reachable from CN)
	if tag, err := getLatestVersionAPI(repo); err == nil {
		return tag, nil
	}
	// Channel 2: HTTP redirect (github.com, no API quota)
	url := fmt.Sprintf("https://github.com/%s/releases/latest", repo)
	client := newHTTPClient(10 * time.Second)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse // don't follow; grab Location
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

// getLatestVersionAPI resolves the latest tag via the GitHub REST API.
func getLatestVersionAPI(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	client := newHTTPClient(10 * time.Second)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "wade-update")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("api returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
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
	return text[tagStart : tagStart+tagEnd], nil
}

// downloadAsset downloads a release asset via the API endpoint
// api.github.com/repos/{repo}/releases/assets/{id}, which redirects to the
// release-assets CDN — reachable from CN networks (unlike github.com).
// assetName is e.g. "wade-windows-amd64.zip".
func downloadAsset(client *http.Client, repo, version, assetName string) (*http.Response, error) {
	apiClient := newHTTPClient(15 * time.Second)
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo), nil)
	req.Header.Set("User-Agent", "wade-update")
	relResp, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer relResp.Body.Close()
	if relResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("api HTTP %d", relResp.StatusCode)
	}

	body, err := io.ReadAll(relResp.Body)
	if err != nil {
		return nil, err
	}
	// Find asset id by name. GitHub asset JSON order:
	// {"url":..., "id":123, "node_id":..., "name":"wade-...zip", ...}
	// So the asset's own "id" appears BEFORE "name" — search backwards.
	assetID, err := assetIDByName(string(body), assetName)
	if err != nil {
		return nil, err
	}

	assetURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%s", repo, assetID)
	dlReq, _ := http.NewRequest("GET", assetURL, nil)
	dlReq.Header.Set("Accept", "application/octet-stream")
	dlReq.Header.Set("User-Agent", "wade-update")
	return client.Do(dlReq)
}

// assetIDByName extracts a release asset's id from the release JSON body.
// GitHub orders asset fields as {"url":..., "id":123, "node_id":...,
// "name":"...", ...} — the id sits BEFORE the name, so we search backwards
// from the name match (the nearest preceding "id": is the asset's own).
func assetIDByName(body, assetName string) (string, error) {
	needle := fmt.Sprintf(`"name":%q`, assetName)
	idx := strings.Index(body, needle)
	if idx == -1 {
		return "", fmt.Errorf("asset %s not found in release", assetName)
	}
	before := body[:idx]
	idStart := strings.LastIndex(before, `"id":`)
	if idStart == -1 {
		return "", fmt.Errorf("asset %s has no id", assetName)
	}
	idStart += len(`"id":`)
	idEnd := strings.IndexAny(before[idStart:], `,}`)
	if idEnd == -1 {
		return "", fmt.Errorf("asset %s id parse failed", assetName)
	}
	id := before[idStart : idStart+idEnd]
	if id == "" {
		return "", fmt.Errorf("asset %s id is empty", assetName)
	}
	return id, nil
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
	// Dual-channel checksum fetch: API asset endpoint first (CN-reachable),
	// github.com direct fallback. shaName is like "wade-windows-amd64.sha256".
	shaName := filepath.Base(checksumURL)

	var body []byte
	if resp, err := downloadAsset(newHTTPClient(30*time.Second), "wadefengx/wade", "", shaName); err == nil {
		b, rerr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if rerr == nil && resp.StatusCode == http.StatusOK {
			body = b
		}
	}
	if body == nil {
		resp, err := http.Get(checksumURL)
		if err != nil {
			return fmt.Errorf("download checksum: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("download checksum failed: HTTP %d", resp.StatusCode)
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read checksum: %w", err)
		}
	}

	fields := strings.Fields(string(body))
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
