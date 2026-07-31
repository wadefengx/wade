package golang

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/wadefengx/wade/internal/config"
)

// RemoteVersion from Go's JSON API
type RemoteVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []struct {
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Arch     string `json:"arch"`
		Kind     string `json:"kind"`
	} `json:"files"`
}

// FetchRemoteVersions fetches Go versions from the remote mirror
func FetchRemoteVersions(mirror string) ([]string, error) {
	url := strings.TrimRight(mirror, "/")
	if strings.Contains(url, "go.dev") {
		url = "https://go.dev/dl/?mode=json"
	} else {
		url += "/?mode=json"
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var versions []RemoteVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	result := make([]string, 0, len(versions))
	for _, v := range versions {
		if v.Stable {
			result = append(result, v.Version)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		vi, _ := semver.NewVersion(strings.TrimPrefix(result[i], "go"))
		vj, _ := semver.NewVersion(strings.TrimPrefix(result[j], "go"))
		if vi == nil || vj == nil {
			return result[i] > result[j]
		}
		return vi.GreaterThan(vj)
	})

	return result, nil
}

// PlatformFilename returns the download filename for the current platform
func PlatformFilename(version string) string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	return fmt.Sprintf("%s.%s-%s.tar.gz", version, goos, goarch)
}

// DownloadURL builds the full URL for downloading a Go version
func DownloadURL(version, mirror string) string {
	mirror = strings.TrimRight(mirror, "/")
	filename := PlatformFilename(version)
	return fmt.Sprintf("%s/%s", mirror, filename)
}

// IsInstalled checks if a version is installed
func IsInstalled(version string) bool {
	dir, err := config.WadeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(dir, "go", "versions", version))
	return err == nil
}

// InstalledVersions lists installed Go versions
func InstalledVersions() ([]string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return nil, err
	}
	versionsDir := filepath.Join(dir, "go", "versions")

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var versions []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "go") {
			versions = append(versions, e.Name())
		}
	}
	return versions, nil
}

// Install downloads and extracts a Go version
func Install(version, mirror string) error {
	if IsInstalled(version) {
		return fmt.Errorf("%s already installed", version)
	}

	dir, err := config.WadeDir()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "wade-go-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	url := DownloadURL(version, mirror)
	archivePath := filepath.Join(tmpDir, "go.tar.gz")

	fmt.Printf("📥 Downloading %s...\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	io.Copy(f, resp.Body)
	f.Close()

	// Go tar.gz has a top-level `go/` directory
	// Extract to temp, then move `go/` → versions/<version>/
	extractDir := filepath.Join(tmpDir, "extract")
	os.MkdirAll(extractDir, 0755)

	gzr, err := gzip.NewReader(openFile(archivePath))
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Strip the leading `go/` directory
		parts := strings.SplitN(header.Name, "/", 2)
		if len(parts) < 2 {
			continue
		}
		relPath := parts[1]
		if relPath == "" {
			continue
		}

		target := filepath.Join(extractDir, relPath)
		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(target, 0755)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(target), 0755)
			out, _ := os.Create(target)
			io.Copy(out, tr)
			out.Close()
			os.Chmod(target, os.FileMode(header.Mode))
		case tar.TypeSymlink:
			os.MkdirAll(filepath.Dir(target), 0755)
			os.Remove(target)
			os.Symlink(header.Linkname, target)
		}
	}

	destDir := filepath.Join(dir, "go", "versions", version)
	os.RemoveAll(destDir)
	os.MkdirAll(filepath.Dir(destDir), 0755)
	if err := os.Rename(extractDir, destDir); err != nil {
		return fmt.Errorf("move: %w", err)
	}

	fmt.Printf("✅ Go %s installed\n", version)
	return nil
}

func openFile(path string) *os.File {
	f, _ := os.Open(path)
	return f
}

// UseVersion activates a Go version via shim
func UseVersion(version string) error {
	if !IsInstalled(version) {
		return fmt.Errorf("%s not installed — run 'wade go install %s'", version, version)
	}

	dir, err := config.WadeDir()
	if err != nil {
		return err
	}

	shimDir := filepath.Join(dir, "shims")
	os.MkdirAll(shimDir, 0755)

	versionBin := filepath.Join(dir, "go", "versions", version, "bin")

	for _, name := range []string{"go", "gofmt"} {
		target := filepath.Join(versionBin, name)
		if _, err := os.Stat(target); err != nil {
			continue
		}
		shim := filepath.Join(shimDir, name)
		os.Remove(shim)
		os.Symlink(target, shim)
	}

	currentFile := filepath.Join(dir, "go", "current")
	os.MkdirAll(filepath.Dir(currentFile), 0755)
	os.WriteFile(currentFile, []byte(version+"\n"), 0644)
	return nil
}

// CurrentVersion returns the active Go version
func CurrentVersion() (string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, "go", "current"))
	if err != nil {
		return "", fmt.Errorf("no active Go version")
	}
	return strings.TrimSpace(string(data)), nil
}

// Uninstall removes a Go version
func Uninstall(version string) error {
	dir, err := config.WadeDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(dir, "go", "versions", version))
}
