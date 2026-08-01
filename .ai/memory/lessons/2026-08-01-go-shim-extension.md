# 经验教训

## 2026-08-01 Go shim 扩展名

### L37: Windows shim 文件必须保留扩展名(go.exe), 不能去掉
cmd/PowerShell 执行 `go` 时按 **PATHEXT**(.COM;.EXE;.BAT;.CMD)找文件——无扩展名的 `shims/go` **永远不会被匹配**。
v0.5.0 错误地 `TrimSuffix(name, ".exe")` 生成无扩展名 shim → `wade go use` 成功但 `go` 找不到。
node shim 一直叫 `node.exe`(保留扩展名)所以 node 一直能用——**对照 node 就知道正确答案**。
**教训**: shim 文件名 = 目标可执行文件名(含扩展名)。PATHEXT 只帮你从 `go` 补成 `go.exe` 去找,**不会**反过来识别无扩展名文件。

### L38: 回归测试要在 Windows 语义下断言
TestUseVersionWindowsShims 之前的断言写死 `go`/`gofmt`(unix 名)——在 Windows 上跑会误判成功(其实 shim 名错了)。测试应按 runtime.GOOS 断言正确扩展名。
