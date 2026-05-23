<p align="center">
  <img src="assets/logo.png" alt="AiT logo" width="420">
</p>

**语言：** [English](README.md) · 简体中文（本文）

# AiT

**AiT** — 跨平台（Linux / macOS / Windows）AI 增强的命令行工具箱，包含但不限于Terminal与SSH，致力于成为运维人员的趁手利器。

## 特性

在兼容 OpenSSH 配置与连接参数的前提下，AiT 在会话管理、恢复与终端内协作上提供以下能力（相对常规 `ssh` 交互会话）。

### 会话管理

| 功能 | AiT | 常规 SSH 客户端 |
|------|-----|-----------------|
| 会话持久化 | SQLite 保存连接参数、别名与认证 TTL（`session.auth_ttl`） | 无会话库；依赖 shell 历史或外部记录 |
| 快速恢复 | `ait session --load` / `--switch`；认证 TTL 内免凭据重连（重新输入密码时刷新 TTL） | 每次须完整指定 `user@host` 与 flags |
| 挂起与上下文 | `!ait suspend` 保存 scrollback 与远程工作目录 | 断开连接后终端上下文不保留 |

### 远程保活

| 功能 | AiT | 常规 SSH 客户端 |
|------|-----|-----------------|
| 进程保活 | `--tmux` 自动 attach/detach `ait-<session_id>` | 须手动创建并管理 tmux 或 screen 会话 |

### 终端内协作

| 功能 | AiT | 常规 SSH 客户端 |
|------|-----|-----------------|
| AI 模式 | 本机 `ait ai`（`AiT AI>`）与会话内 `!ait ai` / `!!`（远程 `run` 须确认）；Deepseek | 无集成；须另开工具窗口 |
| 本地模式 | `!ait local` / `!ait remote` 切换输入目标，SSH 连接保持 | 本机与远程须分终端操作 |
| 文件传输 | `ait get` / `ait put`（SFTP，单文件）；`!ait get` / `!ait put` 复用当前 SSH 会话 | 须自行拼接 scp/rsync 的 SSH 选项 |

### 本地入口

| 功能 | AiT | 常规 SSH 客户端 |
|------|-----|-----------------|
| 统一 CLI | `ait shell` REPL 管理 `login` / `list` / `load` | 无等效一体化 REPL |

## 安装

**前置条件**：Go **1.25+**；PATH 中须存在 **OpenSSH 客户端**（`ssh`，供 `login`/`load` 解析 `~/.ssh/config`）。

**从源码构建**（本地构建与验证）：

```bash
git clone https://github.com/os-artificer/AiT.git
cd AiT
make build
./bin/ait --help
```

**`go install`**（模块可通过 `go install` 解析时）：

```bash
go install github.com/os-artificer/AiT/cmd/ait@latest
ait --help
```

将 `$(go env GOPATH)/bin` 加入 PATH。

> 配置项与操作说明见 [用户手册](docs/user-manual.zh-CN.md)。

## 文档

| 文档 | 路径 |
|------|------|
| 用户手册 | [docs/user-manual.zh-CN.md](docs/user-manual.zh-CN.md) |
| 更新日志 | [CHANGELOG.zh-CN.md](CHANGELOG.zh-CN.md) |
| 贡献指南 | [CONTRIBUTING.zh-CN.md](CONTRIBUTING.zh-CN.md) |
| 行为准则 | [CODE_OF_CONDUCT.zh-CN.md](CODE_OF_CONDUCT.zh-CN.md) |
| 架构与行为 | [docs/design.zh-CN.md](docs/design.zh-CN.md) |
| 代码结构 | [AGENTS.zh-CN.md](AGENTS.zh-CN.md) |
