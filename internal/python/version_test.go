package python

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	builds := []PythonBuild{
		{Version: "3.12.10"},
		{Version: "3.11.15"},
		{Version: "3.11.9"},
	}
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "partial", input: "3.11", want: "3.11.15"},
		{name: "exact", input: "3.12.10", want: "3.12.10"},
		{name: "missing", input: "3.10", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveVersion(tt.input, builds)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("resolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPlatformSuffix(t *testing.T) {
	tests := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "aarch64-apple-darwin"},
		{"darwin", "amd64", "x86_64-apple-darwin"},
		{"windows", "amd64", "x86_64-pc-windows-msvc"},
		{"linux", "amd64", "x86_64-unknown-linux-gnu"},
	}
	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.goarch, func(t *testing.T) {
			if got := platformSuffix(tt.goos, tt.goarch); got != tt.want {
				t.Errorf("platformSuffix(%q, %q) = %q, want %q", tt.goos, tt.goarch, got, tt.want)
			}
		})
	}
}

func TestExtractTarGzStripsPythonDirectory(t *testing.T) {
	temp := t.TempDir()
	archive := filepath.Join(temp, "python.tar.gz")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	writer := tar.NewWriter(gz)
	content := []byte("python")
	if err := writer.WriteHeader(&tar.Header{Name: "python/bin/python3", Mode: 0755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(temp, "dest")
	if err := extractTarGz(archive, dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "bin", "python3"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Errorf("extracted content = %q, want %q", got, content)
	}
	if _, err := os.Stat(filepath.Join(dest, "python")); !os.IsNotExist(err) {
		t.Errorf("top-level python directory was not stripped: %v", err)
	}
}
