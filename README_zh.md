# Wade

> 🏄 一站式 Node.js 版本与镜像源管理器。单二进制，零依赖。

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

📖 [English](README.md) · 🌐 [Website](https://wadefengx.github.io/wade)

---

## 安装

### 预编译二进制（推荐，不需要装 Go）

从 [GitHub Releases](https://github.com/wadefengx/wade/releases) 下载对应平台：

| 平台 | 下载文件 |
|------|---------|
| **macOS (Apple Silicon)** | `wade-darwin-arm64.tar.gz` |
| **macOS (Intel)** | `wade-darwin-amd64.tar.gz` |
| **Linux (x64)** | `wade-linux-amd64.tar.gz` |
| **Windows (x64)** | `wade-windows-amd64.zip` |

```bash
# 以 macOS Apple Silicon 为例
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

### 一行安装（macOS / Linux）

```bash
curl -fsSL https://github.com/wadefengx/wade/releases/latest/download/install.sh | bash
wade -i
```

### PATH 配置

运行 `wade -i` 会自动配置。如需手动配置：

| 平台 | 配置文件 | 添加内容 |
|------|---------|---------|
| macOS / Linux (zsh) | `~/.zshrc` | `export PATH="$HOME/.wade/shims:$PATH"` |
| macOS / Linux (bash) | `~/.bashrc` | `export PATH="$HOME/.wade/shims:$PATH"` |
| Windows (PowerShell) | `$PROFILE` | `$env:Path = "$env:USERPROFILE\.wade\shims;$env:Path"` |
| Windows (CMD) | 系统 PATH | 添加 `%USERPROFILE%\.wade\shims` |

---

## 快速开始

```bash
wade -i          # 交互式设置（推荐）
```

或手动操作：

```bash
wade node install 20           # 安装 Node 20
wade registry use taobao       # 换淘宝镜像
wade status
```

---

## 命令一览

| 命令 | 说明 |
|------|------|
| `wade -i` | 交互式设置向导 |
| `wade init -y` | 非交互式（默认值） |
| `wade node install 20` | 安装 Node 20 |
| `wade node use 20` | 切换到 Node 20 |
| `wade node mirror official` | Node 下载切到官方源 |
| `wade registry use taobao` | npm/yarn/pnpm 切淘宝 |
| `wade registry test` | 镜像测速 |
| `wade status` | 当前状态 |

---

## 开发

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o wade .
```

变更历史见 [CHANGELOG.md](CHANGELOG.md)。AI 工具请先读 [AGENTS.md](AGENTS.md)。

## License

MIT © [wadefengx](https://github.com/wadefengx)
