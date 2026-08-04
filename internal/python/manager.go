package python

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/wadefengx/wade/internal/config"
	"github.com/wadefengx/wade/internal/platform"
)

// PipMirror represents a pip registry mirror
type PipMirror struct {
	Name string
	URL  string
}

// PipPresets returns built-in pip mirrors
func PipPresets() []PipMirror {
	return []PipMirror{
		{Name: "pypi", URL: "https://pypi.org/simple/"},
		{Name: "tsinghua", URL: "https://pypi.tuna.tsinghua.edu.cn/simple/"},
		{Name: "aliyun", URL: "https://mirrors.aliyun.com/pypi/simple/"},
		{Name: "huawei", URL: "https://mirrors.huaweicloud.com/pypi/simple/"},
		{Name: "tencent", URL: "https://mirrors.tencent.com/pypi/simple/"},
		{Name: "ustc", URL: "https://pypi.mirrors.ustc.edu.cn/simple/"},
	}
}

// FindPipMirror finds a pip mirror by name
func FindPipMirror(name string) (*PipMirror, bool) {
	for _, m := range PipPresets() {
		if m.Name == name {
			return &m, true
		}
	}
	return nil, false
}

// UsePipMirror switches pip to a mirror
func UsePipMirror(name string) error {
	mirror, ok := FindPipMirror(name)
	if !ok {
		return fmt.Errorf("unknown pip mirror: %s", name)
	}

	// pip config set global.index-url <url>
	return exec.Command("pip", "config", "set", "global.index-url", mirror.URL).Run()
}

// GoProxy represents a Go proxy mirror
type GoProxy struct {
	Name string
	URL  string
}

// GoProxyPresets returns built-in Go proxies
func GoProxyPresets() []GoProxy {
	return []GoProxy{
		{Name: "official", URL: "https://proxy.golang.org,direct"},
		{Name: "goproxy.cn", URL: "https://goproxy.cn,direct"},
		{Name: "goproxy.io", URL: "https://goproxy.io,direct"},
	}
}

// FindGoProxy finds a Go proxy by name
func FindGoProxy(name string) (*GoProxy, bool) {
	for _, p := range GoProxyPresets() {
		if p.Name == name {
			return &p, true
		}
	}
	return nil, false
}

// UseGoProxy switches Go proxy
func UseGoProxy(name string) error {
	proxy, ok := FindGoProxy(name)
	if !ok {
		return fmt.Errorf("unknown Go proxy: %s", name)
	}
	return exec.Command("go", "env", "-w", fmt.Sprintf("GOPROXY=%s", proxy.URL)).Run()
}

// DetectSystemPython returns the detected Python version
func DetectSystemPython() []string {
	var versions []string
	for _, name := range []string{"python3", "python"} {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		out, err := exec.Command(path, "--version").Output()
		if err != nil {
			continue
		}
		ver := strings.TrimSpace(string(out))
		versions = append(versions, fmt.Sprintf("%s (system: %s)", path, ver))
	}
	return versions
}

