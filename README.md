# Wade

> 🏄 All-in-one Node.js version & registry manager. Install once, use everywhere.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](https://github.com/wadefengx/wade/releases)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Wade** is a single binary that replaces **nvm** + **cgr/nrm** — manages Node.js versions, and switches npm/yarn/pnpm registries with one command.

### Why Wade?

| Problem | nvm | cgr/nrm | **Wade** |
|---------|-----|---------|----------|
| Node version management | ✅ | ❌ | ✅ |
| Registry switching | ❌ | ✅ (npm only) | ✅ (npm + yarn + pnpm) |
| Global install, works everywhere | ❌ (per-version isolation) | ❌ (npm global, tied to Node version) | ✅ (single binary) |
| Cross-platform (macOS + Windows) | ✅ | ❌ | ✅ |

---

## Installation

### macOS

```bash
brew install wadefengx/tap/wade
```

### Windows

```powershell
scoop bucket add wade https://github.com/wadefengx/scoop-wade
scoop install wade
```

### Any platform (one-liner)

```bash
curl -fsSL https://github.com/wadefengx/wade/releases/latest/download/install.sh | bash
```

### Post-install

```bash
wade setup   # Creates ~/.wade/ and prints PATH instructions
```

Add this to your shell config (`.zshrc` / `.bashrc`):

```bash
export PATH="$HOME/.wade/shims:$PATH"
```

---

## Quick Start

```bash
# Registry management (works immediately, no Node required)
wade registry ls                 # List all registries
wade registry use taobao         # Switch all PMs to taobao mirror
wade registry test               # Test latency of all registries
wade registry add corp https://npm.mycompany.com/  # Add custom registry

# Node version management (coming soon)
wade node install 20             # Install Node 20
wade node use 20                 # Switch to Node 20
wade node ls                     # List installed versions

# Status
wade status                      # Show current environment
```

---

## Commands

### `wade registry` — Manage registries

```bash
wade registry ls          # List all registries (built-in + custom)
wade registry use <name>  # Switch npm/yarn/pnpm to a registry
wade registry add <n> <u> # Add custom registry
wade registry del <name>  # Delete custom registry
wade registry test        # Test latency of all registries
```

Built-in registries:

| Name | URL |
|------|-----|
| `npm` | https://registry.npmjs.org/ |
| `taobao` | https://registry.npmmirror.com/ |
| `tencent` | https://mirrors.tencent.com/npm/ |
| `huawei` | https://repo.huaweicloud.com/repository/npm/ |
| `cnpm` | http://r.cnpmjs.org/ |

### `wade node` — Manage Node.js versions (🚧 in progress)

```bash
wade node install <ver>   # Install a Node version
wade node use <ver>       # Switch to a version
wade node ls              # List installed versions
wade node ls-remote       # List available versions
wade node default <ver>   # Set default version
wade node uninstall <ver> # Remove a version
```

### `wade status` — Show current state

```
wade status
────────────
  Registry:   taobao (https://registry.npmmirror.com/)
  Config:     ~/.wade/config.toml
  Wade dir:   ~/.wade
```

---

## How It Works

```
wade (single Go binary — no runtime dependencies)
     │
     ├── ~/.wade/versions/    ← Downloaded Node binaries
     ├── ~/.wade/shims/       ← Symlinks on PATH (set once)
     ├── ~/.wade/config.toml  ← Your preferences
     └── ~/.wade/current      ← Active version
```

- **Registry switching:** Writes to `npm config`, `yarn config`, and `pnpm config` simultaneously
- **Version switching:** Updates symlinks in `~/.wade/shims/` — instant, no shell reload
- **No Node dependency:** Wade runs without Node. Installs Node by downloading pre-built binaries from mirrors

---

## Development

Wade is built with **Go** and follows **Spec-Driven Development (SDD)**.

```bash
# Build
go build -o wade .

# Test
go test ./...

# Run
./wade status
```

### Project Architecture

```
wade/
├── AGENTS.md              # AI master context (read this first)
├── spec/SPEC.md           # Complete specification
├── cmd/                   # CLI commands (cobra)
├── internal/
│   ├── config/            # TOML config management
│   ├── registry/          # Registry switching logic
│   ├── node/              # Node version management (in progress)
│   └── platform/          # Cross-platform abstraction
└── .github/workflows/     # CI/CD
```

### Roadmap

- [x] **M1: Registry Management** — `registry ls/use/add/del/test`
- [ ] **M2: Node Version Management** — `node install/use/ls/default/uninstall`
- [ ] **M3: Cross-Platform Release** — GitHub Actions + Homebrew + Scoop
- [ ] **M4: Polish** — `status`, shell completions, self-update

---

## License

MIT © [wadefengx](https://github.com/wadefengx)
