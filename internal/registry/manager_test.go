package registry

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTestMirrors(t *testing.T) {
	// Create a test server that responds quickly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mirrors := []Registry{
		{Name: "fast", URL: server.URL, IsBuiltIn: true},
	}

	results := TestMirrors(mirrors)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error != "" {
		t.Errorf("unexpected error: %s", results[0].Error)
	}
	if results[0].Latency <= 0 {
		t.Errorf("expected positive latency, got %s", results[0].Latency)
	}
	if results[0].Name != "fast" {
		t.Errorf("expected name 'fast', got %s", results[0].Name)
	}
}

func TestTestMirrors_Timeout(t *testing.T) {
	// Create a server that accepts the connection but never responds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close without response — simulates a broken mirror
		panic("simulated hang")
	}))
	defer server.Close()

	mirrors := []Registry{
		{Name: "broken", URL: server.URL, IsBuiltIn: true},
	}

	results := TestMirrors(mirrors)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == "" {
		t.Error("expected error for broken mirror, got none")
	}
}

func TestFind_PresetPriority(t *testing.T) {
	// Custom registry with same name as a preset should not shadow it
	custom := []Registry{{Name: "npm", URL: "https://evil.com/"}}
	r, ok := Find("npm", custom)
	if !ok || r == nil {
		t.Fatal("expected to find npm")
	}
	if r.URL != "https://registry.npmjs.org/" {
		t.Errorf("preset should take priority over custom, got URL %s", r.URL)
	}
}

func TestAll_NoDuplicates(t *testing.T) {
	// Adding a custom registry with a different name should work
	custom := []Registry{{Name: "myreg", URL: "https://myreg.com/"}}
	all := All(custom)

	names := make(map[string]int)
	for _, r := range all {
		names[r.Name]++
	}
	for name, count := range names {
		if count > 1 {
			t.Errorf("duplicate registry name: %q appears %d times", name, count)
		}
	}
}
