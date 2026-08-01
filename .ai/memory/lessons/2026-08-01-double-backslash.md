# 经验教训

## 2026-08-01 Windows PATH 双反斜杠

### L31: PowerShell 单引号字符串里的 `\\` 就是两个反斜杠, 不是转义
setup.go 用 `strings.ReplaceAll(shimAbs, \`\\\`, \`\\\\\`)` 把 `C:\Users\...` 转义成 `C:\\Users\\...` 再塞进 PowerShell 命令——但 PowerShell 单引号字符串不做转义,`\\` 就是两个字符,于是**双反斜杠路径写进了注册表**。
**后果**: Windows 文件系统解析 `C:\\Users\\wade\\.wade\\shims` 正常(双反斜杠等价单反斜杠),`where node` 能找到 shim——但 Go 的字符串比较(单反斜杠)永远不匹配 → `wade status` 误报 "NOT on PATH"。
**教训**: 往 PowerShell 单引号字符串插路径**不要转义反斜杠**。修复: 不转义 + 比较前 `normalizePath` 折叠双反斜杠。

### L32: "误报"比"真 bug"更难排查
用户 `wade setup --auto` 显示 Moved to FRONT,`where node` 显示 shims\node.exe 第一——功能已生效,但 status 一直警告。查了 4 轮(窗口没重开 → nvm 遮蔽 → explorer 没刷新)才发现是字符串比较问题。
**教训**: 诊断要先用 `where node`/`echo %PATH%` 看**实际解析结果**,再信工具的自我检测。状态提示的"false negative"会严重误导用户。

### L33: 用户环境有 nvm 双保险
用户 Windows + Mac 都装过 nvm。PATH 里 nvm 的 symlink 目录(`AppData\Local\nvm\nodejs`)在 shims 之后——`where node` 显示 shims 第一,所以 wade 赢。但如果 nvm 目录在 shims 前就会遮蔽。
**教训**: 做 Node 版本管理器必须考虑 nvm 共存场景,status 检测 node 实际解析路径(whichNode)而不是只看 PATH 包含。
