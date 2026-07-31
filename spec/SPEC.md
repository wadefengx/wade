# SPEC.md — Wade Tool Master Specification

> **This is the canonical specification for the wade tool.**
> All implementation must match this spec. If the spec is wrong, update the spec first.
> Refactoring = change implementation, keep spec. Feature change = update spec, then implement.

---

## 1. Overview

### 1.1 What is Wade?

Wade is a cross-platform CLI tool that manages **Node.js**, **Go**, and **Python** development environments through a single, self-contained binary. It handles version switching, registry/mirror management, and PATH setup — eliminating the need for nvm, gvm, pyenv, nrm, and other fragmented tools.

### 1.2 Core Principles

1. **Single binary, installed once** — not an npm package, not tied to any Node/Go/Python version
2. **Shim-based version switching** — PATH is set once, switching versions is instant
3. **Registry switching affects all package managers** — one command for npm + yarn + pnpm + pip
4. **Mirror-first downloads** — binary downloads go through configurable mirrors (China-friendly defaults)
5. **Multi-runtime** — manage Node.js, Go, and Python from one tool
6. **Spec-driven development** — every feature is defined in a spec before implementation

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
4. Detects shell (bash/zsh/fish/powershell)
5. Automatically adds `~/.wade/shims` to shell config file (interactive: confirm first; `--auto`: skip prompt)
6. Prints PATH confirmation and next steps

**Flags:**
- `--auto` — skip all prompts, auto-configure
- `--dry-run` — show what would be done without writing

### 2.3 PATH Configuration

The setup command adds one line to the user's shell config:

```bash
# ~/.zshrc or ~/.bashrc
export PATH="$HOME/.wade/shims:$PATH"
```

**Shell detection:** reads `$SHELL` env var, maps to config file:
- `zsh` → `~/.zshrc` (fallback: `~/.zprofile`)
- `bash` → `~/.bashrc`
- `fish` → `~/.config/fish/config.fish`
- `powershell` → `Documents/PowerShell/Microsoft.PowerShell_profile.ps1`

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

**Purpose:** Show current environment status for all runtimes.

**Behavior:**
1. Detect current Node version (from `~/.wade/current`)
2. Detect current registry (from `npm config get registry`)
3. Detect current Go version (from `~/.wade/go/current`)
4. Detect system Go (if not managed by wade)
5. Detect Go mirror and proxy settings
6. Detect system Python versions
7. Display as a formatted dashboard

**Output format:**
```
🏄  wade status
─────────────────────
  🟢 Node:        v20.12.0 (default)
  📦 Registry:    taobao → https://registry.npmmirror.com/
  🔵 Go:          go1.23.4 (default)
  🌐 Go mirror:   npmmirror
  🐍 Python:      3.11.9, 3.12.3
  ⚙️  Config:      ~/.wade/config.toml
  📁 Wade dir:    ~/.wade/

  💡 Try: wade -i | wade go ls | wade python registry ls
```

### 4.14 `wade setup`

**Purpose:** One-command environment setup. Creates directories, configures PATH, and verifies.

**Flags:**
- `--auto` — skip all prompts, auto-configure
- `--dry-run` — show what would be done without writing

**Behavior:**
1. Create `~/.wade/` directory structure (`shims/`, `versions/`, `go/versions/`)
2. Create default `~/.wade/config.toml` if not exists
3. Detect shell from `$SHELL` env var
4. Check if `~/.wade/shims` is already in shell config
5. If not:
   - Interactive mode: show the line, ask `[Y/n]`
   - `--auto` mode: append without prompt
   - `--dry-run` mode: show what would be written without writing
6. Append `export PATH="$HOME/.wade/shims:$PATH"` to shell config file
7. Print success message with next steps
8. Exit with code 0 on success

**Edge cases:**
- Shell config file doesn't exist → create it
- `~/.wade/` already exists → skip creation, don't overwrite config
- Shell unknown → print manual instructions, exit 0

### 4.15 `wade version`

**Purpose:** Print version information.

**Output:** `wade v1.0.0`

### 4.16 `wade init` / `wade -i`

