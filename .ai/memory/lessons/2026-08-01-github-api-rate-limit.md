# 经验教训

## 2026-08-01 GitHub API 限流(续)

### L15: GitHub API 未认证限流会炸掉安装/更新链路
`api.github.com` 未认证 60 req/h/IP。国内共享出口 IP(运营商 NAT)极易撞限流,install.ps1 直接 `Invoke-RestMethod` 拿版本号 → `API rate limit exceeded`。
**根因**: 版本号根本不需要 API —— `https://github.com/REPO/releases/latest` 返回 **302 → /releases/tag/vX.Y.Z**,从 Location 头提取 tag 零配额。
**修复**: install.ps1(`MaximumRedirection 0` 抓 Location)、install.sh(`curl -sfIL` grep location)、update.go(`CheckRedirect → ErrUseLastResponse`)全部去 API 化。
**教训**: 拿"最新版本号"一律用 releases/latest 重定向,别用 api.github.com。下载文件用 `releases/latest/download/<file>` 直链(自动重定向到最新版)。

### L16: PowerShell 5.1 的 Invoke-WebRequest 对 302 会抛异常
`-MaximumRedirection 0` 时 302 是非 2xx,PS 5.1 抛异常,Location 在 `$_.Exception.Response.Headers['Location']`。PS 7+ 有 `-SkipHttpErrorCheck` 但 5.1 没有——用 try/catch 兼容。

### L17: release 资产 sha 名与 zip 名不同
资产是 `wade-windows-amd64.sha256`(无 .zip),install.ps1 原代码拼 `"$fileName.sha256"` = `xxx.zip.sha256` → 资产不存在 → 校验被静默跳过。
**教训**: 引用 release 资产名先查实际命名(API 或 ls-remote),别猜。
