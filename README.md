# Wade

> 🏄 All-in-one runtime manager: Node.js · Go · Python. Single binary, zero deps.

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](https://github.com/wadefengx/wade/releases)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

Wade replaces **nvm + cgr/nrm + gvm + pyenv**. Single binary, install once, works everywhere.

📖 [中文文档](README_zh.md) · 🌐 [Website](https://wadefengx.github.io/wade)

---

## Installation

### Pre-built binary (recommended, no Go required)

Download the latest binary from [GitHub Releases](https://github.com/wadefengx/wade/releases):

| Platform | Download |
|----------|----------|
| **macOS (Apple Silicon)** | `wade-darwin-arm64.tar.gz` |
| **macOS (Intel)** | `wade-darwin-amd64.tar.gz` |
| **Linux (x64)** | `wade-linux-amd64.tar.gz` |
| **Windows (x64)** | `wade-windows-amd64.zip` |

```bash
# Example: macOS Apple Silicon
tar xzf wade-darwin-arm64.tar.gz
sudo mv wade /usr/local/bin/
wade -i
```

### macOS — Homebrew

```bash
brew install wadefengx/tap/wade
wade -i
```

### Windows — Scoop

```powershell
scoop bucket add wade https://github.com/wadefengx/scoop-wade
scoop install wade
wade -i
```

### Linux / macOS — curl installer

```bash
curl -fsSL https://github.com/wadefengx/wade/releases/latest/download/install.sh | bash
wade -i
```

### From source (requires Go 1.23+)

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o /usr/local/bin/wade .
wade -i
```

### Post-install: add to PATH

After `wade -i` this is done automatically. If you need to do it manually:

| Platform | Config file | Line to add |
|----------|------------|-------------|
| macOS / Linux (zsh) | `~/.zshrc` | `export PATH="$HOME/.wade/shims:$PATH"` |
| macOS / Linux (bash) | `~/.bashrc` | `export PATH="$HOME/.wade/shims:$PATH"` |
| Windows (PowerShell) | `$PROFILE` | `$env:Path = "$env:USERPROFILE\.wade\shims;$env:Path"` |
| Windows (CMD) | System PATH | Add `%USERPROFILE%\.wade\shims` via System Properties |

---

## Quick Start

```bash
wade -i          # Interactive setup — choose Node, Go, Python, or all
wade init -y     # Non-interactive: auto-config all three (China-friendly)
```

### Per-runtime

| Runtime | Install | Switch | Lists | Mirror | Registry |
|---------|---------|--------|-------|--------|----------|
| **Node** | `wade node install 20` | `wade node use 20` | `ls` / `ls-remote` | `node mirror` | `registry` |
| **Go** | `wade go install 1.23` | `wade go use 1.23` | `ls` / `ls-remote` | `go mirror` | `go proxy` |
| **Python** | (system) | — | `python ls` | — | `python registry` |

---

## Commands

### `wade -i` / `wade init` — Setup wizard

```bash
wade -i                 # 3-step interactive setup
wade init -y            # Non-interactive (defaults)
wade init               # Write .wade-version for current project
```

### `wade node` — Version management

```bash
wade node install 20        # Install Node 20
wade node use 20            # Switch to Node 20
wade node ls                # List installed versions
wade node ls-remote         # Browse available versions
wade node default 20        # Set default version
wade node uninstall 18      # Remove a version
wade node mirror            # Show download source
wade node mirror mirror     # Use npmmirror.com (fast in China)
wade node mirror official   # Use nodejs.org
```

### `wade registry` — npm/yarn/pnpm mirror

```bash
wade registry ls            # List all registries
wade registry use taobao    # Switch all package managers
wade registry add x https://x.com/   # Add custom
wade registry test          # Speed test
```

Built-in: `npm`, `taobao`, `tencent`, `huawei`, `cnpm`

---

## Development

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o wade .
```

See [CHANGELOG.md](CHANGELOG.md) for version history.
AI tools: read [AGENTS.md](AGENTS.md) for project context.

## License

MIT © [wadefengx](https://github.com/wadefengx)
