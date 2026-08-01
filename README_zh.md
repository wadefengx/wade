# Wade

> 🏄 一站式运行时管理工具：Node.js · Go · Python。单二进制，零依赖。

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

Wade 用一个工具替代 **nvm + cgr/nrm + gvm + pyenv**。

📖 [English](README.md) · 🌐 [Website](https://wadefengx.github.io/wade)

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

### Windows — 一行命令安装（无需 Scoop）

```powershell
# PowerShell（Windows 自带 5.1+，cmd 里也能用）
irm https://github.com/wadefengx/wade/releases/latest/download/install.ps1 | iex
```

安装到 `%LOCALAPPDATA%\wade`，自动加入用户 PATH，校验 SHA256。装完**重开一个终端**，然后 `wade -i`。

### Windows — Scoop（如果你已在用 Scoop）

```powershell
scoop bucket add wade https://github.com/wadefengx/scoop-wade
scoop install wade
wade -i
```

> ⚠️ 还没装 Scoop？先装：`Set-ExecutionPolicy RemoteSigned -Scope CurrentUser; irm get.scoop.sh | iex`——或者直接用上面的一行命令安装。

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

> ⚡ **两个 `update` 别搞混：**
>
> | 命令 | 更新什么 |
> |------|---------|
> | `wade -u` / `wade update` | **wade 工具本身**——下载最新版 wade 二进制 |
> | `wade node update` | **Node.js 运行时**——刷新已安装的 Node 版本 |

| 命令 | 说明 |
|------|------|
| `wade -i` | 交互式设置（选 Node/Go/Python/All） |
| `wade init -y` | 非交互式，自动配置全部（中国友好） |
| `wade -u` / `wade update` | 更新 wade 工具本身 |
| `wade node install 20` | 安装 Node 20 |
| `wade node use 20` | 切换到 Node 20 |
| `wade node update` | 更新已装 Node 版本到最新 |
| `wade node mirror ls` | 选择 Node 下载源 |
| `wade registry use taobao` | npm/yarn/pnpm 切淘宝 |
| `wade go install 1.23` | 安装 Go 1.23 |
| `wade go use 1.23` | 切换到 Go 1.23 |
| `wade go mirror use google-cn` | Go 下载切国内镜像 |
| `wade go proxy use goproxy.cn` | Go 模块代理切国内 |
| `wade python ls` | 检测系统 Python |
| `wade python registry use tsinghua` | pip 切清华源 |
| `wade status` | 当前状态 |
| `wade registry test` | 镜像测速 |

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
