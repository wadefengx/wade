# Changelog

All notable changes to wade. This file is designed to be **AI-friendly** — structured so that Copilot, Claude, Codex, and other AI tools can quickly understand the project's evolution and current state.

---

## [v0.5.5] — 2026-08-04

### Added

- **`wade python` 版本全权接管**(SPEC-002): `install/use/ls/ls-remote/default/uninstall`
  - 下载源: astral-sh/python-build-standalone(uv 同源,便携含 pip,双通道下载)
  - `wade python install 3.11` → 解析最新 3.11.x → 下载解压 → 自动激活
  - `wade python use 3.11` → 切 shim(python3/pip3),支持部分版本匹配
  - `wade python ls` → 托管版本 + 系统检测;`ls-remote` → 平台可用版本
  - `wade python default` → 设默认并切换(同 node)
  - `wade init` 选 All/autoYes → 自动安装 Python 3.12
- **`wade init` 交互式 Go/Python 分支反馈**: 设置 mirror/proxy/pip 后打印 `✓ Go mirror/Go proxy/pip mirror`(之前静默)

---

## [v0.5.4] — 2026-08-04

### Fixed

- **`wade init` 选 All 后只显示 Node,Go/Python 无反馈**: Go/Python 分支实际执行了(mirror/proxy/pip mirror),但**完全静默**且 summary 不显示——用户误以为没配置
  - autoYes(init -y)时打印 `✓ Go mirror: google-cn` / `✓ Go proxy: goproxy.cn` / `✓ pip mirror: tsinghua`
  - summary 增加 Go(版本 + mirror)、Python 行,未托管时明确标注 `(system — not managed by wade)`
  - 新增真实 config 读写测试(registry + go mirror 持久化)

---

## [v0.5.3] — 2026-08-01

### Fixed

- **命令一致性(node/go/python 统一结构)**: `wade go install/use` 现在完全对齐 `wade node` 的 UX——
  - `wade go use 1.23` → 自动匹配已安装的 `go1.23.x`(之前只做字符串拼接,`use 1.23` 报错)
  - `wade go install 1.23.2` 和 `go1.23.2` 都接受(ResolveVersion 自动补全,无需记忆前缀)
  - 错误提示统一不带 `go` 前缀(`not installed — run 'wade go install 1.23'`)
- **Go 1.21+ GOROOT 问题(Windows)**: 官方 go 二进制是 trimmed 的,靠自身可执行路径推断 GOROOT——硬链接 shim 导致 `go: cannot find GOROOT directory`。Windows 改用 **`.cmd` wrapper**(set GOROOT + 真实路径调用),Unix 保持 symlink(可被 os.Executable 解析)

---

## [v0.5.2] — 2026-08-01

### Added

- **`wade status` Go shim 健康检查**: 之前 status 只读 current 文件(显示 go1.23.12)但从不验证 `go` 命令是否真的解析到 wade shim——造成"已激活"错觉。现在检测: ① `go` 解析到系统/nvm → 警告 + `wade go use` 提示 ② `go` 完全不在 PATH → 警告 + `wade setup` 提示(与 node 的 shim 检查对称)

---

## [v0.5.1] — 2026-08-01

### Fixed

- **`wade go use` 后 `go` 仍找不到(v0.5.0 回归)**: v0.5.0 把 Windows shim 名写成无扩展名 `go`——但 cmd/PowerShell 的 **PATHEXT 只匹配 `.exe/.cmd/.bat`**,无扩展名文件永远不会被找到。node shim 一直叫 `node.exe`(保留扩展名)所以能用。**修复**: go shim 在 Windows 保留 `.exe`(`shims/go.exe` → `versions/.../bin/go.exe`)

---

## [v0.5.0] — 2026-08-01

### Fixed

- **Windows 安装 Go 后 `go` 命令找不到**: UseVersion 写死 `bin/go`(无扩展名),Windows 上是 `go.exe` → os.Stat 失败 → **shim 从未创建**(和 node 的 `node.exe` 同类 bug)。现在 Windows 查找 `go.exe`/`gofmt.exe`,shim 名保持无扩展名(`shims/go` → go.exe,cmd 的 PATHEXT 解析)
- **`wade go install` 后自动激活**: 和 `wade node install` 行为一致,装完立即 `use`,不用再单独 `wade go use`
- 新增 Windows shim 布局测试

---

## [v0.4.9] — 2026-08-01

### Fixed