// DetectSystemGo returns the detected Go version
func DetectSystemGo() string {
	path, err := exec.LookPath("go")
	if err != nil {
		return ""
	}
	out, err := exec.Command(path, "version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GoMirror represents a Go download mirror
type GoMirror struct {
	Name string
	URL  string
}

// GoMirrorPresets returns built-in Go download mirrors.
// npmmirror/aliyun Go mirrors are dead (404 as of 2026-08) — removed.
// google-cn verified working (dl/?mode=json + file downloads).
func GoMirrorPresets() []GoMirror {
	return []GoMirror{
		{Name: "official", URL: "https://go.dev/dl/"},
		{Name: "google-cn", URL: "https://golang.google.cn/dl/"},
	}
}

// FindGoMirror finds a Go mirror by name
func FindGoMirror(name string) (*GoMirror, bool) {
	for _, m := range GoMirrorPresets() {
		if m.Name == name {
			return &m, true
		}
	}
	return nil, false
}

// EnsureDir creates the go/python directory under ~/.wade/
func EnsureDir() error {
	dir, err := config.WadeDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(dir, "go", "versions"), 0755)
}

// PythonBuild represents a python-build-standalone release asset.
type PythonBuild struct {
	Version string
	Asset   string
	AssetID int64
}

const pythonBuildRepo = "astral-sh/python-build-standalone"

// PlatformSuffix returns the python-build-standalone platform suffix.
func PlatformSuffix() string {
	return platformSuffix(runtime.GOOS, runtime.GOARCH)
}

func platformSuffix(goos, goarch string) string {
	switch goos {
	case "darwin":
		if goarch == "arm64" {
			return "aarch64-apple-darwin"
		}
		if goarch == "amd64" {
			return "x86_64-apple-darwin"
		}
	case "windows":
		if goarch == "amd64" {
			return "x86_64-pc-windows-msvc"
		}
	case "linux":
		if goarch == "amd64" {
			return "x86_64-unknown-linux-gnu"
		}
	}
	return ""
}

// FetchRemoteVersions returns install-only builds for the current platform.
func FetchRemoteVersions() ([]PythonBuild, error) {
	suffix := PlatformSuffix()
	if suffix == "" {
		return nil, fmt.Errorf("unsupported Python platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("GET", "https://api.github.com/repos/"+pythonBuildRepo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "wade")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch Python releases from GitHub API: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch Python releases: GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release struct {
		Assets []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parse Python release: %w", err)
	}

	builds := make([]PythonBuild, 0)
	for _, asset := range release.Assets {
		version, ok := pythonAssetVersion(asset.Name, suffix)
		if ok {
			builds = append(builds, PythonBuild{Version: version, Asset: asset.Name, AssetID: asset.ID})
		}
	}
	sort.Slice(builds, func(i, j int) bool {
		left, _ := semver.NewVersion(builds[i].Version)
		right, _ := semver.NewVersion(builds[j].Version)
		if left == nil || right == nil {
			return builds[i].Version > builds[j].Version
		}
		return left.GreaterThan(right)
	})
	return builds, nil
}

func pythonAssetVersion(name, suffix string) (string, bool) {
	const prefix = "cpython-"
	const marker = "-install_only.tar.gz"
	if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, marker) ||
		!strings.Contains(name, "-"+suffix+marker) {
		return "", false
	}
	buildVersion := strings.TrimPrefix(name, prefix)
	buildVersion = strings.TrimSuffix(buildVersion, "-"+suffix+marker)
	version, _, found := strings.Cut(buildVersion, "+")
	if !found || version == "" {
		return "", false
	}
	if _, err := semver.NewVersion(version); err != nil {
		return "", false
	}
	return version, true
}

// ResolveVersion resolves a partial Python version to the latest matching build.
func ResolveVersion(input string) (string, error) {
	builds, err := FetchRemoteVersions()
	if err != nil {
		return "", err
	}
	return resolveVersion(input, builds)
}

func resolveVersion(input string, builds []PythonBuild) (string, error) {
	input = strings.TrimPrefix(strings.TrimSpace(input), "v")
	for _, build := range builds {
		if build.Version == input {
			return build.Version, nil
		}
	}
	prefix := input + "."
	for _, build := range builds {
		if strings.HasPrefix(build.Version, prefix) {
			return build.Version, nil
		}
	}
	return "", fmt.Errorf("no Python version matches %q — try 'wade python ls-remote'", input)
}

func isInstalled(version string) bool {
	dir, err := config.WadeDir()
	if err != nil {
		return false
	}
	info, err := os.Stat(filepath.Join(dir, "python", "versions", version))
	return err == nil && info.IsDir()
}

// InstalledVersions returns wade-managed Python versions.
func InstalledVersions() ([]string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(dir, "python", "versions"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		left, _ := semver.NewVersion(versions[i])
		right, _ := semver.NewVersion(versions[j])
		if left == nil || right == nil {
			return versions[i] > versions[j]
		}
		return left.GreaterThan(right)
	})
	return versions, nil
}

