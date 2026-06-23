# 更新日志

**语言：** [English](CHANGELOG.md) · 简体中文（本页）

本文仅记录 **产品与 CLI** 的功能性变更，不包含纯文档类更新。
发版后请将 `[未发布]` 中的条目移至带日期的版本段。

## [Unreleased]

- **CLI 重命名：** 工具箱首个工具为 `ai-ssh`（`bin/ai-ssh`、会话内 `!ait` meta）；默认数据根 `~/.local/share/AiT`，自动从旧 `share/ait` 迁移；keyring 服务名 `AiT`，惰性读取旧 `ait` 条目。
- **破坏性变更：** 会话内 meta 前缀为 `!ait`；`!ai-ssh` 不再识别。`!!` 仍为 AI 快捷方式。
- **AI (Eino refactor):** Replace `sashabaranov/go-openai` with CloudWeGo Eino ChatModel; shared DTOs in `internal/pkg/ai/apitypes`; tiered Chat routes to Ollama via Eino; tiered Agent on Ollama and Anthropic-style providers use `specialized_agent_completion`; `ai-ssh ai agent` 单轮现已与 REPL 一样接入 Capture/CLI/MaxSteps。

## [v1.0.0-beta.2]

- 【新增】SHELL模式支持完整命令行交互，命令搜索路径为：内置命令，其次扩展命令与PATH环境变量。
- 【优化】会话连接唯一键扩展为 username+host+port+alias，同一目标机可用多个 `--alias` 保存独立会话；无 flag login 仅更新无 alias 匹配行。

## v1.0.0-beta.1

- 【新增】AI 增强 SSH 客户端，可以在 SSH 会话内通过 `!! '问题'`、`!ait ai ask`、`!ait ai agent` 或 `!ait ai` 激活 AI 能力。
- 【新增】自带文件传输功能，在 SSH 会话内通过 `!ait get`/`put` 实现文件下载与上传。
- 【新增】在SSH会话内随时进行local与remote模式间的切换。