- **`wade go install 1.23` → 404**(根因三连):
  1. **无版本解析**: `1.23` 直接拼 `go1.23.windows-amd64.zip`(需 `go1.23.x`)。新增 `ResolveVersion`(精确版本直通,模糊版本查 API 找最新 patch)
  2. **go.dev JSON API 只返回最近两个大版本**(1.26/1.25),`1.23` 查不到。改用 `?mode=json&include=all`(357 个版本,实测 `1.23 → go1.23.12`)
  3. **`wade go mirror aliyun` 从未生效**: cobra 里 `aliyun` 不是子命令(`use` 才是),未知参数回落父命令只显示当前源。`goMirrorCmd` 改为有参数即切换(隐式 use)
- **移除失效的 Go 镜像预设**: npmmirror/aliyun 的 Go 镜像已死(实测 404,两种路径都试过),presets 只留 official + google-cn(均实测 200)。pip 镜像全部正常(tsinghua/aliyun/huawei/tencent/ustc 均 200),python 无此问题
- `FetchRemoteVersions` 加 15s 超时
- 新增测试: 精确/模糊/无匹配 + 真实 google-cn API 集成测试

---

## [v0.4.8] — 2026-08-01

### Changed

- **Windows 安装收敛为单一途径**: 移除 Scoop 安装入口,Windows 只保留一行 PowerShell 安装(`irm ... | iex`)。README/README_zh/官网安装卡片/spec 全部同步(删 Scoop 卡片与 i18n key)。scoop-wade bucket 与 release 自动更新保留,不再作为推荐途径

---

## [v0.4.7] — 2026-08-01

### Fixed

- **`wade node default` 只改 config 不切 shim**: `default <version>` 之前只写 `config.toml` 的 DefaultVersion,shim 仍指向 current → `node -v` 不变,用户困惑。现在 `default` 同时调用 `UseVersion`(切换 shim + current 文件),**立即生效**(与 nvm alias default 行为一致)

---

## [v0.4.6] — 2026-08-01

### Fixed

- **`wade status` 误报 "shims NOT on PATH"(实际已生效)**: setup 的 PowerShell 命令里 `strings.ReplaceAll(shimAbs, \`\\\`, \`\\\\\`)` 把路径转义成双反斜杠写入注册表(`C:\\Users\\...`)——Windows 文件系统解析容忍它(`where node` 能找到 shim),但 Go 的字符串比较永远匹配不上 → status 误报
  - setup 不再转义(写单反斜杠),并在写入前自动把已有双反斜杠归一化
  - `pathInEnvPath` / `userPathHasShims` 比较前 `normalizePath` 折叠双反斜杠
  - `wade setup --auto` 重跑一次即可修复已污染的注册表

---

## [v0.4.5] — 2026-08-01

### Fixed

- **交互式 `wade -i` 在 Windows 上仍写 shell rc 文件**(bash 语法 `export PATH=...` 写入 PowerShell profile——PowerShell 不认,cmd 不读,完全无效)。现在 Windows 上 `wade -i` 的 PATH 步骤改调 `runSetup()`(shims 加到用户 PATH 最前,cmd + PowerShell 都生效)
- **`wade status` 诊断增强**: ① shims 在 PATH 但 node 解析到系统/nvm → 明确提示 nvm 遮蔽 ② 用户 PATH 有 shims 但当前窗口旧 → 提示关闭所有窗口重开 + `setx` 方案 ③ 完全没有 → 提示 setup

---

## [v0.4.4] — 2026-08-01

### Added

- **Windows 系统代理自动识别**: `wade -u`/启动检查现在自动读取 Windows 系统代理(注册表 WinINET 设置,和 PowerShell 的 irm 一致)——之前 Go 只认 HTTP_PROXY 环境变量,ClashX 系统代理对 wade 无效,导致 PowerShell 里更新快、cmd 里卡。优先级:环境变量 > Windows 系统代理
- `parseProxyServerLine` 支持纯地址 / http:// 前缀 / 分协议列表(https 优先)+ 单测

---

## [v0.4.3] — 2026-08-01

### Fixed

- **`wade -u` 在 CN 网络超时** (`context deadline exceeded`): update.go 只有 github.com redirect 单通道,但用户网络 api.github.com 通、github.com 被墙
  - 版本解析双通道: API 优先(CN 可达)→ redirect 兜底
  - 下载双通道: API 资产端点 `releases/assets/{id}`(→ CDN,CN 可达)→ github.com 直连兜底
  - checksum 同样双通道
  - 新增 `assetIDByName`(GitHub 资产 JSON 中 `"id"` 在 `"name"` 前,向后找会命中 uploader 的假 id)+ 单测

