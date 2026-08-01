# 经验教训

## 2026-08-01 Windows PowerShell 闪退

### L18: `irm | iex` 管道下的 `exit` = 杀进程
PowerShell 执行 `irm <url> | iex` 时,脚本在**当前会话**内联执行。脚本里任何 `exit` 语句(包括 `exit 1`)会直接终止整个 PowerShell 会话 → **窗口闪退,错误信息完全不可见**。
**修复**: 安装脚本一律不用 `exit`,用 `throw` + 顶层 try/catch;catch 里 `Read-Host` 暂停,让用户能看到失败原因再关窗口。
**教训**: 写 `irm | iex` 类型的安装脚本,`exit` 是禁区。社区惯例:失败用 throw/return,成功自然结束。

### L19: 用户终端命令被 markdown 渲染污染
用户粘贴的命令 `irm @url:https://... | iexh` 带上了 `@url:` 前缀和 typo(`iexh`)。
**教训**: 文档里的命令要给出**可直接复制**的纯文本格式;排查用户问题时先确认他实际输入了什么(typo 可能不是 bug)。
