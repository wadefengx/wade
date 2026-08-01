package golang

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestResolveVersionExact: exact versions pass through without network.
func TestResolveVersionExact(t *testing.T) {
	cases := []struct{ in, want string }{
		{"go1.23.12", "go1.23.12"},
		{"1.23.12", "go1.23.12"},
	}
	for _, c := range cases {
		got, err := ResolveVersion(c.in, "https://go.dev/dl/")
		if err != nil {
			t.Errorf("ResolveVersion(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveVersion(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestResolveVersionPartial: "1.23" resolves to the latest 1.23.x via the
// mirror's JSON API (httptest server, no real network).
func TestResolveVersionPartial(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := `[
{"version":"go1.26.5","stable":true},
{"version":"go1.25.12","stable":true},
{"version":"go1.23.12","stable":true},
{"version":"go1.23.11","stable":true},
{"version":"go1.22.9","stable":true},
{"version":"go1.21.13","stable":true}
]`
		w.Write([]byte(body))
	}))
	defer srv.Close()

	got, err := ResolveVersion("1.23", srv.URL)
	if err != nil {
		t.Fatalf("ResolveVersion(1.23): %v", err)
	}
	if got != "go1.23.12" {
		t.Errorf("ResolveVersion(1.23) = %q, want go1.23.12", got)
	}
}

// TestResolveVersionNoMatch: unknown partial returns an error.
func TestResolveVersionNoMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"version":"go1.26.5","stable":true}]`))
	}))
	defer srv.Close()

	if _, err := ResolveVersion("9.99", srv.URL); err == nil {
		t.Error("expected error for non-existent version")
	}
}

// TestResolveVersionLiveGoogleCN hits the real google-cn JSON API to verify
// include=all exposes old minors (1.23) — the default API only lists the two
// latest minors, which would break `wade go install 1.23`.
func TestResolveVersionLiveGoogleCN(t *testing.T) {
	got, err := ResolveVersion("1.23", "https://golang.google.cn/dl/")
	if err != nil {
		t.Fatalf("ResolveVersion(1.23): %v", err)
	}
	t.Logf("1.23 → %s", got)
	if got != "go1.23.12" {
		t.Errorf("got %s, want go1.23.12", got)
	}
}
