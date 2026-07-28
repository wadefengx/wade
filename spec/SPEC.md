# SPEC.md — Wade Tool Master Specification

> **This is the canonical specification for the wade tool.**
> All implementation must match this spec. If the spec is wrong, update the spec first.
> Refactoring = change implementation, keep spec. Feature change = update spec, then implement.

---

## 1. Overview

### 1.1 What is Wade?

Wade is a cross-platform CLI tool that manages Node.js versions and npm/yarn/pnpm registries through a single, self-contained binary. It is installed once via Homebrew (macOS), Scoop (Windows), or direct binary download, and requires no runtime dependency on Node.js.

### 1.2 Core Principles

1. **Single binary, installed once** — not an npm package, not tied to any Node version
2. **Shim-based version switching** — PATH is set once, switching versions is instant
3. **Registry switching affects all package managers** — one command for npm + yarn + pnpm
4. **Mirror-first downloads** — Node binary downloads go through configurable mirrors (default: npmmirror.com)
5. **Spec-driven development** — every feature is defined in a spec before implementation

### 1.3 Target Platforms

| Platform | Architectures | Package Manager |
|----------|--------------|-----------------|
| macOS | amd64, arm64 | Homebrew |
| Windows | amd64 | Scoop, winget |
| Linux | amd64, arm64 | Direct download |

---

## 2. Installation & Setup

### 2.1 Installation Methods

```bash
# macOS
brew install wadefengx/tap/wade

# Windows
scoop bucket add wade https://github.com/wadefengx/scoop-wade
scoop install wade

# Any platform (auto-detect)
curl -fsSL https://github.com/wadefengx/wade/releases/latest/download/install.sh | bash
```

### 2.2 Post-Install Setup

```bash
wade setup
```

**What setup does:**
1. Creates `~/.wade/` directory structure
2. Creates `~/.wade/shims/` directory
3. Creates initial `~/.wade/config.toml` with defaults
4. Detects shell (bash/zsh/fish/powershell) and prints PATH addition instructions
5. Optionally auto-adds to shell config file

### 2.3 PATH Configuration

The user must add one line to their shell config:

```bash
# ~/.zshrc or ~/.bashrc
export PATH="$HOME/.wade/shims:$PATH"
```

This is done once. After that, `wade node use` works instantly by updating symlinks in `~/.wade/shims/`.

---

## 3. Directory Structure

### 3.1 Wade's Home Directory (`~/.wade/`)

```
~/.wade/
├── shims/                  # Symlinks to current Node binaries (on PATH)
│   ├── node                # → ../versions/v20.12.0/bin/node
│   ├── npm                 # → ../versions/v20.12.0/bin/npm
│   ├── npx                 # → ../versions/v20.12.0/bin/npx
│   ├── yarn                # (if yarn installed in version)
│   └── pnpm                # (if pnpm installed in version)
├── versions/               # Downloaded Node versions
│   ├── v18.20.0/           # Full Node installation directory
│   ├── v20.12.0/
│   └── v22.4.0/
├── config.toml             # User configuration
└── current                 # File containing current version string (e.g. "v20.12.0")
```

### 3.2 Config File (`~/.wade/config.toml`)

```toml
# Default Node version (used when no .wade-version file found)
default_version = "v20.12.0"

# Node download mirror (default: npmmirror.com for China-friendly access)
node_mirror = "https://npmmirror.com/mirrors/node/"

# Current registry
current_registry = "taobao"

# Custom registries
[[registries]]
name = "mycompany"
url = "https://npm.mycompany.com/"
```

---

## 4. Command Specifications

### 4.1 `wade node install <version>

**Purpose:** Download and install a Node.js version.

