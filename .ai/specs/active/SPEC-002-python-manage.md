# SPEC: Python 版本全权接管 + init 运行时安装

## 背景

用户要求 wade **全权接管 node/go/python 的安装、版本更新、镜像、换源**。
现状:
- Node: ✅ 全托管(install/use/ls/ls-remote/mirror)
- Go: ✅ 可托管(install/use/ls/ls-remote/mirror)
- **Python: ❌ 只有系统检测 + pip 换源,没有版本安装/切换**

另外 `wade init` 交互式选 All 后,Go/Python 分支静默(无反馈),v0.5.5 已修交互反馈。

## 目标

1. `wade python install <version>` — 安装 Python 版本(便携发行版)
2. `wade python use <version>` — 切换当前 Python(shim)
3. `wade python ls` — 已安装版本(保留系统检测展示)
4. `wade python ls-remote` — 可用版本列表
5. `wade python default <version>` — 设默认并切换
6. `wade python uninstall <version>` — 删除版本
7. init 选 All 时:询问是否安装 Python(默认装最新 3.12/3.13)

## 下载源:python-build-standalone

- Repo: `astral-sh/python-build-standalone`(uv 同源,便携、含 pip)
- API: `https://api.github.com/repos/astral-sh/python-build-standalone/releases/latest`
- 资产名: `cpython-<ver>+<builddate>-<platform>-install_only.tar.gz`
  - macOS arm64: `aarch64-apple-darwin`
  - macOS x64: `x86_64-apple-darwin`
  - Windows amd64: `x86_64-pc-windows-msvc`
  - Linux amd64: `x86_64-unknown-linux-gnu`
- 解压后顶层是 `python/` 目录,里面是完整 Python(bin/python3, lib/, include/)
- 注意版本格式: `3.11.15+20260728`(版本+构建日期),用户输入 `3.11` 或 `3.11.15` 要解析匹配

## 实现细节

### internal/python/manager.go 新增

```go
// PythonBuild represents a python-build-standalone release
type PythonBuild struct {
    Version string // "3.11.15"
    Asset   string // full asset filename
}

// FetchRemoteVersions() ([]PythonBuild, error) — 拉 releases/latest,筛 cpython-*.install_only.tar.gz,
// 提取版本号,按 semver 降序。用双通道: api.github.com 优先,失败提示。
// 资产 URL 用 releases/assets/{id}(CN 可达 CDN),同 update.go 的 downloadAsset 模式。

// PlatformSuffix() string — 返回当前平台的 python-build-standalone 后缀
// (runtime.GOOS/GOARCH → aarch64-apple-darwin 等)

// Install(version string) error:
//   1. ResolveVersion(partial → full, e.g. "3.11" → "3.11.15")
//   2. 下载 cpython-<ver>+<build>-<suffix>-install_only.tar.gz(双通道: API asset → CDN, github.com fallback)
//   3. 解压到 ~/.wade/python/versions/<ver>/ (剥离顶层 python/ 目录)
//   4. 自动 UseVersion(激活)

// UseVersion(version) error — 建 shim:
//   shims/python3 → <ver>/bin/python3 (unix symlink / windows hardlink)
//   shims/pip3 → <ver>/bin/pip3 (若存在)
//   Windows 注意: python.exe 不需要 wrapper(不依赖自身路径,GOROOT 问题只属于 Go)
//   写 ~/.wade/python/current

// InstalledVersions() ([]string, error) — 列 ~/.wade/python/versions/
// CurrentVersion() (string, error) — 读 current 文件
// Uninstall(version) error — 删版本目录
```

### cmd/go.go 的 Python 命令区新增(cmd/python.go 更合适)

```go
pythonInstallCmd  — wade python install 3.11
pythonUseCmd      — wade python use 3.11 (支持部分匹配已安装)
pythonLsCmd       — 已装 + (system) 检测(保留现有逻辑)
pythonLsRemoteCmd — wade python ls-remote
pythonDefaultCmd  — wade python default 3.11 (设默认+切换,同 node)
pythonUninstallCmd— wade python uninstall 3.11
```

### init.go 的 Python 分支增强

选 All 或 Python 时,检测 ~/.wade/python/versions/ 为空 → 询问:
```
◇  Install a Python version? [3.12 — recommended / 3.11 / Skip]
```
(autoYes 时直接装 3.12)。装完 UseVersion。

### status.go

Python 行:优先显示 wade 托管的版本(CurrentVersion),没有才显示 system 检测。

## 验证

- `go test ./internal/python/... ./cmd/...` 全绿
- `GOOS=windows go build ./...` 通过
- 真实下载小测试(3.10 install_only,~40MB)验证解压结构
- 手动: wade python install 3.11 → wade python use 3.11 → python3 --version

## 不做(YAGNI)

- 不管理 pipenv/poetry/venv
- 不做 Python 构建(只用预编译)
- 不自动卸载系统 Python
