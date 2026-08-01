---
status: draft   # draft | approved | implementing | testing | done
---

# SPEC-XXX: 标题

## Background
为什么做?解决什么问题?(1-3 句)

## Goal
做完后达到什么状态(可验收)。

## Scope
- 做:...
- 不做:...

## Non-goals
明确不做的事(YAGNI 边界)。

## UX / CLI
```bash
wade <command> <args>
# 期望输出
```

## API 契约
- 函数签名/命令语义(跨 lane 依赖先定死)
- 错误处理约定

## Acceptance Criteria
- [ ] ...

## Risks
- 风险 + 缓解

## Tasks
- [ ] T1: ...

## QA Checklist
- [ ] 单测覆盖 happy/edge/error
- [ ] harness verify.sh 全绿
- [ ] CLI 手动实测
