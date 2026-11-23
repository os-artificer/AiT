# AiT 文档

| 分区 | 路径 | 用途 |
|------|------|------|
| 用户手册 | [guide/](guide/) | 各工具 CLI |
| 变更记录 | [changelog/](changelog/) | 功能变更（总入口 [CHANGELOG.md](../CHANGELOG.md)） |
| 系统架构 | [architecture/](architecture/) | 分层、实现与测试夹具 |

## 用户手册

| 工具 | 说明 |
|------|------|
| [ait](guide/ait.md) | Agent 入口与工具箱 shell |
| [aissh](guide/aissh.md) | SSH（`aitssh`） |
| [aiimg](guide/aiimg.md) | 图像元数据 |
| [ailog](guide/ailog.md) | 系统日志 |
| [aihost](guide/aihost.md) | 主机监视 |
| [aitkeep](guide/aitkeep.md) | 进程监管 |
| [aitvid](guide/aitvid.md) | 视频内容检索（OCR / VLM） |

## 变更记录

- [未发布](changelog/unreleased.md)
- [v1.0.0](changelog/v1.0.0.md)

## 系统架构

| 文档 | 说明 |
|------|------|
| [overview](architecture/overview.md) | 工具箱结构与分层 |
| [ait](architecture/ait.md) | `ait` 实现 |
| [aissh](architecture/aissh.md) | `aissh` 实现（含 package layering） |
| [aiimg](architecture/aiimg.md) | `aiimg` 实现 |
| [ailog](architecture/ailog.md) | `ailog` 实现 |
| [aihost](architecture/aihost.md) | `aihost` 实现 |
| [aitkeep](architecture/aitkeep.md) | `aitkeep` 实现 |
| [aitvid](architecture/aitvid.md) | `aitvid` 实现 |
| [testenv-mock](architecture/testenv-mock.md) | 仓内回环 SSH/HTTP mock |

方案归档：`docs/<kebab>.md`（见 [docs-exec-plans](../.cursor/rules/docs-exec-plans.mdc)）。
