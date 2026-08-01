# Workflow: Refactor(重构)

## 触发
技术债、架构演进、代码重复。

## 原则
- **改实现,保持 spec**;spec 需要变 = feature,不是 refactor
- 行为不变:重构前后测试必须全绿(测试是重构的安全网)
- 小步提交,每步可回滚

## 步骤
1. **基线**:`go test ./...` 全绿记录基线
2. **梳理**:列出变更面(文件/函数/调用方),画清依赖
3. **实施**:按依赖序小步改,每步 `go build + go test`
4. **验证**:verify.sh 全绿 + 覆盖率不降
5. **回流**:`.ai/memory/architecture/` 记录架构决策
6. **Commit**:`refactor(scope): 描述`

## 陷阱
- 不做 drive-by 重构:任务没要求的不动
- 文档与实现漂移(如 AGENTS.md 声称 platform 层存在但实际没有)——重构时同步更新文档
