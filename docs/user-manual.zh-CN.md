# AiT 用户手册

**语言：** [English](user-manual.md) · 简体中文（本文）

本文档描述 AiT 的安装后配置、命令用法、提示符、数据目录与安全约定。架构说明见 [design.zh-CN.md](design.zh-CN.md)。

---

## 快速开始

```bash
# AI API Key（系统密钥环；headless Linux 下为 ~/.local/share/ait/keyring 加密文件）
# 无 D-Bus 的服务器建议：export AIT_KEYRING_FILE_PASSWORD='你的主口令'
./bin/ait config set ai.deepseek.api_key <your-key>
./bin/ait config set ai.minimax.api_key <your-key>   # 可选第二 provider
# 别名仍可用：deepseek.api_key、minimax.api_key

# 登录：完整参数说明见 help
./bin/ait login --help

# 交互登录并持久化会话
./bin/ait login deploy@example.com --alias prod
./bin/ait login -p 22 -i ~/.ssh/id_ed25519 deploy@10.0.0.5
./bin/ait login -J bastion app.internal

# 一次性远程命令（类似 ssh host cmd；默认不持久化至 SQLite）
./bin/ait login deploy@example.com uptime

# 列出、恢复、切换会话
./bin/ait session --list
./bin/ait session --load prod
./bin/ait session --switch staging

# 文件同步（可复用已保存会话）
./bin/ait get -p22 -i~/.ssh/id ./app.log user@deploy:/var/log/app.log
./bin/ait get deploy@10.0.0.5 /var/log/app.log ./app.log
./bin/ait get ./app.log /var/log/app.log --session prod
./bin/ait put -l root ./bin/app user@host:/usr/local/bin/app
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
| `ait shell` | 交互模式（`AiT> `）：`login`、`list`、`help`、`clear` 等；`exit` 退出；TTY：UTF-8 输入，**Ctrl+L** 清屏 |
| `ait login` | 连接主机；`--save` 或 `--alias` 保存会话；`host cmd` 为一次性执行 |
| `ait session --list` | 列出已保存会话 |
| `ait session --load <id\|alias>` | 恢复会话；认证过期须 `--password` |
| `ait session --suspend` | 在 SSH 会话内使用 `!ait suspend` |
| `ait session --switch [id\|alias]` | 切换会话（先挂起当前再加载目标） |
| `ait session --delete <id\|alias>` | 删除单个会话 |
| `ait session --delete all` | 删除全部会话（须确认；`--force` 跳过确认） |
| `ait get [flags] <paths...>` | 从远程下载单个文件；OpenSSH 参数同 `login`（`-p`、`-i`、`-J` 等）；`ait get --help` |
| `ait put [flags] <paths...>` | 上传单个文件；flag/路径规则同 `get` |
| `ait config set/get <key>` | 配置项；`ai.<provider>.api_key` 经密钥环存储 |
| `ait ai [--model deepseek\|minimax]` | 本机 AI REPL（TTY 为 UTF-8 输入）；可选本会话云端 provider 覆盖 |
| `ait ai model list` | 列出 AI provider（ENABLED、可达性、Ollama tags） |
| `ait ai chat [message...]` | 单轮提问；无参数时从 stdin 读取；SSH：`!ait ai chat` / `!! chat` |
| `ait ai provider enable <provider>` | 启用 provider；`ollama` 可选检查/安装（[y/N]） |
| `ait ai provider disable <provider>` | 禁用 provider（不再向其发送 AI 请求） |
| `ait ai ollama status` | 查看 tiered/启用状态、API 可达性、已拉取模型 |
| `ait ai ollama setup` | 启用 tiered 并配置 Ollama（安装前 [y/N] 确认） |
| `ait ai ollama start` | 启动 `ollama serve`（`--wait` 等待 API） |
| `ait ai ollama stop` | 停止本机 `ollama serve` |
| `ait ai ollama disable` | 在配置中禁用 Ollama（`--stop` 可同时停服） |

常用标志：`--tmux`（远程 tmux 保活）、`--port`、`--identity`、`--password`、`--alias`。`session --switch --force` 可在当前会话仍标记为 active 时强制切换（不保存当前 scrollback/cwd）。

---

## 本地 REPL 输入（UTF-8）

在 **macOS、Linux、Windows** 上，交互式 `ait shell` 与 `ait ai` 在终端为 UTF-8 时可输入 **Unicode**（含中文）：

- **Linux / macOS：** `export LANG=zh_CN.UTF-8`（或 `en_US.UTF-8`）；Terminal.app、iTerm2、GNOME Terminal 等。
- **Windows：** 使用 [Windows Terminal](https://aka.ms/terminal)；在设置中启用 UTF-8，或在旧版控制台执行 `chcp 65001`。
- **输入法：** 使用系统 IME 组字，确认后字符进入当前行；AiT 不显示应用内预编辑框。
- **Tab：** 在当前行插入制表符（可用于缩进）；与旧版仅支持 ASCII 的行输入不同。
- **非 TTY：** `echo 你好 | ait ai chat` 无需额外 locale 配置。
- **清屏：** `ait shell` 中输入 `clear` 或 TTY 下按 **Ctrl+L** 清空显示，REPL 不退出。`ait ai` 中 **Ctrl+L** 仅清显示；`clear` 清空对话历史。

SSH 会话（`ait login`）由远程 shell 决定编码；本节针对本机 `AiT>` / `AiT AI>` 行编辑。

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

## AI 模式（云端 + 可选本地 Ollama）

至少配置一个**云端** provider 的 API Key（见上文），除非仅用 tiered Ollama 处理简单问题。本地 AI 可运行 **`ait ai ollama setup`**——选择模型存储目录（`OLLAMA_MODELS`；默认 AiT 数据目录，或 `~/.ollama/models`，或 `--models-dir`）；安装 Ollama 前会询问 `[y/N]`；拒绝仍会写入 tiered 配置并打印手动安装说明。用 **`ait ai ollama status`** 查看服务状态；**`start`** / **`stop`** 控制守护进程；**`disable`**（或 `ait ai provider disable ollama`）停止路由。Ollama 无需密钥环。任何 `run` / `ait` 都须您确认后才会执行。

| 场景 | 进入方式 | 执行位置 |
|------|----------|----------|
| 本机（系统 shell 或 `ait shell`） | `ait ai`、`ait ai --model minimax`，或 shell 内 `ai` | `AiT AI>` — `run` 在本机；`ait` 子命令经确认后跑 CLI |
| SSH 会话内 | `!ait ai` 或 `!!`（可带首问） | `AiT AI user@host>` — `run` 写入远程 PTY |

模式内：`help`、`exit`/`quit`、`clear`、`model`（显示路由含 tiered）、`model list`（同 `ait ai model list`）、`model <deepseek|minimax>`（本会话云端覆盖；`model ollama` 拒绝）、`policy per-cmd`（每条确认）、`policy once`（本会话确认一次）。勿用 `!ait ai` 当 meta 子命令。

启用 tiered 时，短句首轮走本地 Ollama；过长、含代码块、复杂关键词或多轮历史走云端。`--model` / `model <云端>` 本会话跳过 Ollama。

云端 provider 失败（网络、5xx、限流）时，按 `ai.fallback` 或自动 `deepseek` → `minimax` 尝试其他已配置 Key 的 provider。tiered 下 Ollama 失败同样升到云端。

---

## SSH 会话内命令

| 输入 | 作用 |
|------|------|
| `!ait help` / `!ait` | 打印会话内命令与可本地执行的 CLI 说明 |
| `!ait <子命令> …` | 在本地执行 `ait` 子命令；`login`/`load`/`ai` 须先退出 SSH |
| `!ait ai [问题]` | 远程 AI 模式（见上） |
| `!! [问题]` | 同 `!ait ai` |
| `!ait local` | 本地模式：命令在本机执行，SSH 连接保持 |
| `!ait remote` | 退出本地模式，恢复向远程发送输入 |
| `!ait suspend` | 挂起：保存 scrollback 与远程 cwd，断开连接 |
| `!ait get <remote> <local>` | 经当前 SSH 下载单个文件（不重新认证） |
| `!ait put <local> <remote>` | 经当前 SSH 上传单个文件（不重新认证） |

---

## 数据与配置目录

| 平台 | 数据目录（SQLite、`scrollback/`） | 配置目录 |
|------|-----------------------------------|----------|
| Linux | `~/.local/share/ait/` | `~/.config/ait/config.yaml` |
| macOS | `~/Library/Application Support/ait/` | 同上 |
| Windows | `%LOCALAPPDATA%\ait\` | `%LOCALAPPDATA%\ait\config.yaml` |

Deepseek API Key 与 SSH 密码缓存存于**操作系统密钥环**，不以明文写入 SQLite。无桌面密钥环时，加密文件位于 `<DataDir>/keyring`（Linux 默认 `~/.local/share/ait/keyring`）。headless 生产环境请设置 `AIT_KEYRING_FILE_PASSWORD`。

---

## 限制说明

- 未使用 `--tmux` 时，挂起后不保证远程进程继续运行。
- `ait get` / `ait put` 仅支持单个文件（不支持目录）。
- `ait put` 会在上传前在远程自动创建**父目录**（若不存在）。
- 传输失败时错误会带 `local file` 或 `remote file` 前缀，便于区分本机与远端。**put** 若出现 `file does not exist` / `no such file or directory`，常见原因是远程路径或父目录不存在，而非本机文件缺失。
- 独立 `get`/`put`：flag 与路径顺序任意；支持粘连写法（`-p22`、`-i~/.ssh/id`）。
- SSH 会话内 `!ait get` / `!ait put` 复用当前连接（仅两个路径，无 flag）。
- 从本机执行 `ait session --switch` 前，若仍处于 SSH 会话中，须先 `!ait suspend` 保存当前状态，或使用 `--force`（不保存当前 scrollback/cwd）。

---

## 延伸阅读

| 文档 | 说明 |
|------|------|
| [README.zh-CN.md](../README.zh-CN.md) | 项目概览与安装 |
| [design.zh-CN.md](design.zh-CN.md) | 架构与行为 |
| [AGENTS.zh-CN.md](../AGENTS.zh-CN.md) | 代码结构 |
| [CONTRIBUTING.zh-CN.md](../CONTRIBUTING.zh-CN.md) | 贡献指南 |
