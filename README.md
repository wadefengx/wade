# Wade

> 🏄 All-in-one Node.js version & npm/yarn/pnpm registry manager. Install once, use everywhere.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](https://github.com/wadefengx/wade/releases)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Wade** replaces **nvm** + **cgr/nrm** — manages Node.js versions, switches npm/yarn/pnpm registries, and controls where Node binaries are downloaded from.

📖 [中文文档](README_zh.md) · 🌐 [Website](https://wadefengx.github.io/wade)

---

## Quick Start

```bash
wade -i          # Interactive setup wizard (recommended!)
wade init -y     # Non-interactive: use all defaults
```

Or manually:

```bash
wade node mirror mirror        # Use npmmirror for Node downloads (fast in China)
wade node install 20           # Install Node 20
wade registry use taobao       # Switch npm/yarn/pnpm to taobao mirror
wade status                    # Check everything
```

---

## Installation

### From Source (all platforms, requires Go 1.23+)

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o /usr/local/bin/wade .
wade -i           # Run interactive setup
```

### macOS (Homebrew — coming soon)

```bash
brew install wadefengx/tap/wade
wade -i
```

### Post-install PATH

After setup, add this to your shell config (`~/.zshrc`):

```bash
export PATH="$HOME/.wade/shims:$PATH"
```

---

## Commands

### `wade init` — Setup wizard

```bash
wade -i                 # Interactive setup (4 steps)
wade init               # Same as above
wade init -y            # Non-interactive, use all defaults
```

Guides through: Node download source → Node version → Registry mirror → PATH setup

### `wade node mirror` — Node download source

```bash
wade node mirror            # Show current source
wade node mirror mirror     # Use npmmirror.com (fast in China, default)
wade node mirror official   # Use official nodejs.org
```

### `wade node` — Version management

```bash
wade node install <ver>   # Install a version (e.g., "18", "20", "lts")
wade node use <ver>       # Activate a version
wade node ls              # Installed versions
wade node ls-remote       # Available versions from mirror
wade node default <ver>   # Set default
wade node uninstall <ver> # Remove a version
wade node current         # Print active version
```

### `wade registry` — npm/yarn/pnpm mirror

```bash
wade registry ls          # All registries (5 built-in + custom)
wade registry use <name>  # Switch ALL package managers at once
wade registry add <n> <u> # Add custom registry
wade registry del <name>  # Delete custom registry
wade registry test        # Latency test (fastest first)
```

Built-in registries: `npm`, `taobao`, `tencent`, `huawei`, `cnpm`

### `wade status` — Check environment

```
wade status
─────────────────────
  📦 Registry:      taobao → https://registry.npmmirror.com/
  🌐 Node download: mirror (npmmirror.com)
  🟢 Node ver:      v20.20.2 (default)
```

---

## How It Works

```
wade (single Go binary, no runtime dependencies)
    │
    ├── ~/.wade/versions/    ← Downloaded Node binaries
    ├── ~/.wade/shims/       ← Symlinks on PATH (set once)
    ├── ~/.wade/config.toml  ← User preferences + mirror config
    └── ~/.wade/current      ← Active Node version
```

---

## Development

Wade is built with **Go** and follows **Spec-Driven Development (SDD)**.

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o wade .
```

## License

MIT © [wadefengx](https://github.com/wadefengx)
