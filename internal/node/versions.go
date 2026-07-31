package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/wadefengx/wade/internal/config"
)

var ErrProjectVersionNotFound = errors.New(".wade-version file not found")

// Version represents a parsed Node version
type Version struct {
	Major uint64
	Minor uint64
	Patch uint64
	Raw   string // e.g. "v18.20.0"
}

// ParseVersion parses a raw version string like "18", "18.20", "18.20.0"
func ParseVersion(input string) (string, error) {
	input = strings.TrimPrefix(input, "v")
	input = strings.TrimSpace(input)

	// Handle "lts" and "latest" — return as-is for remote resolution
	if input == "lts" || input == "latest" {
		return input, nil
	}

	parts := strings.Split(input, ".")
	if len(parts) > 3 {
		return "", fmt.Errorf("invalid version format: %q — use '18', '18.20', or '18.20.0'", input)
	}

	for _, p := range parts {
		if _, err := strconv.Atoi(p); err != nil {
			return "", fmt.Errorf("invalid version format: %q", input)
		}
	}

	return input, nil
}

// ResolveVersion resolves a partial version to a full version by querying the remote mirror
func ResolveVersion(input string, mirror string) (string, error) {
	input, err := ParseVersion(input)
	if err != nil {
		return "", err
	}

	// If input is already exact (e.g., 18.20.0), just return it
	if strings.Count(input, ".") == 2 {
		return "v" + input, nil
	}

	// Fetch remote versions
	versions, err := FetchRemoteVersions(mirror)
	if err != nil {
		return "", fmt.Errorf("failed to fetch version list: %w", err)
	}

	// Build semver constraint
	var constraintStr string
	switch input {
	case "latest", "lts":
		// Just return the latest available
		if len(versions) > 0 {
			return versions[0], nil
		}
		return "", fmt.Errorf("no versions found on mirror")
	default:
		switch strings.Count(input, ".") {
		case 0:
			constraintStr = fmt.Sprintf(">= %s.0.0, < %d.0.0", input, mustAtoi(input)+1)
		case 1:
			constraintStr = fmt.Sprintf(">= %s.0, < %s.%d", input, strings.Split(input, ".")[0], mustAtoi(strings.Split(input, ".")[1])+1)
		default:
			constraintStr = fmt.Sprintf("= %s", input)
		}
	}

	constraint, err := semver.NewConstraint(constraintStr)
	if err != nil {
		return "", fmt.Errorf("invalid version constraint: %w", err)
	}

	// Find best matching version
	var bestMatch *semver.Version
	for _, raw := range versions {
		v, err := semver.NewVersion(strings.TrimPrefix(raw, "v"))
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			if bestMatch == nil || v.GreaterThan(bestMatch) {
				bestMatch = v
			}
		}
	}

	if bestMatch == nil {
		return "", fmt.Errorf("no matching version found for %q", input)
	}

	return "v" + bestMatch.String(), nil
}

// FetchRemoteVersions fetches available Node versions from the mirror index
func FetchRemoteVersions(mirror string) ([]string, error) {
	url := strings.TrimRight(mirror, "/") + "/index.json"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entries []struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("parse index: %w", err)
	}

	versions := make([]string, len(entries))
	for i, e := range entries {
		versions[i] = e.Version
	}

	// Sort descending (newest first)
	sort.Slice(versions, func(i, j int) bool {
		vi, err1 := semver.NewVersion(strings.TrimPrefix(versions[i], "v"))
		vj, err2 := semver.NewVersion(strings.TrimPrefix(versions[j], "v"))
		if err1 != nil || err2 != nil {
			return versions[i] > versions[j]
		}
		return vi.GreaterThan(vj)
	})

	return versions, nil
}

// InstalledVersions lists locally installed Node versions
func InstalledVersions() ([]string, error) {
	dir, err := config.WadeDir()
	if err != nil {
		return nil, err
	}
	versionsDir := filepath.Join(dir, "versions")

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var versions []string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "v") {
			versions = append(versions, e.Name())
		}
	}

	sort.Slice(versions, func(i, j int) bool {
		vi, _ := semver.NewVersion(strings.TrimPrefix(versions[i], "v"))
		vj, _ := semver.NewVersion(strings.TrimPrefix(versions[j], "v"))
		if vi == nil || vj == nil {
			return versions[i] > versions[j]
		}
		return vi.GreaterThan(vj)
	})

	return versions, nil
}

// IsInstalled checks if a version is installed locally
func IsInstalled(version string) bool {
	versions, _ := InstalledVersions()
	for _, v := range versions {
		if v == version {
			return true
		}
	}
	return false
}

// FindProjectVersion finds the nearest .wade-version file between the current directory and the home directory.
func FindProjectVersion() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	cwd = filepath.Clean(cwd)
	home = filepath.Clean(home)
	cwd, err = filepath.EvalSymlinks(cwd)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	home, err = filepath.EvalSymlinks(home)
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	relative, err := filepath.Rel(home, cwd)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", ErrProjectVersionNotFound
	}

	for dir := cwd; ; dir = filepath.Dir(dir) {
		contents, err := os.ReadFile(filepath.Join(dir, ".wade-version"))
		if err == nil {
			return strings.TrimSpace(strings.SplitN(string(contents), "\n", 2)[0]), nil
		}
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("read .wade-version: %w", err)
		}
		if dir == home {
			break
		}
	}

	return "", ErrProjectVersionNotFound
}

func mustAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
