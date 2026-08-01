# 经验教训

## 2026-08-01 Windows shim 路径(续)

### L23: Windows Node 结构 = node.exe @ 根目录, 没有 bin/
Node win-x64 zip 解压后: `node.exe`/`npm.cmd`/`npx.cmd` 全在**版本根目录**(`node-v20.20.2-win-x64/`),`bin/` 只存在于 node_modules/npm/bin(内部)。Unix 才是 `bin/node`。
**教训**: 跨平台路径假设必须按平台分流。shim 名也要带扩展名:Windows 用 `node.exe`/`npm.cmd`(cmd 的 PATHEXT 不补 .cmd)。
**验证**: 用真实 zip 的 namelist 确认布局,别猜。已抽 `shimTargets()` 纯函数 + 双分支单测。

### L24: GOOS=windows 交叉编译的测试不能在 Mac 跑
`GOOS=windows go test` 生成 .exe,"exec format error" 是**环境限制**(Mac 不能执行 PE),不是代码错。用 `GOOS=windows go vet` 验证编译正确性。
**教训**: 交叉验证用 `go vet`/`go build`,别用 `go test` 执行。