**Behavior:**
1. Parse version string (supports: `18`, `18.20`, `18.20.0`, `lts`, `latest`)
2. If version is partial (e.g. `18`), resolve to latest matching version from remote mirror
3. Download Node binary tarball from `node_mirror` config URL
4. Show download progress bar
5. Verify SHA256 checksum (if provided by mirror)
6. Extract tarball to `~/.wade/versions/<full_version>/
7. If this is the first installed version, automatically set as default

**Edge cases:**
- Version already installed → print "already installed", skip
- Mirror unreachable → retry once, then error with helpful message
- Invalid version string → error with supported version format examples
- Disk full → error with clear message

**Exit codes:**
- 0: success
- 1: general error
- 2: version already installed

### 4.2 `wade node use <version>`

**Purpose:** Switch to a specific Node.js version.

**Behavior:**
1. Verify version is installed (if not, error with "try `wade node install <version>` first")
2. Update symlinks in `~/.wade/shims/`:
   - `node` → `~/.wade/versions/<version>/bin/node`
   - `npm` → `~/.wade/versions/<version>/bin/npm`
   - `npx` → `~/.wade/versions/<version>/bin/npx`
   - `yarn` → `~/.wade/versions/<version>/bin/yarn` (if exists)
   - `pnpm` → `~/.wade/versions/<version>/bin/pnpm` (if exists)
3. Write version string to `~/.wade/current`
4. Print: "Now using Node v20.12.0"

**Edge cases:**
- `~/.wade/shims/` doesn't exist → create it automatically
- Symlink already exists → overwrite
- Version not installed → error with suggestion
- No version argument → use current version from `~/.wade/current` (re-link)

### 4.3 `wade node ls`

**Purpose:** List installed Node.js versions.

**Behavior:**
1. Scan `~/.wade/versions/` directory
2. List each version with format:
   - `v20.12.0` (current) — marked with asterisk
   - `v18.20.0` (default) — marked with (D)
   - `v22.4.0`
3. Print in order (descending by version)

**Output format:**
```
Installed Node versions:
  v22.4.0
  v20.12.0 (current) (default)
  v18.20.0
```

### 4.4 `wade node ls-remote`

**Purpose:** List available Node.js versions from the mirror.

**Behavior:**
1. Fetch version list from `node_mirror`/index.json
2. Show last 20 LTS versions + latest current version
3. Mark installed versions with ✓

### 4.5 `wade node default <version>`

**Purpose:** Set the default Node.js version.

**Behavior:**
1. Verify version is installed
2. Update `config.toml` `default_version` field
3. If no version is currently active, also run `wade node use <version>`

### 4.6 `wade node uninstall <version>`

**Purpose:** Remove an installed Node.js version.

**Behavior:**
1. Verify version is installed
2. If version is currently active, refuse with "cannot uninstall active version"
3. If version is default, refuse with "cannot uninstall default version, change default first"
4. Remove `~/.wade/versions/<version>/` directory
5. Confirm removal

### 4.7 `wade node current`

**Purpose:** Print the currently active Node.js version.

**Output:** `v20.12.0`

### 4.8 `wade registry ls`

**Purpose:** List all configured registries.

**Behavior:**
1. Read preset registries from built-in list
2. Read custom registries from `config.toml`
3. Merge and display as table
4. Mark current registry with "*"

**Output format:**
```
  Registry   URL                                            Status
  ────────────────────────────────────────────────────────────────
  npm        https://registry.npmjs.org/
  taobao     https://registry.npmmirror.com/                *
  tencent    https://mirrors.tencent.com/npm/
  huawei     https://repo.huaweicloud.com/repository/npm/
  cnpm       http://r.cnpmjs.org/
  mycompany  https://npm.mycompany.com/                     (custom)
```

### 4.9 `wade registry use <name>`

**Purpose:** Switch all package managers to a registry.

**Behavior:**
1. Look up registry URL by name (preset registry → custom registry → error)
2. Execute in order (skip if package manager not found):
   - `npm config set registry <url>`
   - `yarn config set registry <url>` (if yarn is on PATH)
   - Test: `pnpm config set registry <url>` (if pnpm is on PATH)
3. Update `current_registry` in `config.toml`
4. Print success: "Switched to taobao registry"

**Edge cases:**
- Registry name not found → error with "use `wade registry ls` to see available registries"
- All package managers missing → warn but still save the config

### 4.10 `wade registry add <name> <url>`

**Purpose:** Add a custom registry.

**Behavior:**
1. Validate URL format (must start with http:// or https://)
2. Check for duplicate name (error if already exists)
3. Add to `config.toml` `[[registries]]` array
4. Print success

### 4.11 `wade registry del <name>`

**Purpose:** Delete a custom registry.

**Behavior:**
1. Cannot delete built-in preset registries
2. Verify registry exists in custom list
3. Remove from `config.toml`
4. If deleted registry was current, set current to "npm"

### 4.12 `wade registry test`

**Purpose:** Test latency of all registries.

**Behavior:**
1. For each registry (preset + custom), make a HEAD request to its URL
2. Time the response
3. Display sorted by latency (fastest first)
4. Timeout after 5 seconds per registry

**Output format:**
```
Testing registry latency...
  taobao      https://registry.npmmirror.com/     32ms
  tencent     https://mirrors.tencent.com/npm/     45ms
  npm         https://registry.npmjs.org/         120ms
  cnpm        http://r.cnpmjs.org/                180ms
  huawei      https://repo.huaweicloud.com/       200ms
