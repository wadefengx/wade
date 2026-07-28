# Wade

> 🏄 一站式 Node.js 版本和 registry 管理工具。一次安装，到处使用。

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey)](https://github.com/wadefengx/wade/releases)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Wade** 用一个工具替代 **nvm** + **cgr/nrm**——管理 Node.js 版本，同时一键切换 npm/yarn/pnpm 的 registry。

📖 [English](README.md) · 🌐 [Website](https://wadefengx.github.io/wade)

---

## 为什么需要 Wade？

### nvm 的痛点

- 每个 Node 版本都有独立的全局包目录，切换到 v16 后 v14 里装的 `vite`、`cgr` 等工具全都不见了
- 每次切版本都要 `npm install -g` 重新装一遍全局包
- shell 启动时加载 nvm 会拖慢终端，不懒加载又慢

### cgr/nrm 的痛点

- 通过 `npm install -g cgr` 安装，本质上是某个 Node 版本下的全局包
- 切 Node 版本后就找不到 `cgr` 命令
- 只管理 npm registry，不管 yarn 和 pnpm

### Wade 怎么解决

| 问题 | nvm | cgr/nrm | **Wade** |
|------|-----|---------|----------|
| Node 版本管理 | ✅ | ❌ | ✅ |
| Registry 切换 | ❌ | ✅ (仅 npm) | ✅ (npm + yarn + pnpm) |
| 全局安装，切版本不丢失 | ❌ | ❌ (绑定了 Node 版本) | ✅ |
| 跨平台 (macOS + Windows) | ✅ | ❌ | ✅ |
| 国内镜像加速下载 | ❌ | — | ✅ (默认 npmmirror) |

**核心理念：** Wade 是一个独立的 Go 二进制文件，不依赖 Node.js。安装一次，切版本不受影响。Node 安装包从国内镜像下载，速度飞快。

---

## 安装

### 源码编译（全平台，需要 Go 1.23+）

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o /usr/local/bin/wade .
# 或者放到 ~/.local/bin/ 下
go build -o ~/.local/bin/wade .
wade setup
```

### macOS（Homebrew — 即将支持）

```bash
brew install wadefengx/tap/wade
```

### Windows（Scoop — 即将支持）

```powershell
scoop bucket add wade https://github.com/wadefengx/scoop-wade
scoop install wade
```

### 安装后配置 PATH

将下面这行加到你的 shell 配置文件中（`~/.zshrc` 或 `~/.bashrc`）：

```bash
export PATH="$HOME/.wade/shims:$PATH"
```

然后重启终端，或者执行 `source ~/.zshrc`。

---

## 快速开始

```bash
# ── Registry 管理（不需要安装 Node 就能用）──
wade registry ls                 # 列出所有可用源
wade registry use taobao         # 一键切换 npm/yarn/pnpm 到淘宝镜像
wade registry test               # 测试所有源的响应速度
wade registry add corp https://npm.mycompany.com/  # 添加自定义源

# ── Node 版本管理 ──
wade node install 20             # 安装 Node 20（从 npmmirror.com 下载）
wade node use 20                 # 切换到 Node 20
wade node ls                     # 查看已安装的版本
wade node ls-remote              # 查看所有可用版本
wade node default 20             # 设为默认版本
wade node uninstall 18           # 卸载某个版本

# ── 状态查看 ──
wade status                      # 查看当前环境状态
```

---

## 命令详解

### `wade registry` — Registry 管理

```bash
wade registry ls          # 列出所有 registry（内置 5 个 + 自定义的）
wade registry use <名称>   # 切换 npm/yarn/pnpm 到指定源
wade registry add <名称> <URL>  # 添加自定义 registry
wade registry del <名称>   # 删除自定义 registry
wade registry test        # 测试所有 registry 的响应延迟
```

内置 registry：

| 名称 | URL | 说明 |
|------|-----|------|
| `npm` | https://registry.npmjs.org/ | 官方源 |
| `taobao` | https://registry.npmmirror.com/ | 淘宝镜像（推荐国内使用） |
| `tencent` | https://mirrors.tencent.com/npm/ | 腾讯云镜像 |
| `huawei` | https://repo.huaweicloud.com/repository/npm/ | 华为云镜像 |
| `cnpm` | http://r.cnpmjs.org/ | CNPM 源 |

### `wade node` — Node 版本管理

```bash
wade node install <版本>   # 安装指定版本，支持 "18"、"18.20"、"lts"、"latest"
wade node use <版本>       # 切换到指定版本
wade node ls              # 查看已安装版本
wade node ls-remote       # 查看远程可用版本
wade node default <版本>   # 设置默认版本
wade node uninstall <版本> # 卸载某个版本
wade node current         # 查看当前使用的版本
```

### `wade status` — 环境状态

```
$ wade status
────────────
  Registry:   taobao (https://registry.npmmirror.com/)
  Node ver:   v20.20.2 (default)
  Config:     ~/.wade/config.toml
```

---

## 工作原理

```
wade（单个 Go 二进制，无需运行时依赖）
    │
    ├── ~/.wade/versions/    ← 下载的 Node 二进制文件
    ├── ~/.wade/shims/       ← PATH 中的软链接（只需设置一次）
    ├── ~/.wade/config.toml  ← 用户配置
    └── ~/.wade/current      ← 当前使用的版本
```

- **Registry 切换：** 同时写入 `npm config`、`yarn config`、`pnpm config`
- **Node 版本切换：** 更新 `~/.wade/shims/` 中的软链接，瞬间完成，无需 shell reload
- **不依赖 Node：** Wade 本身不需要 Node；安装 Node 时从 npmmirror.com 下载预编译包

---

## 开发

Wade 使用 **Go** 编写，遵循 **Spec-Driven Development (SDD)** 方法论。

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o wade .
./wade status
```

### 项目结构

```
wade/
├── AGENTS.md              # AI 主控上下文
├── spec/SPEC.md           # 完整规范文档
├── cmd/                   # CLI 命令（cobra 框架）
├── internal/
│   ├── config/            # TOML 配置管理
│   ├── registry/          # Registry 切换逻辑
│   └── node/              # Node 版本管理
└── docs/                  # GitHub Pages 站点
```

### 开发路线

- [x] **M1: Registry 管理** — `registry ls/use/add/del/test`
- [x] **M2: Node 版本管理** — `node install/use/ls/default/uninstall`
- [ ] **M3: 发布** — GitHub Actions 自动构建 + Homebrew + Scoop + 安装脚本
- [ ] **M4: 完善** — Shell 补全、自动更新、Windows 支持

---

## License

MIT © [wadefengx](https://github.com/wadefengx)
