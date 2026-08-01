# Runtime — 执行管线与运行机制

## Pipeline(标准执行管线)

```
需求 → PM 系分(范围/边界/不做项)
→ Architect 设计(模块/API 契约/数据模型)
→ 任务拆分(Lane,按领域切,文件不冲突)
→ 并行开发(每 lane 独立,遵守 ponytail)
→ QA 验收(单测 + harness + 实测)
→ 回流(memory → skill → knowledge)
→ commit + 发布
```

## Lane 切分原则

- 按领域切:internal/node / internal/registry / internal/go / cmd / docs / scripts
- 每个 lane 独占一组文件;跨 lane 共享文件(如 AGENTS.md、go.mod)指定单一 owner
- 契约先定死(函数签名、CLI 命令语义写进 spec),并行不联调
- 单人/单 agent 串行时,lane = 提交单元,仍保持文件域隔离

## Lane 状态机

```
Draft → Ready → Running → Review → QA → Done → Merged
                ↓ 多轮无进展
              Blocked
```

## Confidence 机制

每个 lane 完成时自评置信度(0-1):

| 维度 | 权重 |
|------|------|
| 实现完整性(所有 acceptance criteria 覆盖) | 0.4 |
| 测试覆盖(关键路径有单测) | 0.3 |
| 契约符合度(与 spec API 契约一致) | 0.3 |

- `>= 0.7`:直接进入 QA
- `< 0.7`:自动 Architect Review
- `< 0.5`:外部 review(换一个 agent/人)

## Model Routing

| 任务类型 | 模型策略 |
|---------|---------|
| 日常问答/简单修改 | Flash 级别(快) |
| 复杂实现/重构 | 顶级模型(gpt-5.6-terra / claude) |
| 架构决策/契约设计 | 深度思考模型 |
| 重复机械任务 | 脚本/harness,不费模型 |

## Tool Policy

- CLI 开发:Go 工具链(gofmt/vet/test/build),不引额外工具
- 发布:GitHub Actions + gh + git(见 `.ai/workflows/release.md`)
- 中国网络:镜像优先(npmmirror/goproxy.cn),代理 ClashX 127.0.0.1:7890(git push 卡住时配置 http.proxy)

## Coding Policy

- gofmt 格式化,`internal/pkg` 布局
- 错误包装 `fmt.Errorf("...: %w", err)`;库代码不 panic
- 表驱动测试;HOME 重定向隔离真实 ~/.wade
- 提交 `feat(scope): 描述` 功能粒度
