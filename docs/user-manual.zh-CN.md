# AiT 用户手册

**语言：** [English](user-manual.md) · 简体中文（本文）

本文档描述 AiT 的安装后配置、命令用法、提示符、数据目录与安全约定。架构说明见 [design.zh-CN.md](design.zh-CN.md)。

---

## 快速开始

```bash
# Deepseek API Key（存入操作系统密钥环，不写入配置文件）
./bin/ait config set deepseek.api_key <your-key>

# 登录：完整参数说明见 help
./bin/ait login --help

# 交互登录并持久化会话
./bin/ait login deploy@example.com --alias prod
./bin/ait login -p 2222 -i ~/.ssh/id_ed25519 deploy@10.0.0.5
./bin/ait login -J bastion app.internal

# 一次性远程命令（类似 ssh host cmd；默认不持久化至 SQLite）
./bin/ait login deploy@example.com uptime

# 列出、恢复、切换会话
./bin/ait list
./bin/ait load prod
./bin/ait session switch staging

# 文件同步（可复用已保存会话）
./bin/ait sync ./local prod:/remote/path --session prod
```

配置模板（可选）：

```bash
mkdir -p ~/.config/ait
cp configs/ait.example.yaml ~/.config/ait/config.yaml
```

---

## 命令参考

| 命令 | 说明 |
|------|------|
| `ait shell` | 交互模式（`AiT> `）：可执行 `login`、`list`、`help` 等；`exit` 退出 |
| `ait login` | 连接主机；`ait login --help` 列出全部 flags；`host cmd` 为一次性执行 |
| `ait list` | 列出已保存会话 |
| `ait load <id\|alias>` | 恢复会话；认证过期须 `--password` |
| `ait suspend` | 在 SSH 会话内使用 `!ait suspend` |
| `ait session switch [id\|alias]` | 切换会话（先挂起当前再加载目标） |
| `ait session delete <id\|alias>` | 删除单个会话 |
| `ait session delete all` | 删除全部会话（须确认；`--force` 跳过确认） |
| `ait sync <src> <dst>` | 跨主机同步（`--session` 复用连接参数） |
| `ait config set/get <key>` | 配置项；`deepseek.api_key` 经密钥环存储 |

常用标志：`--tmux`（远程 tmux 保活）、`--port`、`--identity`、`--password`、`--alias`。`session switch --force` 可在当前会话仍标记为 active 时强制切换（不保存当前 scrollback/cwd）。

---

## 提示符

| 场景 | 提示符 |
|------|--------|
| `ait shell` | `AiT> ` |
| SSH 远程 | `AiT <user>@<host> ` |
| `!ait local` | `AiT Local> ` |

---

## 安全与 SSH

默认校验 `~/.ssh/known_hosts`；新主机须先执行 `ssh-keyscan -H <host> >> ~/.ssh/known_hosts`。开发环境可设置 `AIT_SSH_INSECURE=1` 跳过校验（禁止用于生产环境）。`--password` 将出现在进程列表中；生产环境优先使用操作系统密钥环或交互式凭据输入。

---

## SSH 会话内命令

| 输入 | 作用 |
|------|------|
| `!ait help` / `!ait` | 打印会话内命令与可本地执行的 CLI 说明 |
| `!ait <子命令> …` | 在本地执行 `ait` 子命令；`login`/`load` 须先退出 SSH |
| `!ait ai <问题>` | 调用 Deepseek（建议使用） |
| `!! <问题>` | 同上（向后兼容别名；`!!` 后须有空格） |
| `!ait local` | 本地模式：命令在本机执行，SSH 连接保持 |
| `!ait remote` | 退出本地模式，恢复向远程发送输入 |
| `!ait suspend` | 挂起：保存 scrollback 与远程 cwd，断开连接 |

---

## 数据与配置目录

| 平台 | 数据目录（SQLite、`scrollback/`） | 配置目录 |
|------|-----------------------------------|----------|
| Linux | `~/.local/share/ait/` | `~/.config/ait/config.yaml` |
| macOS | `~/Library/Application Support/ait/` | 同上 |
| Windows | `%LOCALAPPDATA%\ait\` | `%LOCALAPPDATA%\ait\config.yaml` |

Deepseek API Key 与 SSH 密码缓存存于**操作系统密钥环**，不以明文写入 SQLite。

---

## 限制说明

- 未使用 `--tmux` 时，挂起后不保证远程进程继续运行。
- 密码认证会话的 `ait sync` 使用内置 SFTP（不调用系统 `rsync`/`scp` 的非交互密码流程）。
- 从本机执行 `ait session switch` 前，若仍处于 SSH 会话中，须先 `!ait suspend` 保存当前状态，或使用 `--force`（不保存当前 scrollback/cwd）。

---

## 延伸阅读

| 文档 | 说明 |
|------|------|
| [README.zh-CN.md](../README.zh-CN.md) | 项目概览与安装 |
| [design.zh-CN.md](design.zh-CN.md) | 架构与行为 |
| [AGENTS.zh-CN.md](../AGENTS.zh-CN.md) | 代码结构 |
| [CONTRIBUTING.zh-CN.md](../CONTRIBUTING.zh-CN.md) | 贡献指南 |
