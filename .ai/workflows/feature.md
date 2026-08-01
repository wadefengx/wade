# Workflow: Feature(新功能)

## 触发
新命令、新能力、新平台支持。

## 步骤
1. **Spec**:`.ai/specs/active/SPEC-XXX.md`,含 Background/Goal/Scope/Non-goals/UX/API 契约/Acceptance Criteria/Tasks/QA Checklist,frontmatter `status: draft`
2. **Review**:PM 确认范围与不做项 → `status: approved`
3. **实现**:按 spec 写代码(ponytail:最短可行),`go build ./... && go vet ./...`
4. **测试**:表驱动单测覆盖 happy path + edge case + error case
5. **QA**:`.ai/harness/verify.sh` 全绿;CLI 手动实测(如 `wade node install` 走真实命令)
6. **回流**:lessons/decisions 写 `.ai/memory/`,通用做法沉淀 skill
7. **Commit**:`feat(scope): 描述`;spec 移入 `completed/`

## DoD
- [ ] Acceptance Criteria 全过
- [ ] verify.sh 全绿
- [ ] 经验已回流