---

## [v0.4.2] — 2026-08-01

### Added

- **`wade init -y` 自动配置 PATH**: 傻瓜式一条命令——`init -y` 现在在配置完 runtimes/镜像后自动执行 setup(--auto 语义),Windows 加用户 PATH,Unix 加 shell rc
- **`wade -u` 下载超时 + 代理提示**: 60s 超时(之前 `http.Get` 无超时,github.com 被墙时 cmd 里干等),失败时提示设置 `HTTP_PROXY`(PowerShell 快是因为 profile 里有代理,cmd 没有)

---

## [v0.4.1] — 2026-08-01

### Fixed

- **Windows shims 加到 PATH 末尾导致 node 仍解析到系统版**: Windows 的 PATH 顺序 = 系统 PATH(含 Program Files\nodejs)在前 + 用户 PATH 在后。旧 setup 把 shims 追加到用户 PATH 末尾 → node 永远先命中系统版
  - `wade setup` 现在把 shims 加到用户 PATH **最前面**,已存在但顺序不对时自动 reorder
  - `wade status` / `wade node use` 分三层诊断: ① 当前会话有 shims → 检查 node 实际解析路径(是否被系统版遮蔽) ② 用户 PATH 有但当前窗口旧 → 提示重开窗口 ③ 完全没有 → 提示 setup
  - `userPathHasShims`(读注册表)+ `whichNode`(exec.LookPath)helper

---

## [v0.4.0] — 2026-08-01

### Added

- **`wade status` 显示 wade 版本**: 第一行 `🏄 Wade: v0.4.0`
- **`wade status` shim 健康检查**: 检测 `~/.wade/shims` 是否在 PATH,不在则警告 `⚠️ ~/.wade/shims is NOT on your PATH — 'node' is the system version, not wade's!` + 修复指引(`wade setup --auto` + 重开终端)——直接定位"node -v 还是系统版本"问题

---

## [v0.3.9] — 2026-08-01

### Fixed

- **Windows `node use` 后 `node -v` 还是系统版本**: shims 目录不在 PATH。原 setup 只写 PowerShell profile(且用的是 bash 语法 `export PATH=...`,PowerShell 不认),cmd 用户完全无效
  - `wade setup` 在 Windows 上改为把 shims 加入**用户 PATH 环境变量**(注册表持久化,cmd + PowerShell 都生效),失败时给手动指引
  - `wade node use` 后检测 shims 不在 PATH → 明确警告 `⚠️ ~/.wade/shims is NOT on your PATH — 'node' will still be the system version!` + 提示 `wade setup`
  - `pathInEnvPath` 平台感知的 PATH 检测 + 单元测试

---

## [v0.3.8] — 2026-08-01

### Fixed

- **Windows `wade node use` 失败**: shim 逻辑写死 Unix 路径 `versions/<v>/bin/node`,但 Windows 的 Node 结构是 `node.exe`/`npm.cmd`/`npx.cmd` 在**版本根目录**(无 bin/)。symlink 失败 fallback 到硬链接时源路径不存在 → "system cannot find the path specified"
  - `shimTargets()` 按平台返回正确的 shim 名 + 目标路径(Windows: node.exe/npm.cmd/npx.cmd @ root;Unix: bin/node 等)
  - yarn/pnpm 同样平台分流(yarn.cmd/pnpm.cmd @ root)
  - 单元测试: Windows + Unix 两条路径分支

---

## [v0.3.7] — 2026-08-01

### Fixed

- **install.ps1 checksum 校验被跳过**: PS 5.1 的 `Invoke-WebRequest` 对 `application/octet-stream` 返回 `byte[]`,`.Trim()` 调用失败 → 校验静默跳过。新增 `Get-ContentString` helper,byte[] 先 UTF8 解码再解析

---

## [v0.3.6] — 2026-08-01

### Fixed

- **CN 网络下 install.ps1 拿不到版本号**: CN 网络特征 = `api.github.com` 可达、`github.com:443` 被墙。v0.3.4 改成 github.com 重定向后 CN 用户直接失败。恢复**双通道**: API 优先(CN 可达),重定向兜底(海外)
- **下载也双通道 + 自动重试**: API 资产端点 `releases/assets/{id}`(→ CDN,CN 可达)重试 2 次 → github.com 直连第 3 次 → 全部失败给出明确原因 + 代理指引
- **checksum** 同样 API 资产优先

