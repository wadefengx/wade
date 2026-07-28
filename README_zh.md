# Wade

> 🏄 一站式 Node.js 版本和镜像源管理工具。一次安装，到处使用。

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Wade** 用一个工具替代 **nvm** + **cgr/nrm**——管理 Node.js 版本、切换 npm/yarn/pnpm 镜像源、控制 Node 下载来源。

📖 [English](README.md) · 🌐 [Website](https://wadefengx.github.io/wade)

---

## 快速开始

```bash
wade -i          # 交互式设置向导（推荐！）
wade init -y     # 非交互式：全部用默认值
```

4 步向导：Node 下载源 → 安装版本 → 选择镜像 → PATH 配置

---

## 安装

```bash
git clone https://github.com/wadefengx/wade.git
cd wade
go build -o /usr/local/bin/wade .
wade -i           # 运行交互式设置
```

---

## 命令一览

| 命令 | 说明 |
|------|------|
| `wade -i` | 交互式设置向导（4 步） |
| `wade init` | 同上 |
| `wade init -y` | 非交互式，全部默认值 |
| `wade node mirror` | 查看/切换 Node 下载源（官方/镜像） |
| `wade node install 20` | 安装 Node 20 |
| `wade node use 20` | 切换到 Node 20 |
| `wade node ls` | 查看已安装版本 |
| `wade registry ls` | 查看所有镜像源 |
| `wade registry use taobao` | 切换 npm/yarn/pnpm 到淘宝 |
| `wade registry test` | 测速 |
| `wade status` | 查看当前状态 |

### Node 下载源

```bash
wade node mirror            # 查看当前
wade node mirror mirror     # 用 npmmirror.com（国内快，默认）
wade node mirror official   # 用 nodejs.org 官方
```

**这与 registry 不同！**

- `wade node mirror` — 控制 Node 二进制从哪下载
- `wade registry use` — 控制 npm install 从哪下载包

### 内置镜像源

| 名称 | URL |
|------|-----|
| `npm` | https://registry.npmjs.org/ |
| `taobao` | https://registry.npmmirror.com/ |
| `tencent` | https://mirrors.tencent.com/npm/ |
| `huawei` | https://repo.huaweicloud.com/repository/npm/ |
| `cnpm` | http://r.cnpmjs.org/ |

---

## License

MIT © [wadefengx](https://github.com/wadefengx)