// Install downloads, extracts, and activates a Python build.
func Install(version string) error {
	builds, err := FetchRemoteVersions()
	if err != nil {
		return err
	}
	resolved, err := resolveVersion(version, builds)
	if err != nil {
		return err
	}
	if isInstalled(resolved) {
		return fmt.Errorf("Python %s is already installed", resolved)
	}

	var build PythonBuild
	for _, candidate := range builds {
		if candidate.Version == resolved {
			build = candidate
			break
		}
	}
	if build.AssetID == 0 {
		return fmt.Errorf("Python build %s has no download asset", resolved)
	}

	tmpDir, err := os.MkdirTemp("", "wade-python-*")
	if err != nil {
		return fmt.Errorf("create temporary directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)
	archive := filepath.Join(tmpDir, build.Asset)
	if err := downloadBuild(build, archive); err != nil {
		return err
	}

	dir, err := config.WadeDir()
	if err != nil {
		return err
	}
	dest := filepath.Join(dir, "python", "versions", resolved)
	if err := extractTarGz(archive, dest); err != nil {
		return fmt.Errorf("extract Python %s: %w", resolved, err)
	}
	if err := UseVersion(resolved); err != nil {
		return fmt.Errorf("activate Python %s: %w", resolved, err)
	}
	fmt.Printf("Python %s installed successfully\n", resolved)
	return nil
}

func downloadBuild(build PythonBuild, dest string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", pythonBuildRepo, build.AssetID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "wade")
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		url := fmt.Sprintf("https://github.com/%s/releases/latest/download/%s", pythonBuildRepo, build.Asset)
		resp, err = client.Get(url)
		if err != nil {
			return fmt.Errorf("download Python %s: %w", build.Version, err)
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download Python %s: HTTP %d", build.Version, resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("write Python archive: %w", err)
	}
	return nil
}

// extractTarGz extracts an archive while stripping its top-level python/ directory.
func extractTarGz(archive, dest string) error {
	file, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	if err := os.RemoveAll(dest); err != nil {
		return err
	}

	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		rel, ok := strings.CutPrefix(filepath.ToSlash(header.Name), "python/")
		if !ok || rel == "" || filepath.IsAbs(rel) || strings.HasPrefix(filepath.Clean(rel), "..") {
			continue
		}
		target := filepath.Join(dest, filepath.FromSlash(rel))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, reader)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
				return err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}

// UseVersion activates a managed Python build through wade shims.
func UseVersion(version string) error {
	if !isInstalled(version) {
		return fmt.Errorf("Python %s is not installed — run 'wade python install %s'", version, version)
	}
	dir, err := config.WadeDir()
	if err != nil {
		return err
	}
	shimDir := filepath.Join(dir, "shims")
	if err := os.MkdirAll(shimDir, 0755); err != nil {
		return err
	}

	names := []string{"python3", "pip3"}
	if runtime.GOOS == "windows" {
		names = []string{"python.exe", "pip.exe"}
	}
	for _, name := range names {
		target := filepath.Join(dir, "python", "versions", version, "bin", name)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		shim := filepath.Join(shimDir, name)
		if err := os.Remove(shim); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove existing shim for %s: %w", name, err)
		}
		if err := createPythonShim(target, shim); err != nil {
			return fmt.Errorf("create shim for %s: %w", name, err)
		}
	}
	return os.WriteFile(filepath.Join(dir, "python", "current"), []byte(version+"\n"), 0644)
}

func createPythonShim(target, shim string) error {
	if runtime.GOOS == "windows" {
		return os.Link(target, shim)
	}
	return platform.Symlink(target, shim)
}

// CurrentVersion returns the active managed Python version.
func CurrentVersion() (string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, "python", "current"))
	if err != nil {
		return "", fmt.Errorf("no active Python version")
	}
	return strings.TrimSpace(string(data)), nil
}

// Uninstall removes a managed Python version.
func Uninstall(version string) error {
	dir, err := config.WadeDir()
	if err != nil {
		return err
	}
	if !isInstalled(version) {
		return fmt.Errorf("Python %s is not installed", version)
	}
	return os.RemoveAll(filepath.Join(dir, "python", "versions", version))
}