```

### 4.13 `wade status`

**Purpose:** Show current environment status.

**Behavior:**
1. Detect current Node version (from `~/.wade/current`)
2. Detect current registry (from `npm config get registry`)
3. Detect shim health (all expected symlinks exist?)
4. Display as a formatted dashboard

**Output format:**
```
wade status
────────────────────────────────
  Node version:  v20.12.0
  Registry:      taobao (https://registry.npmmirror.com/)
  Shim path:     ~/.wade/shims (✓ in PATH)
  Shim health:   3/3 links OK
  Config:        ~/.wade/config.toml
```

### 4.14 `wade setup`

**Purpose:** Initialize wade environment.

**Behavior:**
1. Create `~/.wade/` directory structure
2. Create `~/.wade/shims/`
3. Create default `~/.wade/config.toml`
4. Write initial `~/.wade/current` (empty)
5. Detect shell and print PATH instructions

### 4.15 `wade version`

**Purpose:** Print version information.

**Output:** `wade v1.0.0`

### 4.16 `wade init` / `wade -i`

**Purpose:** Interactive setup wizard (4 steps: mirror → version → registry → PATH).

**Flags:** `-y, --yes` skip all Y/n prompts.

### 4.17 `wade node mirror [official|mirror]`

**Purpose:** Show or set Node.js binary download source. Different from registry: controls where `wade node install` downloads from, not where `npm install` downloads from.

---

## 5. Preset Registry Data

| Name | URL | Built-in |
|------|-----|----------|
| npm | https://registry.npmjs.org/ | ✅ |
| taobao | https://registry.npmmirror.com/ | ✅ |
| tencent | https://mirrors.tencent.com/npm/ | ✅ |
| huawei | https://repo.huaweicloud.com/repository/npm/ | ✅ |
| cnpm | http://r.cnpmjs.org/ | ✅ |

---

## 6. Version Resolution Rules

When a user specifies a partial version:

| Input | Resolves to | Example |
|-------|-------------|---------|
| `18` | Latest 18.x.x | `18.20.0` |
| `18.20` | Latest 18.20.x | `18.20.0` |
| `18.20.0` | Exact version | `18.20.0` |
| `lts` | Latest LTS | `20.12.0` |
| `latest` | Latest current | `22.4.0` |

---

## 7. Error Handling

### 7.1 Error Message Convention

All error messages follow this format:
```
Error: <human-readable message>
Suggestion: <action user can take>
```

Example:
```
Error: Version "v19.0.0" is not installed
Suggestion: Run 'wade node install 19' to install it
```

### 7.2 Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Already exists (version installed, registry name taken) |
| 3 | Not found (version not installed, registry not found) |
| 4 | Permission error (can't write to ~/.wade/) |

---

## 8. Testing Requirements

### 8.1 Unit Tests

Every function in `internal/` must have:
- Happy path test
- Edge case test (missing input, invalid input)
- Error case test

### 8.2 Integration Tests

- Download and extract a Node version (mock HTTP)
- Shim creation and switching
- Registry config write (mock filesystem)

### 8.3 Manual Test Matrix

| Platform | Node install | Node use | Registry switch | Shim health |
|----------|-------------|----------|-----------------|-------------|
| macOS ARM | ✅ | ✅ | ✅ | ✅ |
| macOS Intel | ✅ | ✅ | ✅ | ✅ |
| Windows x64 | ✅ | ✅ | ✅ | ✅ |
| Linux x64 | ✅ | ✅ | ✅ | ✅ |

---

## 9. Performance Targets

| Operation | Target |
|-----------|--------|
| `wade node use vX` | < 100ms |
| `wade registry use taobao` | < 200ms |
| `wade node ls` | < 50ms |
| `wade registry ls` | < 50ms |
| `wade status` | < 100ms |
| Binary size | < 10MB |
| Memory usage | < 10MB |