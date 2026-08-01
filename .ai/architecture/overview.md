# Architecture Overview

## 技术栈
- Go 1.25 + cobra(spf13)
- 配置: pelletier/go-toml/v2 (TOML)
- 交互: AlecAivazis/survey/v2 (init 向导)
- 版本解析: Masterminds/semver/v3
- 跨平台: GOOS/GOARCH 交叉编译, 无 CGO

## 模块

| 模块 | 路径 | 职责 |
|------|------|------|
| cmd | cmd/ | cobra 命令层: root/node/registry/status/init/go/update/setup |
| node | internal/node/ | Node 版本管理: install/use/ls/uninstall + shim + .wade-version |
| registry | internal/registry/ | npm/yarn/pnpm 注册表切换 + 测速 |
| go | internal/go/ | Go 版本管理 + 下载/解压/切换 |
| python | internal/python/ | pip 镜像 + GOPROXY + 系统检测 |
| config | internal/config/ | ~/.wade/config.toml 读写 |
| platform | internal/platform/ | Symlink 抽象(Windows 硬链接 fallback) |
| docs | docs/ | GitHub Pages 单文件站点(中英/明暗/复制按钮) |

## 数据流(核心场景)

```
wade node use 20
→ cmd/node.go → node.UseVersion("v20.x.x")
→ platform.Symlink 更新 ~/.wade/shims/{node,npm,npx}
→ 写 ~/.wade/current
→ (无参时) FindProjectVersion 向上找 .wade-version

wade registry use taobao
→ registry.Use → npm/yarn/pnpm config set registry
→ config.Save 更新 current_registry
```

## ADR
见 `.ai/architecture/adr/`(D1-D5 已记录于 `.ai/memory/decisions/`)
