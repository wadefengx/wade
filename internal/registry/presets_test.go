package registry

import (
	"testing"
)

func TestPresets(t *testing.T) {
	ps := Presets()
	if len(ps) != 5 {
		t.Fatalf("expected 5 presets, got %d", len(ps))
	}
	names := make(map[string]bool)
	for _, p := range ps {
		if !p.IsBuiltIn {
			t.Errorf("preset %q should be built-in", p.Name)
		}
		if names[p.Name] {
			t.Errorf("duplicate preset name: %q", p.Name)
		}
		names[p.Name] = true
		if p.URL == "" {
			t.Errorf("preset %q has empty URL", p.Name)
		}
	}
	// spot check
	if !names["npm"] || !names["taobao"] || !names["tencent"] {
		t.Error("missing expected presets (npm, taobao, tencent)")
	}
}

func TestAll(t *testing.T) {
	custom := []Registry{{Name: "myreg", URL: "https://myreg.com/"}}
	all := All(custom)
	if len(all) != 6 { // 5 presets + 1 custom
		t.Fatalf("expected 6 registries, got %d", len(all))
	}
}

func TestFind(t *testing.T) {
	custom := []Registry{{Name: "myreg", URL: "https://myreg.com/"}}

	// find preset
	r, ok := Find("npm", custom)
	if !ok || r == nil {
		t.Fatal("expected to find 'npm'")
	}
	if r.URL != "https://registry.npmjs.org/" {
		t.Errorf("expected npm URL, got %s", r.URL)
	}

	// find custom
	r, ok = Find("myreg", custom)
	if !ok || r == nil {
		t.Fatal("expected to find 'myreg'")
	}
	if r.URL != "https://myreg.com/" {
		t.Errorf("expected myreg URL, got %s", r.URL)
	}

	// not found
	r, ok = Find("nonexistent", custom)
	if ok || r != nil {
		t.Error("expected not found for nonexistent registry")
	}
}

func TestIsBuiltIn(t *testing.T) {
	if !IsBuiltIn("npm") {
		t.Error("npm should be built-in")
	}
	if !IsBuiltIn("taobao") {
		t.Error("taobao should be built-in")
	}
	if IsBuiltIn("myreg") {
		t.Error("myreg should NOT be built-in")
	}
}
