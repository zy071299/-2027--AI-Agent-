# 验收记录

默认 Mock 的实际运行记录见下表。每份 JSON 的 trace 包含动作、参数、简要依据和工具结果。

| 输入 | 预期状态 | 记录 |
|---|---|---|
| 视频导出格式 | answered | [正常](../examples/normal.json) |
| 报销期限多久 | conflict | [冲突](../examples/conflict.json) |
| 明天会下雨吗 | insufficient | [资料外](../examples/outside.json) |
| 费用票据最晚什么时候提交 | conflict | [改写](../examples/paraphrase.json) |
| 那可以用哪些类型 | answered | [追问](../examples/followup.json) |
| 报销期限，模拟读取失败 | tool_error | [工具失败](../examples/failure.json) |

[测试日志](../examples/test-results.txt)。引用校验保证来源和原文完整性，答案是否得到引用支持仍需评审。Mock 的冲突判断只适用于当前合成数据。


文档变化测试验证仅 D01 时回答 30 天，加入 D02 时补查并报告冲突，将单份原文改为 45 天时答案和引用同步更新。上下文测试验证移除历史后，同一追问返回资料不足。本地 HTTP 测试验证真实模型适配器将检索和读取结果送入下一轮。执行保护测试覆盖伪造引用和 12 步上限。

## DeepSeek 真实模型对照

同一问题「报销期限多久」分别使用仅 D01 和默认资料执行。两次输出的引用 ID、版本日期及原文均与输入资料一致。

- [仅 D01 输入](../examples/d01-only.json) → [真实输出](../examples/deepseek-single.json)。模型检索并读取 D01 后返回 answered，依据原文回答 30 天。初次运行曾添加无出处的标题，收紧引用提示词后，本次回答使用 D01、版本日期和原文。
- [默认输入](../data/docs.json) → [真实输出](../examples/deepseek-conflict.json)。模型读取 D01 和 D02 后补查适用范围、优先级和版本，未找到裁决依据，返回 conflict，并列引用 30 天和 15 天，请用户确认适用制度。

两组对照验证了模型会随资料变化调整读取和补查动作。其他验收场景的运行证据为离线 Mock。

## D03 裁决规则对照

用户在本地使用 DeepSeek 提问「我是中国大陆员工，这笔是差旅报销，报销期限多久？」。以下记录来自用户提供的实际终端输出。

- 使用[默认资料](../data/docs.json)时，模型读取 D01 和 D02，补查后仍缺适用依据，返回 conflict。[运行记录](../examples/deepseek-d03-baseline.json)
- 使用[含 D03 的资料](../examples/docs-with-d03.json)时，首次检索仅命中 D01 和 D02。补查适用范围与优先级时发现 D03，读取原文后返回 answered，引用 D02 和 D03，回答 15 天。[运行记录](../examples/deepseek-d03-resolution.json)

D03 写的是「中国员工」，提问使用「中国大陆员工」，资料未定义两者的关系。这组运行验证了补查发现新资料后触发读取的决策路径，15 天是否适用于提问中的人群仍缺少明确的范围定义。

## 最终代码检查

`go test ./... -count=1 -v` 和 `go vet ./...` 均通过。测试包含模型动作 JSON 无效和请求超时两种失败场景。[完整测试日志](../examples/test-results.txt)
