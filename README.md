# Wade

> 🏄 All-in-one Node.js version & registry manager. Install once, use everywhere.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](https://github.com/wadefengx/wade/releases)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Wade** replaces **nvm** + **cgr/nrm** — manages Node.js versions, and switches npm/yarn/pnpm registries with one command. Single binary, zero dependencies, works on macOS and Windows.

📖 [中文文档](README_zh.md) · 🌐 [Website](https://wadefengx.github.io/wade)

---

## Why Wade?

| Problem | nvm | cgr/nrm | **Wade** |
|---------|-----|---------|----------|
| Node version management | ✅ | ❌ | ✅ |
| Registry switching | ❌ | ✅ (npm only) | ✅ (npm + yarn + pnpm) |
| Global install, works everywhere | ❌ | ❌ (tied to Node version) | ✅ |
| Cross-platform | ✅ | ❌ | ✅ |
| Mirror downloads (China-friendly) | ❌ | — | ✅ |

---

## Installation

### From Source (all platforms, requires Go 1.23+)

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o /usr/local/bin/wade .   # or ~/.local/bin/
wade setup
```

### macOS (Homebrew — coming soon)

```bash
brew install wadefengx/tap/wade
```

### Windows (Scoop — coming soon)

```powershell
scoop bucket add wade https://github.com/wadefengx/scoop-wade
scoop install wade
```

### Post-install: add to PATH

Add this to your shell config (`~/.zshrc` / `~/.bashrc`):

```bash
export PATH="$HOME/.wade/shims:$PATH"
```

Then restart your terminal or run `source ~/.zshrc`.

---

## Quick Start

```bash
# ── Registry management (works immediately, no Node required) ──
wade registry ls                 # List all registries
wade registry use taobao         # Switch npm/yarn/pnpm to taobao mirror
wade registry test               # Test latency of all registries
wade registry add corp https://npm.mycompany.com/  # Add custom registry

# ── Node version management ──
wade node install 20             # Install Node 20 (from npmmirror.com)
wade node use 20                 # Switch to Node 20
wade node ls                     # List installed versions
wade node ls-remote              # See available versions
wade node default 20             # Set as default

# ── Status ──
wade status                      # See current environment
```

---

## Commands

### `wade registry`

```bash
wade registry ls          # All registries (built-in + custom)
wade registry use <name>  # Switch npm/yarn/pnpm at once
wade registry add <n> <u> # Add custom registry
wade registry del <name>  # Delete custom registry
wade registry test        # Latency test (fastest first)
```

Built-in registries:

| Registry | URL |
|----------|-----|
| `npm` | https://registry.npmjs.org/ |
| `taobao` | https://registry.npmmirror.com/ |
| `tencent` | https://mirrors.tencent.com/npm/ |
| `huawei` | https://repo.huaweicloud.com/repository/npm/ |
| `cnpm` | http://r.cnpmjs.org/ |

### `wade node`

```bash
wade node install <ver>   # Install a version (e.g., "18", "18.20", "latest")
wade node use <ver>       # Activate a version
wade node ls              # Installed versions
wade node ls-remote       # Available versions from mirror
wade node default <ver>   # Set default
wade node uninstall <ver> # Remove a version
wade node current         # Print active version
```

### `wade status`

```
wade status
────────────
  Registry:   taobao (https://registry.npmmirror.com/)
  Node ver:   v20.20.2 (default)
  Config:     ~/.wade/config.toml
```

---

## How It Works

```
wade (single Go binary — no runtime dependencies)
    │
    ├── ~/.wade/versions/    ← Downloaded Node binaries
    ├── ~/.wade/shims/       ← Symlinks on PATH (set once)
    ├── ~/.wade/config.toml  ← User preferences
    └── ~/.wade/current      ← Active version
```

- **Registry switching:** Writes to `npm config`, `yarn config`, and `pnpm config` simultaneously
- **Version switching:** Updates symlinks in `~/.wade/shims/` — instant, no shell reload
- **No Node dependency:** Wade runs without Node; installs Node from pre-built binaries on npmmirror.com

---

## Development

Wade is built with **Go** and follows **Spec-Driven Development (SDD)**.

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o wade .
./wade status
```

### Project Structure

```
wade/
├── AGENTS.md              # AI master context
├── spec/SPEC.md           # Complete specification
├── cmd/                   # CLI commands (cobra)
├── internal/
│   ├── config/            # TOML config
│   ├── registry/          # Registry switching
│   └── node/              # Node version management
└── docs/                  # GitHub Pages site
```

### Roadmap

- [x] **M1: Registry Management** — `registry ls/use/add/del/test`
- [x] **M2: Node Version Management** — `node install/use/ls/default/uninstall`
- [ ] **M3: Release** — GitHub Actions + Homebrew + Scoop + install script
- [ ] **M4: Polish** — Shell completions, self-update, Windows support

---

## License

MIT © [wadefengx](https://github.com/wadefengx)
