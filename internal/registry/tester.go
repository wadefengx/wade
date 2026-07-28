package registry

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// TestResult holds the latency test result for one registry
type TestResult struct {
	Name    string
	URL     string
	Latency time.Duration
	Error   string
}

// TestMirrors measures latency for a list of mirrors (no presets added)
func TestMirrors(mirrors []Registry) []TestResult {
	results := make([]TestResult, len(mirrors))

	var wg sync.WaitGroup
	for i, reg := range mirrors {
		wg.Add(1)
		go func(idx int, r Registry) {
			defer wg.Done()
			results[idx] = testOne(r)
		}(i, reg)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].Error != "" && results[j].Error == "" {
			return false
		}
		if results[i].Error == "" && results[j].Error != "" {
			return true
		}
		return results[i].Latency < results[j].Latency
	})

	return results
}

// Test measures latency for all registries (presets + custom)
// Deprecated: Use TestMirrors for pure mirror testing
func Test(custom []Registry) []TestResult {
	all := All(custom)
	results := make([]TestResult, len(all))

	var wg sync.WaitGroup
	for i, reg := range all {
		wg.Add(1)
		go func(idx int, r Registry) {
			defer wg.Done()
			results[idx] = testOne(r)
		}(i, reg)
	}
	wg.Wait()

	// Sort by latency (fastest first), errors last
	sort.Slice(results, func(i, j int) bool {
		if results[i].Error != "" && results[j].Error == "" {
			return false
		}
		if results[i].Error == "" && results[j].Error != "" {
			return true
		}
		return results[i].Latency < results[j].Latency
	})

	return results
}

func testOne(reg Registry) TestResult {
	tr := TestResult{Name: reg.Name, URL: reg.URL}

	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()

	resp, err := client.Head(reg.URL)
	if err != nil {
		tr.Error = fmt.Sprintf("failed: %v", err)
		return tr
	}
	resp.Body.Close()

	tr.Latency = time.Since(start)
	return tr
}
