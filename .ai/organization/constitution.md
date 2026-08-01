# Constitution — Wade 项目组织宪法

> AI-Native 协作的五大对象模型与三系统。派生自 wade-ai/AI_DEV_INSTRUCTION.md §4,按单人 Go CLI 项目裁剪。

## 五大对象模型

```
Organization(协作) → Specification(做什么) → Workflow(怎么做) → Knowledge(知道什么) → Runtime(如何执行)
```

| 对象 | 位置 | 职责 |
|------|------|------|
| **Organization** | `.ai/organization/` | 角色分工、协作边界、DoD(完成定义) |
| **Specification** | `.ai/specs/` | 唯一真相源。实现前必须有 spec;代码与 spec 冲突时以 spec 为准并回写 spec |
| **Workflow** | `.ai/workflows/` | 可复用流程:feature/bugfix/refactor/release |
| **Knowledge** | `.ai/knowledge/` + `.ai/memory/` | 被动参考 + 长期记忆(经验/决策/教训) |
| **Runtime** | `.ai/runtime/` | 执行管线、lane 状态机、置信度、工具/模型策略 |

## 三系统(贯穿生命周期)

- **Memory**:每个完成的任务回流经验/决策/教训 → `.ai/memory/{lessons,decisions}/`
- **Skill**:可复用能力沉淀 → `.ai/skills/`(经 Architect review 后从 memory 晋升)
- **Harness**:AI 质量系统 → `.ai/harness/verify.sh`(可执行验收,PASS/FAIL 计数 + 退出码)

## 核心原则

1. **AI-first,人定目标**:需求由人提出,spec 驱动,agent 执行,人验收。
2. **Spec 是唯一真相源**:无 spec 不写代码。Refactoring = 改实现保持 spec;Feature = 改 spec 再实现。
3. **一切沉淀回流**:任务完成必须回流 memory → skill → knowledge,下个任务不重复踩坑。
4. **Ponytail 哲学**:最短可行实现;复用 > 新建;根因修复;不建无用的抽象。
5. **质量门禁全绿**:lint + vet + test + build + harness 每期必须过。

## 角色(单人项目精简版)

| 角色 | 职责 | 边界 |
|------|------|------|
| **PM** | 需求系分:范围/边界/不做项,拆 lane,定验收标准 | 不写实现代码 |
| **Architect** | 架构决策、API 契约、技术选型、ADR 记录 | 不写业务代码(可写骨架) |
| **Engineer** | lane 实现,遵守 spec + ponytail | 不越界改其他 lane 文件 |
| **QA** | 单测 + harness 回归 + 浏览器实测 | 不修代码,只报缺陷 |
| **Supervisor(人)** | 提出需求、验收、拍板 | 最终决策权 |

单人项目 = 同一人/agent 身兼多角色,但**角色边界在流程中仍须区分**(先当 PM 拆任务,再当 Engineer 实现,再当 QA 验收)。

## DoD(完成定义)

一个 lane 的完成标准:
- [ ] 实现符合 spec 的 Acceptance Criteria
- [ ] 单测通过(`go test ./...`)
- [ ] 质量门禁全绿(`make test` + `go vet` + `gofmt`)
- [ ] 关键流程 harness 验收通过(如 `.ai/harness/verify.sh`)
- [ ] 经验回流 `.ai/memory/`(lessons 或 decisions)
- [ ] 功能粒度 commit(`feat(scope): 描述`)
