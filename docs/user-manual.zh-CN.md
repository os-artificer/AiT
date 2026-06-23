# AiT 用户手册

**语言：** [English](user-manual.md) · 简体中文（本文）

本文档描述 AiT 的安装后配置、命令用法、提示符、数据目录与安全约定。架构说明见 [design.zh-CN.md](design.zh-CN.md)。

---

## 快速开始

```bash
# AI API Key（系统密钥环；headless Linux 下为 <AppRoot>/keyring 加密文件）
# 首次配置：./bin/aitssh config init
# 无 D-Bus 的服务器建议：export AIT_KEYRING_FILE_PASSWORD='你的主口令'
./bin/aitssh config set ai.deepseek.api_key <your-key>
./bin/aitssh config set ai.minimax.api_key <your-key>   # 可选第二 provider
# 别名仍可用：deepseek.api_key、minimax.api_key

# 登录：完整参数说明见 help
./bin/aitssh --help

# 交互登录并持久化会话
./bin/aitssh deploy@example.com --alias prod
./bin/aitssh -p 22 -i ~/.ssh/id_ed25519 deploy@10.0.0.5
./bin/aitssh -J bastion app.internal

# 一次性远程命令（类似 ssh host cmd；默认不持久化至 SQLite）
./bin/aitssh deploy@example.com uptime

# 列出、恢复、切换会话
./bin/aitssh session --list
./bin/aitssh session --list -v    # 含 FULL_ID，便于脚本
./bin/aitssh session --load       # 按列表 # 交互选择
./bin/aitssh session --load prod
./bin/aitssh session --switch staging

# 文件同步（可复用已保存会话）
./bin/aitssh get -p22 -i~/.ssh/id ./app.log user@deploy:/var/log/app.log
./bin/aitssh get deploy@10.0.0.5 /var/log/app.log ./app.log
./bin/aitssh get ./app.log /var/log/app.log --session prod
./bin/aitssh put -l root ./bin/app user@host:/usr/local/bin/app
```

配置模板（可选）：

```bash
./bin/aitssh config init
# 或手动：mkdir -p ~/.local/share/AiT && cp configs/ait.example.yaml ~/.local/share/AiT/config.yaml
```

---

## 命令参考

| 命令 | 说明 |
|------|------|
| `aitssh shell` | 交互模式（`AiT> `）：内置（`ll` 本机长列表，仅 shell）→ `<AppRoot>/bin` → aitssh CLI → PATH；`exit` 退出；TTY：UTF-8，**Ctrl+L** 清屏 |
| `aitssh [flags] destination` | 与 OpenSSH 同形连接；`--save`/`--alias` 保存；无 flag 重连仅更新**无 alias** 匹配行；`host cmd` 一次性执行；**无 `login` 前缀** |
| `aitssh session --list` | 列出已保存会话（`#` 列；`-v` 输出完整 UUID） |
| `aitssh session --load [<id\|alias>]` | 恢复会话（可省略 id 交互选择）；请用 alias 或 `-v` 的 FULL_ID，勿用 ID 列短码 |
| `aitssh session --suspend` | 在 SSH 会话内使用 `!ait suspend` |
| `aitssh session --switch [id\|alias]` | 切换会话（当前会话仍为 `active` 时阻断；须先 `!ait suspend` 或 `--force`） |
| `aitssh session --delete <id\|alias>` | 删除单个会话 |
| `aitssh session --delete all` | 删除全部会话（须确认；`--force` 跳过确认） |
| `aitssh get [flags] <paths...>` | 从远程下载单个文件；OpenSSH 参数同默认 SSH 连接（`-p`、`-i`、`-J` 等）；`aitssh get --help` |
| `aitssh put [flags] <paths...>` | 上传单个文件；flag/路径规则同 `get` |
| `aitssh config init [--dry-run]` | 创建或补全 `config.yaml`（仅缺失项）；`--dry-run` 仅打印计划变更 |
| `aitssh config view` | 打印 `config.yaml` 到 stdout（缺失时自动 init） |
| `aitssh config edit` | 用 `$VISUAL`/`$EDITOR` 编辑（Unix 默认 vi，Windows 默认 Notepad）；退出后自动校验 |
| `aitssh config validate` | 校验 YAML 与配置项；逐条打印错误原因；失败 exit 1 |
| `aitssh config set/get <key>` | 配置项；`ai.<provider>.api_key` 经密钥环存储 |
| `aitssh ai [--model deepseek\|minimax]` | 本机 AI REPL（TTY 为 UTF-8 输入）；可选本会话云端 provider 覆盖 |
| `aitssh ai model list` | 列出 AI provider（ENABLED、可达性、Ollama tags） |
| `aitssh ai ask [message...]` | 单轮提问；无参数时从 stdin 读取；SSH：`!ait ai ask` / `!! ask` |
| `aitssh ai agent [message...]` | 有消息时为单轮 agent；SSH 内无参 `!ait ai agent` / `!! agent` 进入 Agent REPL（同 `!ait ai` / `!!`） |
| `aitssh ai provider enable <provider>` | 启用 provider；`ollama` 可选检查/安装（[y/N]）；标志 `--yes`、`--no-install`、`--no-pull` |
| `aitssh ai provider disable <provider>` | 禁用 provider（不再向其发送 AI 请求） |
| `aitssh ai ollama status` | 查看 tiered/启用状态、API 可达性、已拉取模型 |
| `aitssh ai ollama setup [--dry-run]` | 启用 tiered 并配置 Ollama（安装前 [y/N]）；`--dry-run` 仅打印计划步骤 |
| `aitssh ai ollama start` | 启动 `ollama serve`（`--wait` 等待 API） |
| `aitssh ai ollama stop` | 停止本机 `ollama serve` |
| `aitssh ai ollama disable` | 在配置中禁用 Ollama（`--stop` 可同时停服） |

