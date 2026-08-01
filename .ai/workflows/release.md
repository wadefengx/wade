# Workflow: Release(发布)

## 触发
里程碑完成、版本迭代。

## 版本号
`vMAJOR.MINOR.PATCH` — 参考 CHANGELOG.md 顶部区块。

## 步骤
1. **CHANGELOG**:顶部加 `## [vX.Y.Z] — 日期`,列 Added/Fixed/Changed
2. **质量门禁**:`make test` + `go vet` + `gofmt -l .` 全绿
3. **Commit**:`docs: CHANGELOG vX.Y.Z`
4. **Tag + Push**:
   ```bash
   git tag vX.Y.Z
   git push origin main && git push origin vX.Y.Z
   ```
5. **监控**:GitHub Actions 自动跑 build(5 平台)→ release(11 资产含 install.sh)→ update-tap(homebrew-tap + scoop-wade 自动更新)
6. **验证**:
   - `curl -sI .../releases/latest/download/install.sh` → 200
   - tap/bucket 版本号 = vX.Y.Z
   - `wade version` 输出 vX.Y.Z(版本注入 ldflags 生效)

## 自动发布链路(已配置)
```
tag push → release.yml → build matrix (darwin×2/linux×2/windows)
→ softprops/action-gh-release (资产: 5×tar.gz/zip + 5×sha256 + install.sh)
→ update-tap job → scripts/release-shas.sh --update → push 到 homebrew-tap/scoop-wade
```

## 陷阱
- github.com 被墙时:`git config http.proxy http://127.0.0.1:7890` 再 push,完事 unset
- update-tap 失败先查 tap/bucket 仓库默认分支是否 main
- TAP_TOKEN secret 需要 homebrew-tap + scoop-wade 的写权限
