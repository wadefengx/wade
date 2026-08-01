# 经验教训

## 2026-08-01 wade -u CN 网络超时(续)

### L28: 版本解析/下载都要双通道, 不能只改一处
install.ps1 改了双通道(API 优先 + redirect 兜底),但 cmd/update.go 的 `getLatestVersion` 还只有 redirect——用户 `wade -u` 必挂(`context deadline exceeded`)。下载也是 github.com 单通道。
**教训**: 同样的网络适配逻辑必须全链路一致(install + update + 启动检查),漏一处就有一处挂。改完用 CN 网络特征(api 通 / github.com 不通)端到端实测。

### L29: GitHub 资产 JSON 的 id 在 name 前面
`{"url":..., "id":123, "node_id":..., "name":"xxx.zip", "uploader":{"id":...}}` —— 资产自己的 `"id"` 在 `"name"` **之前**。向后找会命中 `uploader.id`(41898282 假 id)。
**修复**: 从 name 位置**往回**找最近的 `"id":`。用真实 API body 验证 + 单测固化。

### L30: 用真实 API 数据验证解析逻辑
assetIDByName 先拿真实 release body 验证(JSON 紧凑格式, 无空格), 再写单测。别用 json.dumps(会加空格改变格式)验证。
