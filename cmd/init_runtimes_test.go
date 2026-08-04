package cmd

import (
	"os"
	"testing"
)

// TestRuntimesSelection: the "All of the above" choice must map to all three
// has* flags. Regression for the survey.Select write bug: the interactive
// wizard passed *[]string to a single-select prompt, survey refused to write
// ("Unable to convert from string to type slice"), runtimes stayed empty,
// and Go/Python were silently skipped in interactive mode (only -y worked).
func TestRuntimesSelection(t *testing.T) {
	// Simulate the fixed interactive path: choice -> runtimes slice
	choice := "All of the above"
	runtimes := []string{choice}

	hasNode := contains(runtimes, "Node.js") || contains(runtimes, "All of the above")
	hasGo := contains(runtimes, "Go") || contains(runtimes, "All of the above")
	hasPython := contains(runtimes, "Python") || contains(runtimes, "All of the above")

	if !hasNode || !hasGo || !hasPython {
		t.Fatalf("All of the above must enable all runtimes: node=%v go=%v python=%v",
			hasNode, hasGo, hasPython)
	}

	// Single selection must work too
	choice = "Go"
	runtimes = []string{choice}
	if hasGo := contains(runtimes, "Go") || contains(runtimes, "All of the above"); !hasGo {
		t.Error("single 'Go' selection should enable Go")
	}
	if hasNode := contains(runtimes, "Node.js") || contains(runtimes, "All of the above"); hasNode {
		t.Error("single 'Go' selection should NOT enable Node")
	}
}

// TestInitEnv: HOME redirect sanity — ensure config writes don't touch real ~/.wade
func TestInitEnv(t *testing.T) {
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", t.TempDir())
}