常用标志：`--tmux`（远程 tmux 保活）、`--port`、`--identity`、`--password`、`--alias`。`session --switch --force` 可在当前会话仍标记为 active 时强制切换（不保存当前 scrollback/cwd）。

---

## 本地 REPL 输入（UTF-8）

在 **macOS、Linux、Windows** 上，交互式 `aitssh shell` 与 `aitssh ai` 在终端为 UTF-8 时可输入 **Unicode**（含中文）：

- **Linux / macOS：** `export LANG=zh_CN.UTF-8`（或 `en_US.UTF-8`）；Terminal.app、iTerm2、GNOME Terminal 等。
- **Windows：** 使用 [Windows Terminal](https://aka.ms/terminal)；在设置中启用 UTF-8，或在旧版控制台执行 `chcp 65001`。
- **输入法：** 使用系统 IME 组字，确认后字符进入当前行；AiT 不显示应用内预编辑框。
- **Tab：** 在当前行插入制表符（可用于缩进）。
- **TTY 快捷键（本机 REPL）：** Ctrl+A/E（行首/行尾）、Left/Right（移动光标）、Up/Down（召回已加载的输入历史，含跨会话 `<AppRoot>/repl_history`）、**Ctrl+R**（输入历史反向增量搜索）、Ctrl+U/K/W（删行/删至行尾/删词）、Backspace、Ctrl+C（取消当前行）、Ctrl+D（空行退出）、Ctrl+L（清屏保留缓冲）。在 `aitssh shell` / `aitssh ai` 中输入 `help` 查看完整 Keys 列表。
- **历史：** 仅主提示符 `AiT>` / `AiT AI(Agent)>` 下 Enter 提交的命令会进入历史，不含 `[y/N]` 确认、选 session 输入或 AiT 内部 env 设置。使用 **`history`** 内置命令（bash 子集：列出、`history n`、`-c`、`-d`、`-s`、`-a`、`-r`、`-w`）管理输入历史；跨会话持久化于 **`<AppRoot>/repl_history`**（见 **数据目录**）。`aitssh ai` 中 **`clear`** 仅清对话历史；**`history`** 为输入行历史。
- **非 TTY：** `echo 你好 | aitssh ai ask` 无需额外 locale 配置。
- **清屏：** `aitssh shell` 中输入 `clear` 或 TTY 下按 **Ctrl+L** 清空显示，REPL 不退出。`aitssh ai` 中 **Ctrl+L** 仅清显示；`clear` 清空对话历史。

SSH 会话（顶层 `aitssh user@host` 或 `session --load`）与 `!ait local` 由远端/本地 shell 提供行编辑；本节针对本机 `AiT>` / `AiT AI(Agent)>`。

## 提示符

| 场景 | 提示符 |
|------|--------|
| `aitssh shell` | `AiT> ` |
| `aitssh ai`（本机） | 默认 `AiT AI(Agent)> `；`mode ask` 后为 `AiT AI(Ask)> ` |
| SSH 远程 | `AiT user@host:目录[:git分支]> `（`session.remote_prompt.path`：`short` / `absolute`；Git 分支可选） |
| SSH 内层 | 会话内 `ssh` 到更深层后同样显示 AiT 提示符 | 或 `aitssh -J jump,final` 一次登录到目标主机 |
| SSH AI 模式（`!ait ai` / `!!`） | `AiT AI(Ask) user@host> ` / `AiT AI(Agent) user@host> ` |
| `!ait local` | `AiT Local> ` |

**`aitssh user@host`** 或 **`aitssh session --load`** 连接成功后，远程 **`AiT user@host:…>`** 提示符应直接显示，无需额外按 Enter。**`session --load` / `--switch`** 不会在屏幕上回放挂起前的命令输出；已保存 scrollback 仅写入内存，供 suspend/AI 上下文使用。

---

## 安全与 SSH

默认校验 `~/.ssh/known_hosts`；新主机须先执行 `ssh-keyscan -H <host> >> ~/.ssh/known_hosts`。开发环境可设置 `AIT_SSH_INSECURE=1` 跳过校验（禁止用于生产环境）。`--password` 将出现在进程列表中；生产环境优先使用操作系统密钥环或交互式凭据输入。

---

## AI 模式（云端 + 可选本地 Ollama）

至少配置一个**云端** provider 的 API Key（见上文），除非仅用 tiered Ollama 处理简单问题。本地 AI 可运行 **`aitssh ai ollama setup`**——选择模型存储目录（`OLLAMA_MODELS`；默认 AiT 数据目录，或 `~/.ollama/models`，或 `--models-dir`）；安装 Ollama 前会询问 `[y/N]`；拒绝仍会写入 tiered 配置并打印手动安装说明。用 **`aitssh ai ollama status`** 查看服务状态；**`start`** / **`stop`** 控制守护进程；**`disable`**（或 `aitssh ai provider disable ollama`）停止路由。Ollama 无需密钥环。默认 **Agent** 模式下，**只读命令**（如 `ls`、`git status`）自动执行；写命令与高危及 Ask 模式仍须确认。

| 场景 | 进入方式 | 执行位置 |
|------|----------|----------|
| 本机（系统 shell 或 `aitssh shell`） | `aitssh ai`、`aitssh ai --model minimax`，或 shell 内 `ai` | 默认 `AiT AI(Agent)>`（`mode ask` → `AiT AI(Ask)>`）— `run` 在本机；`aitssh` 经确认后跑 CLI |
| SSH 会话内 | `!ait ai` 或 `!!`（可带首问） | 默认 `AiT AI(Agent) user@host>` — `run` 写入远程 PTY |

写命令或高危 `run` / Agent 执行前有多行确认：**`y`** 仅本次，**`c`** 信任本会话内同**首词**命令，**`s`** 信任本会话所有低危命令，**`n`** 跳过。高危仅 **`y`** / **`n`**。Agent 模式下只读命令（如 `ls`、`git status`）无需确认即可执行。

模式内：`help`、`exit`/`quit`、`clear`、`mode ask|agent`（默认 **agent**；**`mode ask`** 仅问答）、`approve each|session`（默认 each；`trust` 同 session）、`model` / `model list` / `model <deepseek|minimax>`，自然语言（ask 问答 / agent 自动规划）或 `run` / 本机 `aitssh`。`policy` 已弃用（请用 `approve`）。勿用 `!ait ai` 当 meta 子命令。

启用 tiered 时，短句首轮走本地 Ollama；过长、含代码块、复杂关键词或多轮历史走云端。`--model` / `model <云端>` 本会话跳过 Ollama。

云端 provider 失败（网络、5xx、限流）时，按 `ai.fallback` 或自动 `deepseek` → `minimax` 尝试其他已配置 Key 的 provider。tiered 下 Ollama 失败同样升到云端。

---

## SSH 会话内命令

会话内 meta 前缀为 **`!ait`**，不是 `!aitssh`；独立 CLI/二进制名为 **`aitssh`**。

| 输入 | 作用 |
|------|------|
| `!ait help` / `!ait` | 打印会话内命令与可本地执行的 CLI 说明 |
| `!ait <子命令> …` | 在本地执行 `aitssh` 子命令；ssh 形 `[flags] dest` / `load` / `session --switch` 可在 SSH 内链式切换（见下） |
| `!ait [flags] dest` | 必要时自动落库 → 挂起当前会话 → 登录新主机 |
| `!ait load [id\|alias]` | 必要时自动落库 → 挂起 → 加载已保存会话 |
| `!ait session --switch [id\|alias]` | 必要时自动落库 → 挂起 → 切换到另一会话 |
| `!ait ai [问题]` | 远程 AI 模式（见上） |
| `!! [问题]` | 同 `!ait ai` |
| `!ait local` | 本地模式：命令在本机执行，SSH 连接保持 |
| `!ait remote` | 退出本地模式，恢复向远程发送输入 |
| `!ait suspend` | 必要时自动落库 → 挂起：保存 scrollback 与远程 cwd，断开连接 |
| `!ait get <remote> [<local>]` | 经当前 SSH 下载单个文件（不重新认证）；省略 local 时默认为 `./basename` |
| `!ait put <local> [<remote>]` | 经当前 SSH 上传单个文件（不重新认证）；省略 remote 时默认为 `./basename` |
| `!ait chain [on\|off]` | 切换链式模式：PTY 管道 ZMODEM 用于多跳 SSH；`on` = lrzsz ZMODEM，`off` = 优先 lrzsz 再 SFTP（默认） |

---

## 数据与配置目录

| 平台 | 默认 `AppRoot`（config、数据库、scrollback、历史等） |
|------|------------------------------------------------------|
| Linux | `~/.local/share/AiT/` |
| macOS | `~/Library/Application Support/AiT/` |
| Windows | `%LOCALAPPDATA%\AiT\` |

可通过 **`AIT_HOME`** 或 **`AIT_DATA_DIR`** 覆盖。Linux 从旧版升级时，旧路径 `$XDG_CONFIG_HOME/ait/config.yaml`（通常为 `~/.config/ait/config.yaml`）会在首次读取时自动迁入 `AppRoot`。安装后建议执行 **`aitssh config init`** 补全缺失配置项。

Deepseek API Key 与 SSH 密码缓存存于**操作系统密钥环**，不以明文写入 SQLite。无桌面密钥环时，加密文件位于 `<AppRoot>/keyring`。headless 生产环境请设置 `AIT_KEYRING_FILE_PASSWORD`。

---

## 错误提示（Try:）

许多 CLI 错误会附带 **`Try:`** 建议命令块（见 `internal/pkg/errhint`）：

| 场景 | 典型提示 |
|------|----------|
| `--load` 使用列表 **ID 列**（8 位前缀） | `The list ID column is for display only` → 请用 alias、`-v` 的 FULL_ID 或交互式 `--load` |
| 当前会话仍为 **active** 时本机 `session --switch` | 在 SSH 内 `!ait suspend` 或 `session --switch --force`；SSH 内可用 `!ait session --switch` 自动挂起 |
| 空 `aitssh ai ask`（无消息/stdin） | `ask requires a message` → 传入消息或管道 stdin |
| 未知配置键 | `aitssh config get <key>` 或参考示例 YAML |
| `aitssh config validate` 失败 | 按字段路径修正或 `aitssh config edit` |
| `validate` 时无配置文件 | 先执行 `aitssh config init` |
| `edit` 无可用编辑器 | 设置 `EDITOR`（Windows：`set EDITOR=code --wait`） |

---

## 限制说明

- 未使用 `--tmux` 时，挂起后不保证远程进程继续运行。
- `aitssh get` / `aitssh put` 仅支持单个文件（不支持目录）。
- `aitssh put` 会在上传前在远程自动创建**父目录**（若不存在）。
- 传输失败时错误会带 `local file` 或 `remote file` 前缀，便于区分本机与远端。**put** 若出现 `file does not exist` / `no such file or directory`，常见原因是远程路径或父目录不存在，而非本机文件缺失。
- 独立 `get`/`put`：flag 与路径顺序任意；支持粘连写法（`-p22`、`-i~/.ssh/id`）。
- SSH 会话内 `!ait get` / `!ait put` 复用当前连接（一至两个路径，无 flag）；省略的目标路径默认为源文件 basename。
- **`!ait put` 的本机路径**指运行 AiT 的机器（笔记本/桌面），不是远程 shell 当前目录。**`!ait get` 的本机路径**同样在 AiT 客户端。远程路径相对于**当前 active shell**（链式模式下可能是嵌套跳板）。
- SSH 会话内 **`!ait get` / `!ait put` 会在 stderr 输出分阶段进度与 `done:` 完成行**（如 `get … → …`、`handshake…`、`copying… 50%`，SFTP 在 100% 后短暂 **`closing…`**，链式 ZMODEM 为 **`finishing…`**，再 `done: get …`）。TTY 上 **`copying… N%` 在同一行内刷新**（get/put、SFTP 与链式 ZMODEM 文案一致）。阶段行与 done 行在 raw TTY 下使用 CRLF（Linux/macOS/Windows 客户端均适用）。
- 从本机执行 `aitssh session --switch` 前，若仍处于 SSH 会话中，须先 `!ait suspend` 保存当前状态，或使用 `--force`（不保存当前 scrollback/cwd）。

---

## 代理链式模式

**login 时使用 ProxyJump / `-J`** 经 `DialChain` 到达最终主机；该登录下的会话内 `!ait get`/`put` **优先探测 lrzsz**，缺失时**回退 SFTP**，与独立 `aitssh get`/`put`（仅 SFTP）一致。**链式模式**适用于**交互 shell 比拨号连接更深**的情况（通常在会话内手动 `ssh` 之后）。

当你 login 到堡垒机再手动 `ssh` 到内网主机时，AiT 的直连仅到达拨号目标。SFTP 默认只能对该目标生效。使用 PTY 管道的功能（`!!`、`!ait ai`、scrollback）在任何深度均可正常工作。

当 chain 关闭且位于**拨号目标同 host** 时，`!ait get`/`put` **使用 SFTP**（即使远端已安装 lrzsz）。**仅链式模式**在检测到 `sz`/`rz` 时使用 **PTY ZMODEM**。直连且无 lrzsz 时**回退 SFTP**（带提示）。**手动 SSH 到更深层** 时，**自动启用链式模式**（主机名与拨号主机及登录别名均不匹配时）。

**`!ait chain on`** 可手动启用链式模式，切换为 PTY 管道注入：

- **文件传输**：**优先 lrzsz** — 有 `sz`/`rz` 时经 PTY 跑 ZMODEM。链式且无 lrzsz 时，AiT 给出安装指引；交互 TTY 下询问 **`Install lrzsz on <host>? [y/N]`**，授权后在 active shell 执行 OS 对应安装命令（Linux 包管理器、Homebrew、choco/winget；可能需要远端 `sudo`）。进度含 `handshake…`、**`copying… NN%`**（与 SFTP 相同；TTY 上行内更新），100% 后 **`finishing…`** 再 `done:`（get/put 对称）。AiT 客户端无需本地 lrzsz。独立 `aitssh get/put` 仍走 SFTP（100% → **`closing…`** → `done:`）。
- **tmux**：tmux 操作通过 PTY 管道注入，而非新开 SSH 会话。

**`!ait chain off`** 返回默认 SFTP/直连会话模式。

**`!ait chain`**（无参数）显示当前链式模式状态。

所有 AiT 注入的命令均为静默——不会出现在 shell 历史记录或方向键导航中。

```text
AiT user@bastion> ssh internal-server
…经堡垒机登录到内网服务器…
AiT user@bastion> !ait get /var/log/app.log ./app.log
  chain mode auto-enabled: active shell is on internal-server (connected to bastion).
  File transfer will use PTY-pipe injection.
  Use !ait chain off to disable.
AiT user@bastion> !ait chain   # 显示 on
AiT user@bastion> !ait chain off
  chain mode disabled
  File transfer and tmux will use SFTP/direct sessions.
```

**链式模式对 active shell 主机的要求**：ZMODEM 传输需要 **lrzsz**（`sz`/`rz`）。链式缺失时可手动安装或批准 `[y/N]` 自动安装。v1 对 Windows OpenSSH 远端自动安装支持有限。

---

## 延伸阅读

| 文档 | 说明 |
|------|------|
| [README.zh-CN.md](../README.zh-CN.md) | 项目概览与安装 |
| [design.zh-CN.md](design.zh-CN.md) | 架构与行为 |
| [AGENTS.zh-CN.md](../AGENTS.zh-CN.md) | 代码结构 |
| [CONTRIBUTING.zh-CN.md](../CONTRIBUTING.zh-CN.md) | 贡献指南 |