### Added

- **自动重试机制**: 下载最多 3 次(API×2 + github.com×1),失败在 cmd 里明确显示原因和代理方案

---

## [v0.3.5] — 2026-08-01

### Fixed

- **PowerShell 闪退**: `irm | iex` 管道下,install.ps1 里的 `exit` 会关闭整个 PowerShell 会话(错误不可见)。重构:所有 `exit` → `throw`,主体包 try/catch,失败时 `Read-Host` 暂停让窗口停留显示错误

---

## [v0.3.4] — 2026-08-01

### Fixed

- **GitHub API rate limit broke install/update** (`API rate limit exceeded` on shared CN IPs): all version lookups now use the `releases/latest` HTTP **redirect Location** (`302 → /releases/tag/vX.Y.Z`) instead of `api.github.com` — zero API quota
  - `scripts/install.ps1`: redirect-based version resolve + `latest/download` checksum fetch; fixed sha asset name (`wade-windows-amd64.sha256`, was wrongly `xxx.zip.sha256` which silently skipped verification)
  - `scripts/install.sh`: `curl -sfIL` Location extraction
  - `cmd/update.go`: `CheckRedirect` → `ErrUseLastResponse`, parse Location (also used by `wade -u` / startup check)

---

## [v0.3.3] — 2026-08-01

### Added

- **`wade -u` shortcut**: 与 `wade update` 等效,一条命令自更新
- **启动版本检查**(oh-my-zsh 风格): 每次运行命令前检查 GitHub 最新版(24h 缓存,3s 超时,静默失败),发现新版本打印 `✨ New version available` 并交互询问 `Update now? [y/N]`;非交互环境只提示不阻塞

---

## [v0.3.2] — 2026-08-01

### Fixed

- **Windows `wade node install` → `gzip: invalid header`**: Windows 下载的是 `.zip` 但代码硬编码 tar.gz 解压。新增 `extractArchive` 按扩展名分流(zip/tar.gz),Node + Go 管理器同修(Go 的 Windows 资产也是 zip)。用真实 node-v20.20.2-win-x64.zip 回归测试验证。

### Added

- **`wade node update [version]`**: 更新已装 Node 版本到最新——带参更新指定 major 线,无参更新全部已装版本

---

## [v0.3.1] — 2026-08-01

### Added

- **Windows one-line installer** (`install.ps1`): `irm .../install.ps1 | iex` — 下载 zip + SHA256 校验 + 解压到 %LOCALAPPDATA%\wade + 自动加用户 PATH,无需 Scoop、cmd/PowerShell 通用
- Scoop manifest `post_install` 改为非交互 `setup --auto`(原交互式会卡死 scoop 安装)

---

## [v0.3.0] — 2026-07-31

### Added

- **`.wade-version` project pinning**: `wade node use` (no arg) auto-detects the version file walking cwd → home, like nvm's `.nvmrc`
- **`internal/platform/` layer**: `Symlink()` abstraction with Windows hard-link fallback (fixes shim creation failing without admin/Developer Mode)
- **CI quality gate** (`.github/workflows/ci.yml`): gofmt/vet/test/build on push + PR
- **cmd/ unit tests**: checksum verification + archive extraction (cmd coverage 0 → 10%)
- **update flow hardened**: SHA256 verification + real extraction (was installing raw archive)

### Fixed

- `wade update` installed the raw `.tar.gz` archive as the binary (never extracted) — now extracts and verifies checksum against the real `wade-<os>-<arch>.sha256` asset names
- Shim creation error swallowing (`os.Remove`/`os.Symlink` failures now propagate)

---

## [v0.2.1] — 2026-07-31

### Added

- **M8 Distribution complete**: Homebrew tap `wadefengx/tap` + Scoop bucket `wadefengx/scoop-wade` live with production formula/manifest (real SHA256 digests)
- **GitHub Pages site** live at https://wadefengx.github.io/wade/ (multi-runtime features, install guide, command cheatsheet, milestone status)
- **release.yml version injection**: `-X main.version=<tag>` via `main.version` var
- **release.yml update-tap job**: auto-updates tap/bucket manifests on release via `scripts/release-shas.sh --update` (needs `TAP_TOKEN` secret)
- **install.sh now uploaded** as release asset (fixes `releases/latest/download/install.sh` 404)
- **scripts/release-shas.sh**: fetch latest release digests, print or `--update` templates

### Fixed

