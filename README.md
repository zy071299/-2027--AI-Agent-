# 企业知识问答 Agent

需要 Go 1.22+，无需安装第三方依赖。在仓库根目录执行。

```powershell
go run ./src -q '报销期限多久'
go run ./src -q '视频导出格式'
go run ./src -q '报销期限' -fail read
go test ./... -v
```

默认使用离线 Mock。真实模型模式使用兼容 Chat Completions 且支持 JSON 对象输出的服务。设置 `AGENT_ENDPOINT` 为完整接口地址，设置 `AGENT_MODEL` 和服务所需的 `AGENT_API_KEY`，然后运行 `go run ./src -mode model -q '报销期限多久'`。

用 `-docs` 替换合成资料，用 `-history` 传入历史文本。输出 JSON 包含答案、引用、资料缺口、任务状态和执行 Trace。评审入口见 [评估记录](docs/03-evaluation.md)。
