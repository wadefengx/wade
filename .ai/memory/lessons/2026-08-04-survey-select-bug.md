# 经验教训

## 2026-08-04 survey.Select 写入 []string 静默失败

### L39: survey v2 单选 Select 只能写入 *string,传 *[]string 会静默失败
`survey.Select`(单选)的 answer 目标是 `*string`(或 `*int`)。传 `*[]string` 时,
core/write.go 的 writeAnswer 对 slice 目标返回 "Unable to convert from string to type slice",
survey 把错误**吞掉**(AskOne 的 error 常被忽略),目标保持零值 → 后续判断全 false。

**症状**: `wade init` 交互式选 "All of the above" 后,Go/Python 分支从未执行
(mirror/proxy/pip 都没配置),但 `init -y` 正常——因为 -y 直接赋值 runtimes 不走 survey。
排查了 3 轮(v0.5.4 summary 反馈、v0.5.5 Python 接管)才发现是交互入口的写入 bug。

**教训**:
1. survey 单选用 `var choice string`,多选用 `var choices []string`(MultiSelect)——别混
2. `survey.AskOne` 的 error 不要忽略,至少 `if err != nil { return err }`
3. 交互分支和 -y 分支行为不一致时,优先怀疑**交互输入写入路径**,不是业务逻辑

### L40: 交互/非交互双路径 = 双份测试面
有 autoYes 分支的代码,测试要覆盖两条路:直接赋值(易)+ survey 写入(难,可抽 choice→runtimes 纯函数测)。
本次 TestRuntimesSelection 覆盖了选择映射逻辑。
