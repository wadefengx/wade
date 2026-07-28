# Wade Implementation Plan

> **For AI agents:** Follow the SDD cycle in AGENTS.md. Spec first, then implement.
> Each task is 2-5 minutes. Test after every task. Commit after every task.

**Goal:** Build an all-in-one Node version & registry manager (wade) in Go.
**Architecture:** Single binary, shim-based Node switching, registry switching via config writes. Cross-platform (macOS, Windows, Linux).
**Tech Stack:** Go + cobra + viper + go-toml + tablewriter
**Spec:** `spec/SPEC.md` — read this before implementing any feature.

---

## M0: Project Skeleton + CLI Framework

### Task 0.1: Initialize Go module

**Objective:** Create the Go module with cobra CLI framework

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`

**Steps:**

```bash
cd ~/Desktop/wade
go mod init github.com/wadefengx/wade
go get github.com/spf13/cobra
go get github.com/spf13/viper
go get github.com/pelletier/go-toml/v2
go get github.com/olekukonko/tablewriter
go get github.com/schollz/progressbar/v3
```

**Create `main.go`:**
```go
package main

import "github.com/wadefengx/wade/cmd"

func main() {
	cmd.Execute()
}
```

**Create `cmd/root.go`:**
```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wade",
	Short: "All-in-one Node.js version & npm registry manager",
	Long: `Wade manages Node.js versions and npm/yarn/pnpm registries.
Single binary, installed once, no Node.js dependency.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("help", "h", false, "help for wade")
}
```

**Verify:**
```bash
go build -o wade .
./wade --help
```

**Commit:**
```bash
git add go.mod go.sum main.go cmd/root.go
git commit -m "feat: initialize Go module with cobra CLI framework"
```

---

### Task 0.2: Create directory structure

**Objective:** Create all package directories

**Files:**
- Create: `internal/node/`
- Create: `internal/registry/`
- Create: `internal/config/`
- Create: `internal/platform/`
- Create: `spec/`

**Steps:**
```bash
mkdir -p internal/node internal/registry internal/config internal/platform
```

Add placeholder `.gitkeep` files if needed.

**Commit:**
```bash
git add internal/
git commit -m "chore: create package directory structure"
```

---

### Task 0.3: Implement config module

**Objective:** Config read/write from `~/.wade/config.toml`

**Spec reference:** Section 3.2

**Files:**
- Create: `internal/config/config.go`

**Steps:**

1. Define `Config` struct with:
   - `DefaultVersion string`
   - `NodeMirror string`
   - `CurrentRegistry string`
   - `Registries []CustomRegistry`
   - `CustomRegistry {Name string, URL string}`

2. Functions:
   - `LoadConfig() (*Config, error)` — read from `~/.wade/config.toml`, return defaults if not found
   - `SaveConfig(cfg *Config) error` — write to `~/.wade/config.toml`
   - `ConfigPath() string` — return `~/.wade/config.toml`
   - `WadeDir() string` — return `~/.wade/`

3. Defaults:
   - `NodeMirror`: `"https://npmmirror.com/mirrors/node/"`
   - `CurrentRegistry`: `"npm"`
   - Built-in registries (in Presets registry module, not here)

**Test:**
```bash
go test ./internal/config/ -v
```

**Commit:**
```bash
git add internal/config/config.go
git commit -m "feat(config): add config load/save with TOML support"
```

---

## M1: Registry Management (quick win, no Node dependency)

### Task 1.1: Implement registry presets

**Objective:** Define built-in registries and lookup logic

**Spec reference:** Section 5, Section 4.8-4.12

**Files:**
- Create: `internal/registry/presets.go`

**Steps:**

1. Define `Registry` struct:
   - `Name string`
   - `URL string`
   - `IsBuiltIn bool`

2. Define `PresetRegistries()` function returning all built-in registries:

```go
func PresetRegistries() []Registry {
    return []Registry{
        {Name: "npm", URL: "https://registry.npmjs.org/", IsBuiltIn: true},
        {Name: "taobao", URL: "https://registry.npmmirror.com/", IsBuiltIn: true},
        {Name: "tencent", URL: "https://mirrors.tencent.com/npm/", IsBuiltIn: true},
        {Name: "huawei", URL: "https://repo.huaweicloud.com/repository/npm/", IsBuiltIn: true},
        {Name: "cnpm", URL: "http://r.cnpmjs.org/", IsBuiltIn: true},
    }
}
```

