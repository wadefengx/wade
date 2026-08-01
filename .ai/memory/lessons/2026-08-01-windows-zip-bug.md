# 经验教训

## 2026-08-01 Windows 安装体验(续)

### L12: Windows 资产是 zip, 解压逻辑必须按平台分流
`wade node install` 在 Windows 下载 `.zip` 但代码硬编码 tar.gz 解压 → `gzip: invalid header`。
Go 官方 Windows 资产同样是 zip(`go1.x.windows-amd64.zip`),node 和 go 两个 manager 都有此 bug。
**教训**: 跨平台工具的"下载→解压"链路必须按扩展名分流(zip.OpenReader vs tar+gzip),且每个平台格式要实测(不能只测 macOS)。
**验证**: 用真实 node-v20.20.2-win-x64.zip 写回归测试,零网络依赖。

### L13: 用户要求"update 命令自动更新"
用户希望 `wade node update` / `-u` 不用手动查最新版本号,直接更新已装版本。
实现: 带参更新指定 major 线, 无参更新全部已装版本(先 resolve 再比版本号, 不同才装)。
**教训**: 工具类 CLI 要提供"无脑更新"路径, 别让用户手动 ls-remote + install。

### L14: scoop 安装链路要非交互
`post_install: "wade.exe setup"` 交互式会卡死 scoop 安装 → 必须 `setup --auto`。
