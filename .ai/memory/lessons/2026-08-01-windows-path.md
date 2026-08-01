# 经验教训

## 2026-08-01 Windows PATH(续)

### L25: Windows 加 PATH 必须用用户环境变量, 不是 shell profile
`wade init` 把 shims 写进 PowerShell profile 但: (1) cmd 不读 PowerShell profile; (2) 写入的 `export PATH=...` 是 bash 语法, PowerShell 也不认。结果 `wade node use` 显示成功但 `node -v` 还是系统版本。
**修复**: Windows 上 `wade setup` 用 `[Environment]::SetEnvironmentVariable('Path', ..., 'User')`(注册表持久化, cmd+PS 全生效)。跨 shell 的工具必须改用户级环境变量, 别依赖特定 shell 的 rc 文件。

### L26: PATH 检测要平台感知
`os.PathListSeparator` 在 Windows 是 `;`, Unix 是 `:`。测试用 Windows 路径 + Unix 分隔符会误判(`C:\...` 里的 `:` 会被切)。
**教训**: 跨平台 PATH 相关测试必须用当前平台的分隔符和路径风格构造数据。

### L27: 用户困惑点 = "成功了但没生效"
`wade node use` 成功但 node 还是系统版 —— 最迷惑用户的时刻。工具应在成功后主动检测环境是否真的生效(shims 在 PATH?),不在就明确警告 + 给修复命令。
**教训**: "成功"要验证副作用, 不是只打印成功消息。