- `update-tap` job used dead sed placeholders (templates were hardcoded) → now uses release-shas.sh
- `docs/index.html` refreshed for multi-runtime + new commands (M0-M8 status table)

---

## [v0.2.0] — 2026-07-28

### Added

- **Interactive setup wizard** (`wade -i` / `wade init`)
  - 3-step flow: Node mirror → version → PATH
  - `wade init -y` for non-interactive auto-config
  - `wade init` (no flags) writes `.wade-version`
- **Node download mirror** (`wade node mirror`)
  - `wade node mirror official` → nodejs.org
  - `wade node mirror mirror` → npmmirror.com (default)
  - Separated from registry management — two independent concepts
- **Emoji-enhanced CLI output** for liveliness
- **GitHub Actions CI**: cross-compile darwin/arm64, darwin/amd64, linux/amd64, windows/amd64
- **Release templates**: Homebrew formula, Scoop manifest, install script
- **Self-update** (`wade update`)
- **Shell completions** via cobra built-in
- **GitHub Pages** landing page (Apple dark theme)
- **CHANGELOG.md** (this file)

### Changed

- `--version` flag now works correctly (ldflags injection)
- Registry switching tolerates pnpm failures on older Node versions
- `wade status` output simplified with helpful tips
- README rewritten with cross-platform install guide (pre-built + brew + scoop + curl)

### Fixed

- Stale binary at `~/.local/bin` shadowing updated binary at `/usr/local/bin`
- pnpm v11 crash on Node v20 (now shows warning, doesn't block)
- `wade registry ls` table rendering for wide URLs

---

## [v0.1.0] — 2026-07-28

### Added

- **Node version management** (`wade node`)
  - `install`, `use`, `ls`, `ls-remote`, `default`, `uninstall`, `current`
  - Shim-based switching via `~/.wade/shims/`
  - Mirror download from npmmirror.com
  - Partial version resolution (e.g., `20` → `v20.20.2`)
- **Registry management** (`wade registry`)
  - `ls`, `use`, `add`, `del`, `test`
  - 5 built-in registries: npm, taobao, tencent, huawei, cnpm
  - Switches npm + yarn + pnpm simultaneously
- **Core infrastructure**
  - Go module with cobra CLI framework
  - TOML config at `~/.wade/config.toml`
  - `wade setup` for directory creation + PATH hint
- **SDD foundation**
  - `AGENTS.md` — AI master context
  - `spec/SPEC.md` — complete specification
  - `.hermes/plans/` — implementation plan

---

## Architecture Notes (for AI tools)

### Key design decisions

1. **Shim-based switching** — `~/.wade/shims/` contains symlinks to current Node binaries. Set PATH once, switching is instant.
2. **No Node.js dependency** — Wade is a compiled Go binary. Full `http`, `os/exec`, `archive/tar` in stdlib.
3. **Registry vs Mirror** — `wade registry use` controls npm/yarn/pnpm package registries. `wade node mirror` controls where Node.js binaries are downloaded from. These are completely separate.
4. **Per-PM error tolerance** — If pnpm fails (e.g., wrong Node version), npm and yarn still succeed. Config is saved if at least one PM succeeds.

### File structure

```
wade/
├── main.go                 # Entry point, ldflags version injection
├── cmd/                    # CLI commands (cobra)
│   ├── root.go             # Root command, -i shortcut, version flag
│   ├── init.go             # Interactive setup wizard
│   ├── node.go             # Node version management + mirror
│   ├── registry.go         # Registry management + table render
│   ├── status.go           # Status dashboard
│   ├── update.go           # Self-update from GitHub Releases
│   └── setup.go            # Directory creation + shell detection
├── internal/
│   ├── config/config.go    # TOML config load/save
│   ├── registry/           # Registry switching logic
│   │   ├── presets.go      # 5 built-in registries
│   │   ├── manager.go      # Use/Add/Remove + per-PM exec
│   │   └── tester.go       # Concurrent latency testing
│   └── node/
│       ├── versions.go     # Version parsing (semver) + remote index
│       ├── manager.go      # Download + tar.gz extraction
│       └── shim.go         # Symlink management for PATH
├── AGENTS.md               # AI master context
├── spec/SPEC.md            # Complete specification
├── CHANGELOG.md            # This file
├── README.md + README_zh.md  # User docs
├── docs/index.html         # GitHub Pages
├── .github/workflows/      # CI/CD
└── scripts/                # Homebrew, Scoop, install.sh
```