**Purpose:** Interactive setup wizard with runtime selection.

**Behavior:**
1. **Step 0: Runtime selection** — ask "Which runtimes to configure?" (Node.js / Go / Python / All of the above)
2. For each selected runtime, configure its mirror, version, and settings in sequence
3. **Node.js:** mirror source → install version → set as default
4. **Go:** mirror source → proxy → install version → set as default
5. **Python:** pip mirror → show detected system Python versions
6. Create `~/.wade/` directory structure if needed
7. Configure PATH (same as `wade setup`)

**Flags:**
- `-y, --yes` — skip all prompts, use defaults (China-friendly: npmmirror + goproxy.cn + tsinghua pip)
- `-i` — shortcut for `wade init` (registered on root command)

**`wade init` (no flags, in a project dir):** Write `.wade-version` file for current project (Node.js version pinning).

### 4.17 `wade node mirror [official|mirror]`

**Purpose:** Show or set Node.js binary download source. Different from registry: controls where `wade node install` downloads from, not where `npm install` downloads from.

---

### 4.18 `wade go install <version>`

**Purpose:** Download and install a Go version.

**Behavior:**
1. Parse version string (supports `1.23`, `go1.23.4`, `1.23.4`)
2. Auto-prefix `go` if missing (e.g., `1.23` → `go1.23`)
3. Download Go tarball from configured mirror (default: `go.dev/dl/`)
4. Extract to `~/.wade/go/versions/<version>/`
5. If first installed version, auto-set as default

**Edge cases:**
- Version already installed → error "already installed"
- Mirror unreachable → retry once, then error

### 4.19 `wade go use <version>`

**Purpose:** Switch to a Go version via shim.

**Behavior:**
1. Verify version is installed
2. Update symlinks in `~/.wade/shims/`: `go` and `gofmt` → `~/.wade/go/versions/<version>/bin/`
3. Write version to `~/.wade/go/current`

### 4.20 `wade go ls`

**Purpose:** List installed Go versions.

**Output format:**
```
📦 Installed Go versions:
  go1.23.4 (current) (default)
  go1.22.2
  go1.21.0
```

### 4.21 `wade go ls-remote`

**Purpose:** List available Go versions from the mirror.

**Behavior:**
1. Fetch version list from mirror's `?mode=json` endpoint
2. Filter stable versions only
3. Sort descending (semver)
4. Show last 20 versions, mark installed with ✓

### 4.22 `wade go mirror`

**Purpose:** Show current Go download mirror.

### 4.23 `wade go mirror ls`

**Purpose:** List Go download mirrors.

**Built-in mirrors:**
| Name | URL |
|------|-----|
| official | https://go.dev/dl/ |
| google-cn | https://golang.google.cn/dl/ |
| npmmirror | https://npmmirror.com/mirrors/go/ |
| aliyun | https://mirrors.aliyun.com/go/ |

### 4.24 `wade go mirror use <name>`

**Purpose:** Switch Go download mirror.

### 4.25 `wade go mirror test`

**Purpose:** Test latency of all Go mirrors.

### 4.26 `wade go proxy`

**Purpose:** Show current Go module proxy.

### 4.27 `wade go proxy ls`

**Purpose:** List Go proxies.

**Built-in proxies:**
| Name | URL |
|------|-----|
| official | https://proxy.golang.org,direct |
| goproxy-cn | https://goproxy.cn,direct |
| aliyun | https://mirrors.aliyun.com/goproxy/,direct |

### 4.28 `wade go proxy use <name>`

**Purpose:** Switch Go module proxy by running `go env -w GOPROXY=<url>`.

---

### 4.29 `wade python ls`

**Purpose:** Detect and list system Python versions.

**Behavior:**
1. Probe `python3 --version` and `python --version`
2. Display all detected versions with source info

**Output format:**
```
🐍 Detected Python:
  Python 3.11.9 (system: /usr/bin/python3)
  Python 3.12.3 (system: /opt/homebrew/bin/python3)
```

**Note:** Wade does NOT install Python interpreters. It relies on system Python and only manages pip mirrors.

### 4.30 `wade python registry ls`

