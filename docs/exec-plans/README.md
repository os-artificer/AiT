# exec-plans

本目录存放 **AiT 架构文档** 与 **经确认后归档的 Cursor 执行方案**。

## 常驻文档

| 文件 | 说明 |
|------|------|
| [design.md](design.md) | 架构与行为：CLI、SQLite 模型、`!ait` meta 协议、tmux/TTL、sync、AI 等（当前主干实现） |

## 归档方案

除 `design.md` 外，其它 `*.md` 为在 Cursor 中制定、经确认需要保留的执行计划或重构方案（例如功能实现计划、全库风格重构说明）。

**命名**：`kebab-case` 英文主题，如 `go-style-refactor.md`；需要按时间排序时可使用 `YYYY-MM-DD-topic.md`。

**何时归档**

- 用户明确要求将方案写入仓库；或
- 方案已按计划在代码库中实施完毕，且用户确认需要保留文档记录。

**不归档**

- 纯对话、已作废方案；
- 仅存在于 Cursor 本地 `.cursor/plans/`、未要求入库的草稿。

**Agent 约定**：见 [`.cursor/rules/docs-exec-plans.mdc`](../../.cursor/rules/docs-exec-plans.mdc)。

用户向导读 [README.md](../../README.md)；开发向导读 [AGENTS.md](../../AGENTS.md)。
