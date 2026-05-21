<p align="center">
  <img src="assets/logo.png" alt="AiT logo" width="420">
</p>

# AiT

**AI Terminal** — 跨平台（Linux / macOS / Windows）的 AI 增强 SSH 客户端。

- SQLite 保存会话；TTL 内 `load` / `switch` 可免密重连（**仅在你重新输入密码时才会刷新** TTL，见 `session.auth_ttl`）
- 可选 `--tmux` 在远程保活 shell；默认仅保存连接信息并在重连时恢复 scrollback / 工作目录
- Deepseek 会话内助手（`!ait ai` / `!!`）、本地模式（`!ait local`）与缺失系统命令时的安装建议
- `sync` 子命令：优先 `rsync` / `scp`，回退 SFTP

架构与行为详见 **[docs/exec-plans/design.md](docs/exec-plans/design.md)**。

## 快速开始

```bash
make build

# Deepseek API Key（存入系统钥匙串，不写配置文件）
./bin/ait config set deepseek.api_key <your-key>

# 登录：完整参数说明见 help（逐项列出 OpenSSH 与 AiT flags）
./bin/ait login --help

# 交互登录并保存会话
./bin/ait login deploy@example.com --alias prod
./bin/ait login -p 2222 -i ~/.ssh/id_ed25519 deploy@10.0.0.5
./bin/ait login -J bastion app.internal

# 一次性远程命令（类似 ssh host cmd，默认不写库）
./bin/ait login deploy@example.com uptime

# 列出 / 恢复 / 切换会话
./bin/ait list
./bin/ait load prod
./bin/ait session switch staging    # 无参数时交互选择

# 文件同步（可复用已保存会话）
./bin/ait sync ./local prod:/remote/path --session prod
```

复制配置模板（可选）：

```bash
# Linux 示例路径
mkdir -p ~/.config/ait
cp configs/ait.example.yaml ~/.config/ait/config.yaml
```

## 命令参考

| 命令 | 说明 |
|------|------|
| `ait shell` | 交互模式（`AiT> `）：可直接 `login`、`list`、`help` 等；`exit` 退出 |
| `ait login` | 连接主机；`ait login --help` 查看全部 flags（-p/-l/-i/-F/-J/-o 等）；`host cmd` 为一次性执行 |
| `ait list` | 列出已保存会话 |
| `ait load <id\|alias>` | 恢复会话；过期需 `--password` |
| `ait suspend` | 说明：在 SSH 内使用 `!ait suspend` |
| `ait session switch [id\|alias]` | 切换会话（先挂起当前再加载目标） |
| `ait session delete <id\|alias>` | 删除单个会话 |
| `ait session delete all` | 删除全部会话（需确认，`--force` 跳过） |
| `ait sync <src> <dst>` | 跨主机同步（`--session` 复用连接） |
| `ait config set/get <key>` | 配置项；`deepseek.api_key` 走钥匙串 |

常用标志：`--tmux`（远程 tmux 保活）、`--port`、`--identity`、`--password`、`--alias`。`session switch --force` 可在当前会话仍标记为 active 时强制切换（**不会**保存当前 scrollback/cwd）。

### 提示符

| 场景 | 提示符 |
|------|--------|
| `ait shell` | `AiT> ` |
| SSH 远程 | `AiT <user>@<host> ` |
| `!ait local` | `AiT Local> ` |

**安全与 SSH：** 默认校验 `~/.ssh/known_hosts`；新主机需先 `ssh-keyscan -H <host> >> ~/.ssh/known_hosts`。开发可设 `AIT_SSH_INSECURE=1` 跳过校验（勿用于生产）。`--password` 会出现在进程列表中，生产环境优先用钥匙串/交互提示。

### SSH 交互内

| 输入 | 作用 |
|------|------|
| `!ait help` / `!ait` | 打印会话内命令与可本地执行的 `ait` CLI 说明 |
| `!ait <子命令> …` | 在本地执行 `ait` 子命令（如 `!ait list`）；`login`/`load` 需先退出 SSH |
| `!ait ai <问题>` | 调用 Deepseek（推荐） |
| `!! <问题>` | 同上（兼容别名；`!!` 后须有空格） |
| `!ait local` | 本地模式：命令在本机执行，SSH 保持连接；提示符为 `AiT Local> ` |
| `!ait remote` | 退出本地模式，恢复向远程发送输入 |
| `!ait suspend` | 挂起：保存 scrollback / 远程 cwd，断开连接 |

## 数据与配置位置

| 平台 | 数据目录（SQLite、`scrollback/`） | 配置目录 |
|------|-----------------------------------|----------|
| Linux | `~/.local/share/ait/` | `~/.config/ait/config.yaml` |
| macOS | `~/Library/Application Support/ait/` | 同上 |
| Windows | `%LOCALAPPDATA%\ait\` | `%LOCALAPPDATA%\ait\config.yaml` |

敏感信息（Deepseek API Key、SSH 密码缓存）保存在 **OS 钥匙串**，不落 SQLite 明文。

## 目录说明

| 目录 | 用途 |
|------|------|
| `cmd/ait` | 主程序入口 |
| `internal/ait` | Cobra CLI、`ait shell` REPL（`login`、`list`、`load`、`session` 等） |
| `internal/pkg` | 实现库：`store`、`session`、`ssh/args`、`ssh/client`、`ai`、`sync`、`config`、`creds`、`pathsafe`…（详见 [AGENTS.md](AGENTS.md)） |
| `configs` | `ait.example.yaml` 配置模板 |
| `docs/exec-plans` | [design.md](docs/exec-plans/design.md) 架构与行为；归档的 Cursor 执行方案 |
| `pkg` | 对外可导入库（当前未使用） |
| `.cursor/rules` | Cursor / Agent 规则（含 Go 编码规范 `go-structure.mdc`） |
| `build`、`scripts` | 构建与 CI |

其余目录遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout)，见各目录 `README.md`。

## 开发

```bash
make build       # 编译到 bin/ait
make test        # 运行测试
make vet         # 静态检查
make fmt         # 格式化
make test-cross  # Windows / macOS 交叉编译门禁
./bin/ait --help
```

需要 **Go 1.25+**。

架构细节见 [docs/exec-plans/design.md](docs/exec-plans/design.md)；在本仓库用 Agent 或贡献 Go 代码时请参阅 [AGENTS.md](AGENTS.md)（包布局与 [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) 编码规范）。Cursor 方案归档见 [docs/exec-plans/](docs/exec-plans/)。

## 模块路径

```
github.com/os-artificer/AiT
```

## 限制（摘要）

- 未使用 `--tmux` 时，挂起后**不保证**远程进程继续运行。
- 密码会话的 `ait sync` 走内置 SFTP（不用系统 `rsync`/`scp` 的非交互密码）。
- 从本机执行 `ait session switch` 前，若仍在 SSH 中，请先用 `!ait suspend` 保存当前会话，或使用 `--force`（不保存当前状态）。

代理与贡献说明见 [AGENTS.md](AGENTS.md)。
