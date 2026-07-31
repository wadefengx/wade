package golang

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestFetchRemoteVersions(t *testing.T) {
	validServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.URL.Query().Get("mode") != "json" {
			t.Errorf("unexpected request: %s", r.URL.String())
		}
		_, _ = w.Write([]byte(`[
			{"version":"go1.22.9","stable":true},
			{"version":"go1.24.0","stable":false},
			{"version":"go1.23.4","stable":true}
		]`))
	}))
	defer validServer.Close()

	invalidJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer invalidJSONServer.Close()

	tests := []struct {
		name    string
		mirror  string
		want    []string
		wantErr bool
	}{
		{
			name:   "filters unstable versions and sorts stable versions descending",
			mirror: validServer.URL,
			want:   []string{"go1.23.4", "go1.22.9"},
		},
		{
			name:    "returns error for unreachable mirror",
			mirror:  "http://127.0.0.1:1",
			wantErr: true,
		},
		{
			name:    "returns error for invalid JSON",
			mirror:  invalidJSONServer.URL,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FetchRemoteVersions(tt.mirror)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FetchRemoteVersions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchRemoteVersions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformFilename(t *testing.T) {
	version := "go1.23.4"
	want := version + "." + runtime.GOOS + "-" + runtime.GOARCH + ".tar.gz"
	if got := PlatformFilename(version); got != want {
		t.Errorf("PlatformFilename() = %q, want %q", got, want)
	}
}

func TestDownloadURL(t *testing.T) {
	const version = "go1.23.4"
	mirror := "https://example.com/go///"
	want := "https://example.com/go/" + PlatformFilename(version)
	if got := DownloadURL(version, mirror); got != want {
		t.Errorf("DownloadURL() = %q, want %q", got, want)
	}
}

func TestIsInstalledAndInstalledVersions(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	versionsDir := filepath.Join(tmpHome, ".wade", "go", "versions")
	for _, name := range []string{"go1.22.9", "go1.23.4", "node-v20", "go-file"} {
		path := filepath.Join(versionsDir, name)
		if name == "go-file" {
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, nil, 0644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{name: "installed version", version: "go1.23.4", want: true},
		{name: "missing version", version: "go1.24.0", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInstalled(tt.version); got != tt.want {
				t.Errorf("IsInstalled(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}

	got, err := InstalledVersions()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"go1.22.9", "go1.23.4"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("InstalledVersions() = %v, want %v", got, want)
	}
}

func TestInstalledVersionsMissingDirectory(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	got, err := InstalledVersions()
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("InstalledVersions() = %v, want nil", got)
	}
}

func TestUseVersion(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	const version = "go1.23.4"
	binDir := filepath.Join(tmpHome, ".wade", "go", "versions", version, "bin")
	for _, name := range []string{"go", "gofmt"} {
		if err := os.MkdirAll(binDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(binDir, name), nil, 0755); err != nil {
			t.Fatal(err)
		}
	}

	if err := UseVersion(version); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"go", "gofmt"} {
		shim := filepath.Join(tmpHome, ".wade", "shims", name)
		info, err := os.Lstat(shim)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("%s is not a symlink", shim)
		}
		target, err := os.Readlink(shim)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(binDir, name)
		if target != want {
			t.Errorf("symlink target = %q, want %q", target, want)
		}
	}

	data, err := os.ReadFile(filepath.Join(tmpHome, ".wade", "go", "current"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != version+"\n" {
		t.Errorf("current file = %q, want %q", data, version+"\n")
	}

	if err := UseVersion("go1.24.0"); err == nil {
		t.Error("UseVersion() missing version error = nil, want non-nil")
	}
}

func TestCurrentVersion(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	const version = "go1.23.4"
	binDir := filepath.Join(tmpHome, ".wade", "go", "versions", version, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "go"), nil, 0755); err != nil {
		t.Fatal(err)
	}
	if err := UseVersion(version); err != nil {
		t.Fatal(err)
	}

	got, err := CurrentVersion()
	if err != nil {
		t.Fatal(err)
	}
	if got != version {
		t.Errorf("CurrentVersion() = %q, want %q", got, version)
	}
}

func TestCurrentVersionMissingFile(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	if _, err := CurrentVersion(); err == nil {
		t.Error("CurrentVersion() missing file error = nil, want non-nil")
	}
}

func TestUninstall(t *testing.T) {
	tmpHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", oldHome)

	const version = "go1.23.4"
	versionDir := filepath.Join(tmpHome, ".wade", "go", "versions", version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := Uninstall(version); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(versionDir); !os.IsNotExist(err) {
		t.Errorf("version directory still exists or stat failed: %v", err)
	}
}
