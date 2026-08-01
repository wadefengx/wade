package node

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/wadefengx/wade/internal/config"
)

// PlatformFilename returns the download filename for a version on the current platform
func PlatformFilename(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	switch goos {
	case "darwin":
		if goarch == "arm64" {
			return fmt.Sprintf("node-%s-darwin-arm64.tar.gz", version)
		}
		return fmt.Sprintf("node-%s-darwin-x64.tar.gz", version)
	case "windows":
		return fmt.Sprintf("node-%s-win-x64.zip", version)
	default:
		return fmt.Sprintf("node-%s-linux-x64.tar.gz", version)
	}
}

// DownloadURL builds the full URL for downloading a Node version
func DownloadURL(version, mirror string) string {
	mirror = strings.TrimRight(mirror, "/")
	filename := PlatformFilename(version)
	return fmt.Sprintf("%s/%s/%s", mirror, version, filename)
}

// Install downloads and extracts a Node.js version
func Install(version, mirror string) error {
	// Check if already installed
	if IsInstalled(version) {
		return fmt.Errorf("version %s is already installed", version)
	}

	dir, err := config.WadeDir()
	if err != nil {
		return err
	}

	// Create temp download directory
	tmpDir, err := os.MkdirTemp("", "wade-node-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Download
	url := DownloadURL(version, mirror)
	archivePath := filepath.Join(tmpDir, "node-"+PlatformFilename(version))

	fmt.Printf("Downloading %s...\n", url)
	if err := downloadFile(url, archivePath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	// Extract
	destDir := filepath.Join(dir, "versions", version)
	fmt.Printf("Extracting to %s...\n", destDir)
	if err := extractArchive(archivePath, destDir); err != nil {
		return fmt.Errorf("extract: %w", err)
	}

	fmt.Printf("Node %s installed successfully\n", version)
	return nil
}

// Uninstall removes an installed Node version
func Uninstall(version string) error {
	if !IsInstalled(version) {
		return fmt.Errorf("version %s is not installed", version)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Check if it's the default
	if cfg.DefaultVersion == version {
		return fmt.Errorf("cannot uninstall default version %s — change default first", version)
	}

	dir, err := config.WadeDir()
	if err != nil {
		return err
	}

	versionsDir := filepath.Join(dir, "versions", version)
	return os.RemoveAll(versionsDir)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

// extractArchive dispatches to the right extractor based on file extension
func extractArchive(src, dest string) error {
	if strings.HasSuffix(src, ".zip") {
		return extractZip(src, dest)
	}
	return extractTarGz(src, dest)
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Remove dest if exists
	os.RemoveAll(dest)

	for _, f := range r.File {
		// Node zip archives are structured as `node-v18.20.0-win-x64/...`
		// Strip the first directory component
		parts := strings.SplitN(f.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		relPath := parts[1]
		if relPath == "" {
			continue
		}
		target := filepath.Join(dest, relPath)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		outFile, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(outFile, rc); err != nil {
			outFile.Close()
			rc.Close()
			return err
		}
		outFile.Close()
		rc.Close()
		os.Chmod(target, os.FileMode(f.Mode()))
	}

	return nil
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Remove dest if exists
	os.RemoveAll(dest)

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Node archives are structured as `node-v18.20.0-darwin-arm64/...`
		// Strip the first directory component
		parts := strings.SplitN(header.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		relPath := parts[1]

		target := filepath.Join(dest, relPath)

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			io.Copy(outFile, tr)
			outFile.Close()
			// Preserve permissions
			os.Chmod(target, os.FileMode(header.Mode))
		case tar.TypeSymlink:
			os.MkdirAll(filepath.Dir(target), 0755)
			os.Remove(target) // remove if exists
			os.Symlink(header.Linkname, target)
		}
	}

	return nil
}