**Purpose:** List pip mirrors.

**Behavior:**
1. Display built-in pip mirror presets

**Built-in mirrors:**
| Name | URL |
|------|-----|
| official | https://pypi.org/simple/ |
| tsinghua | https://pypi.tuna.tsinghua.edu.cn/simple/ |
| aliyun | https://mirrors.aliyun.com/pypi/simple/ |
| tencent | https://mirrors.cloud.tencent.com/pypi/simple/ |
| ustc | https://pypi.mirrors.ustc.edu.cn/simple/ |

### 4.31 `wade python registry use <name>`

**Purpose:** Switch pip to a mirror by running `pip config set global.index-url <url>`.

---

## 5. Milestone Roadmap

| Milestone | Content | Status |
|-----------|---------|--------|
| **M0: Skeleton** | Go module, cobra CLI, AGENTS.md, SPEC.md | ✅ |
| **M1: Registry** | `wade registry ls/use/add/del/test` | ✅ |
| **M2: Node** | `wade node install/use/ls/default/uninstall/ls-remote` | ✅ |
| **M3: Release** | GitHub Actions, Homebrew, Scoop, install script | ✅ |
| **M4: Polish** | `wade status`, shell completions, self-update | ✅ |
| **M5: Init Wizard** | Multi-runtime `wade -i` / `wade init`, survey UX | ✅ |
| **M6: Multi-Runtime** | Go version/mirror/proxy, Python pip mirror | ✅ |
| **M7: Quality** | Unit tests, integration tests, `wade setup --auto` | ✅ |
| **M8: Distribution** | Create Homebrew tap, Scoop bucket, publish docs site | ✅ |

---

## 6. Preset Registry Data

| Name | URL | Built-in |
|------|-----|----------|
| npm | https://registry.npmjs.org/ | ✅ |
| taobao | https://registry.npmmirror.com/ | ✅ |
| tencent | https://mirrors.tencent.com/npm/ | ✅ |
| huawei | https://repo.huaweicloud.com/repository/npm/ | ✅ |
| cnpm | http://r.cnpmjs.org/ | ✅ |

---

## 7. Version Resolution Rules

When a user specifies a partial version:

| Input | Resolves to | Example |
|-------|-------------|---------|
| `18` | Latest 18.x.x | `18.20.0` |
| `18.20` | Latest 18.20.x | `18.20.0` |
| `18.20.0` | Exact version | `18.20.0` |
| `lts` | Latest LTS | `20.12.0` |
| `latest` | Latest current | `22.4.0` |

---

## 8. Error Handling

### 8.1 Error Message Convention

### 8.2 Exit Codes
```
Error: <human-readable message>
Suggestion: <action user can take>
```

Example:
```
Error: Version "v19.0.0" is not installed
Suggestion: Run 'wade node install 19' to install it
```

### 8.2 Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Already exists (version installed, registry name taken) |
| 3 | Not found (version not installed, registry not found) |
| 4 | Permission error (can't write to ~/.wade/) |

---

## 9. Testing Requirements

### 9.1 Unit Tests

Every function in `internal/` must have:
- Happy path test
- Edge case test (missing input, invalid input)
- Error case test

### 9.2 Integration Tests

- Download and extract a Node version (mock HTTP)
- Shim creation and switching
- Registry config write (mock filesystem)

### 9.3 Manual Test Matrix

| Platform | Node install | Node use | Registry switch | Shim health |
|----------|-------------|----------|-----------------|-------------|
| macOS ARM | ✅ | ✅ | ✅ | ✅ |
| macOS Intel | ✅ | ✅ | ✅ | ✅ |
| Windows x64 | ✅ | ✅ | ✅ | ✅ |
| Linux x64 | ✅ | ✅ | ✅ | ✅ |

---

## 10. Performance Targets

| Operation | Target |
|-----------|--------|
| `wade node use vX` | < 100ms |
| `wade registry use taobao` | < 200ms |
| `wade node ls` | < 50ms |
| `wade registry ls` | < 50ms |
| `wade status` | < 100ms |
| Binary size | < 10MB |
| Memory usage | < 10MB |