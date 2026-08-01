# 经验教训

## 2026-08-01 CN 网络双通道(续)

### L20: CN 网络特征 = api.github.com 通, github.com:443 不通
实测: `api.github.com` 可达(HTTP 200), `github.com:443` 连接超时。这是国内网络的典型状态(api 域名未被墙,主站被墙)。
**教训**: 涉及 GitHub 的客户端工具,网络通道必须是**双通道**: API 优先(CN 可达),github.com 兜底(海外)。只做单通道必然在某种网络下挂。
**验证方法**: `socket.create_connection(("github.com", 443), 5)` 超时 = 被墙;api.github.com 正常。

### L21: release 资产下载用 API 端点 `releases/assets/{id}`
`api.github.com/repos/OWNER/REPO/releases/assets/{id}` 匿名可下载(带 `Accept: application/octet-stream`),重定向到 `release-assets.githubusercontent.com` CDN —— **CN 可达,不经过被墙的 github.com**。比 `github.com/releases/download/...` 可靠。
**教训**: 需要下载 GitHub release 资产的客户端,优先用 API asset 端点;`browser_download_url`(github.com 域名)在 CN 不可用。

### L22: 安装脚本要自动重试 + 明确失败原因
用户要求: 下载自动重试最多 3 次,失败在 cmd 显示原因。实现: API×2 + github.com×1,每步打印失败原因,最终 Fail 带原因 + 代理指引。
**教训**: 面向大众的安装器,网络失败是常态,要有重试 + 可诊断的错误信息,别让用户猜。