3. `FindRegistry(name string) (*Registry, bool)` — search presets then custom from config

**Test:**
```go
func TestPresetRegistries(t *testing.T) {
    regs := PresetRegistries()
    if len(regs) != 5 {
        t.Errorf("expected 5 preset registries, got %d", len(regs))
    }
}

func TestFindRegistry(t *testing.T) {
    reg, ok := FindRegistry("taobao")
    if !ok { t.Fatal("expected to find taobao") }
    if reg.URL != "https://registry.npmmirror.com/" {
        t.Errorf("wrong URL: %s", reg.URL)
    }
}
```

**Commit:**
```bash
git add internal/registry/presets.go
git commit -m "feat(registry): add built-in registry presets"
```

---

### Task 1.2: Implement registry manager

**Objective:** Core registry switching logic

**Spec reference:** Section 4.9, 4.10, 4.11

**Files:**
- Create: `internal/registry/manager.go`

**Steps:**

1. `UseRegistry(name string) error`:
   - Find registry by name
   - Execute `npm config set registry <url>`
   - Execute `yarn config set registry <url>` (if yarn on PATH)
   - Execute `pnpm config set registry <url>` (if pnpm on PATH)
   - Save current registry to config

2. `AddRegistry(name, url string) error`:
   - Validate URL (must start with http:// or https://)
   - Check for duplicate
   - Add to config and save

3. `DeleteRegistry(name string) error`:
   - Verify it's a custom registry (not built-in)
   - Remove from config
   - If current registry was deleted, reset to "npm"

4. Helper: `execPackageManagerConfig(pm, registryURL string) error`
   - Uses `os/exec` to run `pm config set registry <url>`

**Test:**
```go
func TestUseRegistry(t *testing.T) {
    // Test with mock exec
}
```

**Commit:**
```bash
git add internal/registry/manager.go
git commit -m "feat(registry): add registry switching logic"
```

---

### Task 1.3: Implement registry tester

**Objective:** Latency testing for all registries

**Spec reference:** Section 4.12

**Files:**
- Create: `internal/registry/tester.go`

**Steps:**

1. `TestRegistries(registries []Registry) []TestResult`:
   - For each registry, make HEAD request with 5s timeout
   - Measure response time
   - Return sorted results (fastest first)

2. `TestResult` struct:
   - `Name string`
   - `URL string`
   - `Latency time.Duration`
   - `Error error`

**Test:**
```go
func TestTestRegistries(t *testing.T) {
    // Test with mock HTTP server
}
```

**Commit:**
```bash
git add internal/registry/tester.go
git commit -m "feat(registry): add latency testing"
```

---

### Task 1.4: Implement registry CLI commands

**Objective:** Wire up cobra commands for all registry operations

**Spec reference:** Section 4.8-4.12

**Files:**
- Create: `cmd/registry.go`

**Steps:**

Use cobra to create:
```go
var registryCmd = &cobra.Command{Use: "registry", Short: "Manage npm/yarn/pnpm registries"}

// Subcommands
var registryLsCmd = &cobra.Command{Use: "ls", RunE: ...}
var registryUseCmd = &cobra.Command{Use: "use <name>", RunE: ...}
var registryAddCmd = &cobra.Command{Use: "add <name> <url>", RunE: ...}
var registryDelCmd = &cobra.Command{Use: "del <name>", RunE: ...}
var registryTestCmd = &cobra.Command{Use: "test", RunE: ...}
```

**Verify:**
```bash
go build -o wade .
./wade registry ls
./wade registry use taobao
./wade registry test
```

**Commit:**
```bash
git add cmd/registry.go
git commit -m "feat(registry): add CLI commands for registry management"
```

---

## M2: Node Version Management

### Task 2.1: Implement version resolution

**Objective:** Parse and resolve Node version strings

**Spec reference:** Section 6

**Files:**
- Create: `internal/node/versions.go`

**Steps:**

1. `ParseVersion(input string) (string, error)` — handle `18`, `18.20`, `18.20.0`, `lts`, `latest`
2. `ResolveVersion(partial string) (string, error)` — fetch remote index and resolve partial to full
3. `FetchRemoteVersions(mirrorURL string) ([]string, error)` — fetch and parse index.json from mirror
4. `InstalledVersions() []string` — scan `~/.wade/versions/` directory
5. `SortVersions(versions []string)` — semver sort (descending)

**Test:**
```go
func TestParseVersion(t *testing.T) {
    v, err := ParseVersion("18.20.0")
    if err != nil { t.Fatal(err) }
    if v != "18.20.0" { t.Errorf("expected 18.20.0, got %s", v) }
}

func TestParsePartialVersion(t *testing.T) {
    v, err := ParseVersion("18")
    if err != nil { t.Fatal(err) }
    if v != "18" { t.Errorf("expected 18, got %s", v) }
}
```

**Commit:**
```bash
git add internal/node/versions.go
git commit -m "feat(node): add version parsing and resolution"
```

---

### Task 2.2: Implement Node download and extraction

**Objective:** Download Node binary from mirror and extract

**Spec reference:** Section 4.1

**Files:**
- Create: `internal/node/manager.go`

**Steps:**

1. `InstallVersion(version string) error`:
   - Resolve version (if partial)
   - Check if already installed
   - Build download URL: `<mirror>/v<version>/node-v<version>-<os>-<arch>.tar.xz`
   - Download with progress bar
   - Extract to `~/.wade/versions/v<version>/`
   - If first version, set as default

2. `platformFilename(version, os, arch string) string`:
   - macOS ARM: `node-v<version>-darwin-arm64.tar.xz`
   - macOS Intel: `node-v<version>-darwin-x64.tar.xz`
   - Windows: `node-v<version>-win-x64.zip`
   - Linux: `node-v<version>-linux-x64.tar.xz`

3. `downloadFile(url, dest string) error` — HTTP download with progress
4. `extractTarXZ(src, dest string) error` — extract tar.xz

**Commit:**
```bash
git add internal/node/manager.go
git commit -m "feat(node): add Node download and extraction"
```

---

### Task 2.3: Implement shim mechanism

**Objective:** Symlink-based version switching

**Spec reference:** Section 5.1, 4.2

**Files:**
- Create: `internal/node/shim.go`

**Steps:**

1. `ShimDir() string` — return `~/.wade/shims/`
2. `UseVersion(version string) error`:
   - Verify version directory exists
   - Create/update symlinks in shims dir:
     - `node` → `~/.wade/versions/<version>/bin/node`
     - `npm` → `~/.wade/versions/<version>/bin/npm`
     - `npx` → `~/.wade/versions/<version>/bin/npx`
     - `yarn` → (if exists in version)
     - `pnpm` → (if exists in version)
   - Write version to `~/.wade/current`
3. `CurrentVersion() (string, error)` — read `~/.wade/current`
4. `ShimHealth() (ShimStatus, error)` — check all symlinks are valid

**Test:**
```go
func TestUseVersion(t *testing.T) {
    // Test with temp dir
    tmpDir := t.TempDir()
    // Create mock version directory
    // Call UseVersion
    // Verify symlinks
}
```

**Commit:**
```bash
git add internal/node/shim.go
git commit -m "feat(node): add shim-based version switching"
```

---

### Task 2.4: Implement Node CLI commands

**Objective:** Wire up cobra commands for all Node operations

**Spec reference:** Section 4.1-4.7

**Files:**
- Create: `cmd/node.go`

**Steps:**

Create cobra commands:
```go
var nodeCmd = &cobra.Command{Use: "node", Short: "Manage Node.js versions"}

var nodeInstallCmd = &cobra.Command{Use: "install <version>", RunE: ...}
var nodeUseCmd = &cobra.Command{Use: "use <version>", RunE: ...}
var nodeLsCmd = &cobra.Command{Use: "ls", RunE: ...}
var nodeLsRemoteCmd = &cobra.Command{Use: "ls-remote", RunE: ...}
var nodeDefaultCmd = &cobra.Command{Use: "default <version>", RunE: ...}
var nodeUninstallCmd = &cobra.Command{Use: "uninstall <version>", RunE: ...}
var nodeCurrentCmd = &cobra.Command{Use: "current", RunE: ...}
```

**Verify:**
```bash
go build -o wade .
./wade node ls
./wade node current
```

**Commit:**
```bash
git add cmd/node.go
git commit -m "feat(node): add CLI commands for Node version management"
```

---

## M3: Status + Setup + Polish

### Task 3.1: Implement `wade status`

**Spec reference:** Section 4.13

**Files:**
- Create: `cmd/status.go`

**Steps:**

1. Gather:
   - Current Node version (from `~/.wade/current`)
   - Current registry (from `npm config get registry`)
   - Shim health (check symlinks)
   - Config file location
2. Display as formatted table

**Commit:**
```bash
git add cmd/status.go
git commit -m "feat: add wade status command"
```

---

### Task 3.2: Implement `wade setup`

**Spec reference:** Section 4.14

**Files:**
- Create: `cmd/setup.go`

**Steps:**

1. Create `~/.wade/` directory structure
2. Create default config
3. Detect shell, print PATH addition instructions
4. Optionally auto-add to shell config

**Commit:**
```bash
git add cmd/setup.go
git commit -m "feat: add wade setup command"
```

---

## M4: Cross-Platform + Release

### Task 4.1: Platform abstraction layer

**Files:**
- Create: `internal/platform/darwin.go`
- Create: `internal/platform/windows.go`
- Create: `internal/platform/linux.go`

**Steps:**

1. Define interface:
   ```go
   type Platform interface {
       OS() string
       Arch() string
       HomeDir() string
       PathSeparator() string
       ShellConfigFile() string  // ~/.zshrc, ~/.bashrc, etc.
       NodeBinaryName() string    // "node" or "node.exe"
   }
   ```

2. Implement for each platform using build tags: `//go:build darwin`

**Commit:**
```bash
git add internal/platform/
git commit -m "feat(platform): add OS abstraction layer"
```

---

### Task 4.2: GitHub Actions CI/CD

**Files:**
- Create: `.github/workflows/release.yml`

**Steps:**

1. On push tag `v*`:
   - Build for: `darwin/arm64`, `darwin/amd64`, `windows/amd64`, `linux/amd64`
   - Create archives (tar.gz for macOS/Linux, zip for Windows)
   - Upload to GitHub Releases
   - Include checksums

**Commit:**
```bash
git add .github/workflows/release.yml
git commit -m "ci: add cross-compile release workflow"
```

---

### Task 4.3: Homebrew formula template

**Files:**
- Create: `scripts/homebrew-formula.rb`

**Steps:**

1. Create formula template with placeholder for SHA
2. Create `Makefile` with build/test/release targets

**Commit:**
```bash
git add scripts/ Makefile
git commit -m "docs: add Homebrew formula template and Makefile"
```

---

### Task 4.4: Install script

**Files:**
- Create: `scripts/install.sh`

**Steps:**

1. `curl | bash` installer:
   - Detect OS/arch
   - Download latest binary from GitHub Releases
   - Install to `/usr/local/bin/` or `~/.local/bin/`
   - Run `wade setup`

**Commit:**
```bash
git add scripts/install.sh
git commit -m "feat: add curl | bash install script"
```

---

## M5: Final Polish

### Task 5.1: Shell completions

- Generate bash/zsh/fish/powershell completions
- `wade completion bash/zsh/fish/powershell`

### Task 5.2: Self-update

- `wade update` — check latest release, download and replace binary

### Task 5.3: Project pinning

- `wade init` — write `.wade-version` file in current directory
- Auto-detect `.wade-version` on `wade node use`

---

## M6: Interactive Setup Wizard

### Task 6.1: `wade -i` shortcut
- Add `-i` flag to root command → launches interactive wizard

### Task 6.2: `wade init` wizard
- 4-step interactive wizard: Node mirror → version → registry → PATH
- `wade init -y` auto-mode (skip prompts)

### Task 6.3: `wade node mirror`
- `wade node mirror` — show current download source
- `wade node mirror official` — switch to nodejs.org
- `wade node mirror mirror` — switch to npmmirror.com

---

## Verification Checklist

After each milestone, run:

```bash
# Build
go build -o wade .

# Test
go test ./... -v

# Lint
gofmt -d ./

# Commit
git add -A
git commit -m "milestone: M<N> complete"
```

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| npm mirror index format changes | Node install breaks | Version resolution as separate module, easy to patch |
| Windows symlink restrictions | Shim mechanism fails | Use directory junctions on Windows |
| brew install complexity | User friction | Also provide direct binary download |
| Cross-platform path differences | Bugs on Windows | Platform abstraction layer from day one |