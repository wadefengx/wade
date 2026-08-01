# 决策记录

## D1: Go 而非 Node 实现 wade
- 日期: 2026-07-28
- 决策: 用 Go + cobra 构建单二进制, 不用 Node
- 理由: 单二进制安装一次, 无 Node 依赖, 跨平台(brew/scoop/curl)
- 教训: wade 自身绝不依赖 Node(nvm/cgr 的痛点根源)

## D2: Shim 机制而非 PATH 修改
- 日期: 2026-07-28
- 决策: ~/.wade/shims/ 符号链接切换版本, 不每次改 PATH
- 理由: PATH 设置一次, 切换瞬间完成
- 演进: Windows 无权限时 fallback 硬链接(platform 层, 2026-07-31)

## D3: 镜像优先(China-friendly)
- 日期: 2026-07-28
- 决策: Node 下载默认 npmmirror, Go 默认 google-cn + goproxy.cn, pip 默认 tsinghua
- 理由: 目标用户在中国大陆网络环境

## D4: 发布自动化全链路
- 日期: 2026-07-31
- 决策: tag push → Actions 构建 5 平台 → release(含 install.sh)→ update-tap 自动更新 tap/bucket
- 理由: 发布零人工干预
- 坑: tap/bucket 默认分支必须是 main(创建时 auto_init 是 master, 导致 update-tap 推 main 失败)

## D5: .wade-version 项目锁定
- 日期: 2026-07-31
- 决策: 仿 nvm .nvmrc, cwd 向上找 .wade-version, `wade node use` 无参自动读取
- 理由: 进项目目录自动切版本
