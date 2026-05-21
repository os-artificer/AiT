# 更新日志

**语言：** [English](CHANGELOG.md) · 简体中文（本页）

本文仅记录 **产品与 CLI** 的功能性变更，不包含纯文档类更新。发版后请将 `[未发布]` 中的条目移至带日期的版本段。

## [未发布]

### 【新增】

- AI 增强 SSH 客户端核心：`login`、`list`、`load`、`suspend`、`session switch`、`session delete`
- SQLite 会话持久化、别名、`session.auth_ttl` 与操作系统 keyring 凭据
- 通过 `ssh -G` 解析 OpenSSH 配置，数据面使用 Go `x/crypto/ssh`
- 可选远程 `--tmux` 保活（`ait-<session_id>`）
- 会话内 `!ait`：`ai` / `!!`、`local` / `remote`、`suspend` 与 help
- Deepseek 助手（`!ait ai`）
- 文件传输 `ait get` / `ait put`（SFTP，单文件）；会话内 `!ait get` / `!ait put`
- 本地入口 `ait shell` REPL（`AiT>` 提示符）

### 【变更】

- **破坏性变更：** 移除 `ait sync`，请改用 `ait get` / `ait put`

### 【优化】

- CLI 包结构收敛至 `internal/ait`
- 拆分 `internal/ait` 与 `internal/pkg/ssh` 等包，职责更清晰
- 会话内 `!ait` 交互与输出行为

### 【修复】

- SSH 登录认证、host key 校验与远程 prompt 初始化
- 会话内 `!ait` 元命令输出问题
