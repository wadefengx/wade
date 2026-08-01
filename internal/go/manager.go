package golang

import (
	"archive/tar"
	"archive/zip"
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
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/platform"
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

// ResolveVersion resolves a partial version (e.g. "1.23") to the latest
// matching release (e.g. "go1.23.12") by querying the mirror's JSON API.
// Exact versions (e.g. "go1.23.12" or "1.23.12") pass through.
func ResolveVersion(input, mirror string) (string, error) {
	ver := input
	if !strings.HasPrefix(ver, "go") {
		ver = "go" + ver
	}

	// Exact version (go1.23.12) — no lookup needed
	if parts := strings.SplitN(strings.TrimPrefix(ver, "go"), ".", 3); len(parts) == 3 {
		return ver, nil
	}

	versions, err := FetchRemoteVersions(mirror)
	if err != nil {
		return "", fmt.Errorf("failed to fetch Go version list: %w", err)
	}
	prefix := ver + "."
	for _, v := range versions {
		if strings.HasPrefix(v, prefix) {
			return v, nil
		}
	}
	return "", fmt.Errorf("no Go version matches %q — try 'wade go ls-remote'", strings.TrimPrefix(input, "go"))
}

// FetchRemoteVersions fetches Go versions from the remote mirror.
// Uses ?mode=json&include=all: the default API only returns the two latest
// minor releases (go1.26, go1.25), which breaks `install 1.23` — older
// version files still exist on the mirrors, they're just not listed.
func FetchRemoteVersions(mirror string) ([]string, error) {
	url := strings.TrimRight(mirror, "/") + "/?mode=json&include=all"

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
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
	if goos == "windows" {
		return fmt.Sprintf("%s.%s-%s.zip", version, goos, goarch)
	}
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
	archivePath := filepath.Join(tmpDir, PlatformFilename(version))

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

	// Go archives (tar.gz on unix, zip on windows) have a top-level `go/` directory
	// Extract to temp, then move `go/` → versions/<version>/
	extractDir := filepath.Join(tmpDir, "extract")
	os.MkdirAll(extractDir, 0755)

	if strings.HasSuffix(archivePath, ".zip") {
		if err := extractGoZip(archivePath, extractDir); err != nil {
			return err
		}
	} else {
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
	}

	destDir := filepath.Join(dir, "go", "versions", version)
	os.RemoveAll(destDir)
	os.MkdirAll(filepath.Dir(destDir), 0755)
	if err := os.Rename(extractDir, destDir); err != nil {
		return fmt.Errorf("move: %w", err)
	}

	fmt.Printf("✅ Go %s installed\n", version)

	// Auto-activate like `wade node install` does — user expects `go` to
	// work right after install, not after a separate `wade go use`.
	if err := UseVersion(version); err != nil {
		return fmt.Errorf("installed but activation failed: %w", err)
	}
	return nil
}

// extractGoZip extracts a Go zip archive (top-level `go/` dir) into dest
func extractGoZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
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
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			out.Close()
			rc.Close()
			return err
		}
		out.Close()
		rc.Close()
		os.Chmod(target, os.FileMode(f.Mode()))
	}
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
	versionRoot := filepath.Join(dir, "go", "versions", version)

	names := []string{"go", "gofmt"}
	if runtime.GOOS == "windows" {
		names = []string{"go.exe", "gofmt.exe"}
	}
	for _, name := range names {
		target := filepath.Join(versionBin, name)
		if _, err := os.Stat(target); err != nil {
			continue
		}

		if runtime.GOOS == "windows" {
			// Go 1.21+ binaries are trimmed: they infer GOROOT from their OWN
			// executable path. A hardlink shim (shims/go.exe) makes
			// os.Executable() return the shim path → GOROOT resolution fails
			// ("go binary is trimmed and GOROOT is not set").
			// Fix: a .cmd wrapper that calls the REAL binary path and sets
			// GOROOT explicitly. PATHEXT finds go.cmd when typing `go`.
			shimName := strings.TrimSuffix(name, filepath.Ext(name)) + ".cmd"
			shim := filepath.Join(shimDir, shimName)
			// remove stale hardlink shims (go.exe) so .cmd wins PATHEXT
			for _, stale := range []string{filepath.Join(shimDir, name), filepath.Join(shimDir, strings.TrimSuffix(name, filepath.Ext(name)))} {
				os.Remove(stale)
			}
			content := fmt.Sprintf(
				"@echo off\r\nset GOROOT=%s\r\n\"%s\" %%*\r\n",
				versionRoot, target,
			)
			if err := os.WriteFile(shim, []byte(content), 0755); err != nil {
				return fmt.Errorf("create shim for %s: %w", name, err)
			}
			continue
		}

		// Unix: symlink directly to bin/go — symlinks are resolved by
		// os.Executable(), so GOROOT inference works.
		shim := filepath.Join(shimDir, name)
		if err := os.Remove(shim); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove existing shim for %s: %w", name, err)
		}
		if err := platform.Symlink(target, shim); err != nil {
			return fmt.Errorf("create shim for %s: %w", name, err)
		}
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
