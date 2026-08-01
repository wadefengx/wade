package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyChecksum
func TestVerifyChecksum(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "wade-darwin-arm64.tar.gz")
	content := []byte("release archive")
	if err := os.WriteFile(archive, content, 0600); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)

	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr string
	}{
		{
			name: "valid checksum",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "%x  wade-darwin-arm64.tar.gz\n", sum)
			},
		},
		{
			name: "mismatched checksum",
			handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, strings.Repeat("0", 64))
			},
			wantErr: "checksum mismatch",
		},
		{
			name: "missing checksum",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			},
			wantErr: "HTTP 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			err := verifyChecksum(archive, server.URL)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("verifyChecksum() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("verifyChecksum() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestExtractBinaryTarGz(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "wade-darwin-arm64.tar.gz")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(archive)
	tarWriter := tar.NewWriter(gz)
	content := []byte("fake wade binary")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "wade", Mode: 0755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(dir, "wade")
	if err := extractBinary(archivePath, destPath); err != nil {
		t.Fatalf("extractBinary() error = %v", err)
	}
	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("extracted content = %q, want %q", got, content)
	}
}
