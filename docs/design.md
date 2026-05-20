
# AiT：AI 增强 SSH 客户端实现计划

## 现状

项目 CLI 入口：[`internal/ait/root.go`](internal/ait/root.go)（`Run` + cobra root）。依赖见 [`go.mod`](go.mod)（cobra、modernc sqlite、`golang.org/x/crypto/ssh`、`creack/pty`、go-openai、keyring、pkg/sftp 等）。目录约定：`cmd/ait` 薄入口 → `internal/ait` + `internal/pkg`。

## 目标 CLI 形态

```text
ait login [ssh-options] destination [command ...] [--alias NAME] [--password] [--tmux]
        # OpenSSH 兼容：-p -l -i -F -J -o 等；destination 为 [user@]host 或 config Host 别名
        # 需本机 OpenSSH 客户端 (ssh -G) 解析 config；数据面为 Go x/crypto/ssh（Win/macOS/Linux）
        # 带 command 时为一次性执行（类似 ssh host cmd），默认不写 SQLite
        # 旧写法 --host/--user 仍兼容
ait list
ait load <id|alias> [--tmux]  # 7 天内免密恢复；过期后提示输入 SSH 密码/密钥口令
ait suspend                   # 交互中挂起；--tmux 时 detach 远程 tmux
ait session switch [id|alias] # 切换到另一会话；无参数时交互选择
ait session delete all        # 删除全部已保存会话及本地 sidecar
ait session delete <id|alias> # 删除指定会话
ait sync <src> <dst> [--session ID]
ait config set|get ...        # 含 session.auth_ttl（默认 168h = 7 天）
```

交互 SSH 内：以 **`!ait ai <问题>`**（推荐）或 **`!! <问题>`**（兼容别名）唤起 Deepseek；**`!ait local` / `!ait remote`** 切换本地/远程输入模式（SSH 不断开）；**`!ait suspend`** 挂起。未配置 API Key 时，行首 `!!` / `!ait` 在本地 fail-closed，不进入远程 shell。详见下文 [会话内 meta 协议](#会话内-meta-协议ait--) 与 [本地模式](#本地模式ait-local--remote)。

### 两种恢复模式（默认 vs `--tmux`）

| 模式 | 触发 | load 行为 | 远程进程 |
|------|------|-----------|----------|
| **默认（书签）** | 无 `--tmux` | 用已存连接信息重连；可选回放上次 suspend 的本地 scrollback | 不保证保活 |
| **tmux 保活** | `login`/`load` 加 `--tmux` | SSH 后 `tmux attach -t ait-<session_id>`（不存在则 `new -s`） | 断开连接后 shell/进程仍在 |

**默认仅持久化登录信息**（host/port/user/identity/alias、认证过期时间）；scrollback/`remote_cwd` 仅在 suspend 时作为增强项写入，不作为「真正保活」承诺。真正保活必须显式 `--tmux`。

## 架构总览

```mermaid
flowchart TB
  subgraph cli [internal/ait]
    Run[Run args]
    Commands[login list load suspend sync config session]
  end
  subgraph pkgs [internal/pkg]
    Store[store sqlite]
    Sess[session manager]
    SSH[sshclient PTY]
    Term[terminal scrollback state]
    AI[ai deepseek]
    Plat[platform paths exec probe]
    Sync[sync rsync sftp]
  end
  Run --> Commands
  Commands --> Sess
  Sess --> Store
  Sess --> SSH
  Sess --> Term
  Commands --> AI
  Commands --> Sync
  Sync --> Plat
  AI --> Plat
  SSH --> Term
```

## 核心设计决策

| 主题 | 方案 | 说明 |
|------|------|------|
| CLI 框架 | [spf13/cobra](https://github.com/spf13/cobra) | 子命令清晰，与 `Run(args)` 对接 |
| SQLite | [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) | 纯 Go、无 CGO，利于 Windows 交叉编译 |
| SSH | `golang.org/x/crypto/ssh` + `RequestPty` | 标准库生态，跨平台 |
| 本地 PTY | Unix: `github.com/creack/pty`；Windows: ConPTY（`golang.org/x/sys/windows` 或成熟封装） | 分文件 `pty_unix.go` / `pty_windows.go` |
| Deepseek | OpenAI 兼容 HTTP（`https://api.deepseek.com`） | 可用轻量 `net/http` 或 `github.com/sashabaranov/go-openai` 设 `BaseURL` |
| 密钥/密码 | 私钥路径入 SQLite；SSH 密码/密钥口令经 keyring 缓存；**认证 TTL 默认 7 天** | 过期后 `load` 必须重新输入 |
| SSH 主机密钥 | 默认 `~/.ssh/known_hosts`；未知主机报错并提示 `ssh-keyscan` | `AIT_SSH_INSECURE=1` 仅开发用 |
| 系统命令 | `exec.LookPath` + `runtime.GOOS` | 缺失时 Deepseek 安装建议 |
| 远程保活 | 仅 `--tmux` 时启用 | 调用系统 `tmux`（远程需已安装，缺失则 AI 建议安装） |

### 会话语义与 7 天认证 TTL

**`login` / `load` 与认证 TTL：**

- **`login` 成功**：写入 `auth_expires_at = now + session.auth_ttl`（默认 **168h / 7 天**）
- **`load` / `switch` 免密重连**：在 TTL 内使用 keyring 缓存，**不刷新** `auth_expires_at`
- **重新输入密码后**（`--password` 或交互提示，且 dial 成功）：刷新 `auth_expires_at`
- **超过 TTL**：`load` / `switch` 前必须重新输入 SSH 密码；验证成功后刷新 TTL

**挂起时持久化：**

- **必有**：连接配置 + `auth_expires_at` + `status`
- **可选（suspend 时）**：本地 `cols`/`rows`、scrollback sidecar、`remote_cwd`（便于默认模式下的体验，非 tmux 保活）
- **`--tmux` 时额外**：`tmux_session_name`（固定 `ait-<session_id>`）、`tmux_enabled` 标志

**默认 `load`（无 `--tmux`）：**

1. 检查 TTL → 过期则要求密码/口令
2. SSH 重连
3. 若有 `session_state`：回放 scrollback、`cd` 到 `remote_cwd`（有则执行）
4. 进入交互

**`load --tmux`：**

1. TTL 检查同上
2. SSH 连接 → 检测远程 `tmux has-session -t ait-<id>` → attach，否则 `tmux new-session -s ait-<id>`
3. 进入 tmux 前台（suspend 时对已存在会话执行 `detach`）

**限制说明：** 未加 `--tmux` 时，挂起仍断开 SSH，**不保证**远程进程继续运行；加 `--tmux` 且远程有 tmux 时，才实现真正会话保持。

### 会话切换（`session switch`）

在多个已保存会话之间切换，无需手动 `suspend` 再 `load`。

| 场景 | 行为 |
|------|------|
| 本地无前台 SSH | 等同 `load <target>`（TTL/密码规则相同） |
| 本地有前台 SSH | 先 **suspend 当前**（写 `session_state`、tmux detach）→ 再 **load 目标**；`--force` 仅标记当前为 suspended，**不**保存 scrollback/cwd |
| 未指定目标 | `list` 展示编号菜单，用户选择 id/alias（支持 `--tmux` 继承当前连接模式或按目标会话 DB 中 `tmux_enabled`） |

```text
ait session switch              # 交互式选择
ait session switch prod         # 按 alias/id 切换
ait session switch prod --tmux  # 切换并以 tmux 模式 attach 目标
```

**交互 SSH 内：** 会话切换以 CLI `ait session switch` 为主；挂起使用 `!ait suspend`。

**与 `load` 区别：** `load` 仅连接目标；`switch` 保证当前会话先挂起保存，避免丢失当前 scrollback/cwd，并更新 `app_meta.current_session_id`。

**tmux 模式切换：** 当前会话 `detach` → 目标会话 SSH → `tmux attach -t ait-<target_id>`；两台主机上的 tmux 会话互不影响。

## 数据模型（SQLite）

文件位置（[`internal/pkg/platform`](internal/pkg/platform)）：

- Linux: `~/.local/share/ait/ait.db`
- macOS: `~/Library/Application Support/ait/ait.db`
- Windows: `%LOCALAPPDATA%\ait\ait.db`

```sql
CREATE TABLE sessions (
  id              TEXT PRIMARY KEY,
  alias           TEXT UNIQUE,
  host            TEXT NOT NULL,
  port            INTEGER NOT NULL DEFAULT 22,
  username        TEXT NOT NULL,
  identity_path   TEXT,
  auth_method     TEXT NOT NULL,  -- key | password | agent
  auth_expires_at INTEGER NOT NULL,  -- Unix 秒；load 免密截止
  tmux_enabled    INTEGER NOT NULL DEFAULT 0,
  tmux_session    TEXT,           -- 默认 ait-<id>
  created_at      INTEGER NOT NULL,
  updated_at      INTEGER NOT NULL,
  status          TEXT NOT NULL   -- active | suspended
);

CREATE TABLE session_state (
  session_id        TEXT PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
  cols              INTEGER,
  rows              INTEGER,
  remote_cwd        TEXT,
  scrollback_path   TEXT,
  suspended_at      INTEGER
);

CREATE TABLE app_meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
-- 键 current_session_id：当前/上次前台会话 UUID（switch/load/login 后更新）
```

- `login` 成功 → upsert `sessions`，设置 `auth_expires_at`；若 `--tmux` 则 `tmux_enabled=1` 并初始化远程 tmux；写 `app_meta.current_session_id`
- `suspend` → 写 `session_state`，`status=suspended`；tmux 模式先 `tmux detach` 再断 SSH
- `session delete <id>` / `session delete all` → 删 DB 行 + 删除 scrollback sidecar；tmux 模式可选 `--kill-tmux` 远程杀会话（`delete all` 默认不杀远程 tmux，避免误删；文档说明）

**凭证存储：** SSH 密码不落明文 SQLite；7 天内有期缓存放 **OS keyring**（key: `ait/session/<id>/unlock`），过期自动要求重输。私钥仍只存路径。

## 包布局（新增）

| 路径 | 职责 |
|------|------|
| [`internal/ait/`](internal/ait/) | `Run` → cobra root；注册子命令 |
| `internal/ait/cmd_*.go` | `login`, `list`, `load`, `suspend`, `sync`, `config`, `session` 命令 |
| `internal/pkg/tmux/` | 远程 tmux 探测、`new-session` / `attach` / `detach` / `has-session` 封装 |
| `internal/pkg/store/` | SQLite 打开、迁移、CRUD |
| `internal/pkg/session/` | 会话生命周期：创建、挂起、加载、**切换**、列表 |
| `internal/pkg/sshclient/` | 拨号、认证、PTY 会话、交互循环 |
| `internal/pkg/terminal/` | scrollback 缓冲、尺寸、回放、原始模式 stdin |
| `internal/pkg/ai/` | Deepseek 客户端、`SuggestInstall(os, cmd)`、`Assist(ctx, prompt)` |
| `internal/pkg/sync/` | `rsync`/`scp` 包装；无则 SFTP 回退 |
| `internal/pkg/platform/` | 数据目录、`LookPath`、OS 描述字符串 |
| `configs/ait.example.yaml` | 可选：默认模型、`deepseek_base_url`（密钥仍走 keyring） |

`pkg/` 暂不放公开 API，除非后续需要被其他项目导入。

## 关键流程

### 提示符（AiT> / AiT Local> / AiT user@host>）

| 场景 | 提示符 | 设置方式 |
|------|--------|----------|
| `ait shell`（未连远程） | `AiT> ` | REPL 打印 [`BasePrompt`](internal/pkg/sshclient/prompt.go) |
| SSH 远程 | `AiT <user>@<host> ` | [`RemotePromptInitialCommands`](internal/pkg/sshclient/prompt.go) 在 `load`/`login` 后注入远程 shell |
| SSH 本地模式 | `AiT Local> ` | 静默 rc / env（[`localmode.go`](internal/pkg/sshclient/localmode.go)） |

**`ait shell`**：行级 REPL，可直接输入 `login`、`load`、`list`、`help` 等（无需 `!ait` 前缀）。在 shell 内 `login` 结束后回到 `AiT> `；系统 shell 执行 `ait login` 结束后回到 OS shell。

### 会话内 meta 协议（!ait / !!）

实现位于 [`internal/pkg/sshclient/metascanner.go`](internal/pkg/sshclient/metascanner.go) 与 [`client.go`](internal/pkg/sshclient/client.go) 的 `forwardStdin` / `handleMetaLine`。已弃用 `\ait` 前缀。

| 命令 | 行为 |
|------|------|
| `!ait help` / `!ait` | 打印会话内命令与可本地执行的 CLI 说明 |
| `!ait <子命令> …` | 在本地执行对应 `ait` CLI（如 `!ait list`）；`login`/`load` 请用 `ait shell` |
| `!ait ai <问题>` | 调用 Deepseek，结果打印到本地终端，不转发远程 |
| `!! <问题>` | 上列别名（`!!` 后须有空格） |
| `!ait local` | 进入本地模式（见下节） |
| `!ait remote` | 退出本地模式，恢复远程 stdin |
| `!ait suspend` | 挂起会话（等同 CLI `suspend`） |

**识别规则：**

- 字节级 `metaScanner`：严格匹配 opener（`!ait ai `、`!ait local`、`!! ` 等）；行内出现 `!` 亦进入 holding，避免远程 zsh 在无换行提示符下吞掉 `!ait`。
- `aiAvailable == false`（未配置 API Key）时 fail-closed：行首 `!!…` / `!ait…` 不进远程；仅**非空** AI 问题才 stderr 提示配置 key；空问题只打印 usage。
- meta 行由本地 echo，不写入远程 PTY；用户可见消息经 `termPrintln` / `aitModeNotice` 使用 `termEOL`（`\r\n`）适配 raw TTY。

### login

```mermaid
sequenceDiagram
  participant User
  participant CLI as ait_login
  participant DB as sqlite
  participant SSH as sshclient
  User->>CLI: ait login user@host ...
  CLI->>DB: upsert session
  CLI->>SSH: Dial + PTY
  loop interactive
    User->>SSH: input
    alt meta line !ait / !!
      SSH->>SSH: handleMetaLine
      opt AI question
        SSH->>CLI: Deepseek assist
        CLI-->>User: suggestion
      opt local / remote / suspend
        SSH->>SSH: local mode or suspend
      end
    else normal input
      SSH->>User: remote IO
    end
  end
```

- **配置解析**：`internal/pkg/sshargs` 调用本机 `ssh -G`（`platform.OpenSSHPath`：Linux/macOS PATH；Windows 另试 `System32\OpenSSH\ssh.exe`）
- **拨号**：`sshclient.DialChain` 支持 `ProxyJump` / `ProxyCommand`（Unix `sh -c`，Windows `cmd /c`）
- **一次性命令**：`ait login host cmd` → `RunRemoteCommand`，退出码透传；默认不写 SQLite（`--alias` 时仍 upsert）
- 认证顺序：`-i` / config `IdentityFile` → SSH agent → `~/.ssh/id_ed25519` / `id_rsa`
- 交互 login 成功即写入/更新会话记录

### 本地模式（!ait local / remote）

SSH 连接与子 shell **保持**；stdin 在本地 PTY 与远程 `stdin` 间切换，远程 stdout 在本地模式下暂停显示（仍写入 scrollback）。

```mermaid
flowchart LR
  User[User stdin] --> fwd[forwardStdin]
  fwd --> branch{localMode active?}
  branch -->|yes| localWrite[localShell.write]
  branch -->|no| remoteIn[sess.StdinPipe]
  remoteOut[remote stdout] --> pause{pauseRemoteOut}
  pause -->|false| termOut[os.Stdout]
  pause -->|true| scrollOnly[scrollback only]
```

| 主题 | 说明 |
|------|------|
| 子进程 | [`localmode.go`](internal/pkg/sshclient/localmode.go)：`creack/pty` 启动 `$SHELL`（Windows 为 `ComSpec`） |
| 提示符 | 常量 `AiT Local> `；bash `PS1` 使用 `\[\]…\[\]`；环境变量 `AIT_LOCAL_MODE=1`、`PS1`、`PROMPT`（Windows cmd 另设 `PROMPT`） |
| 静默设置 | **不向终端打印** `export PS1=…` 等 setup 行：bash 用 `--rcfile` 临时 rc；zsh 用 `ZDOTDIR` 最小 `.zshrc`（source 用户 rc 后再次 `export PROMPT`）；临时目录在 `stop()` 时删除 |
| 生命周期 | `enterLocalMode` / `exitLocalMode` / `onLocalShellExit`；`closing` 抑制预期 ptmx 关闭错误 |
| 尺寸 | [`resize.go`](internal/pkg/sshclient/resize.go)：SIGWINCH 同步本地 PTY 与远程 `WindowChange` |
| 高级 | 用户可在 `.zshrc` 中检测 `AIT_LOCAL_MODE=1` 自定义主题（可能覆盖 `PROMPT`） |

### suspend / list / load

- **suspend**：`!ait suspend`；tmux 模式先 detach；写 `session_state`（可选 scrollback/cwd）→ 关 SSH
- **list**：输出 `id`, `alias`, `host`, `user`, `status`, `tmux`, `auth_expires_at`（剩余天数）
- **load**：TTL 校验 →（过期则读密码）→ 重连 → 默认回放+cwd 或 `--tmux` attach → 更新 `current_session_id`
- **switch**：若 `current_session_id` 存在且与目标不同 → `Suspend(current)` → `Load(target)`；目标与当前相同则 no-op 并提示

### session switch 流程

```mermaid
sequenceDiagram
  participant User
  participant CLI as session_switch
  participant Sess as session_mgr
  participant DB as sqlite
  User->>CLI: session switch [target]
  alt no target
    CLI->>DB: list sessions
    CLI-->>User: numbered picker
    User->>CLI: select target
  end
  CLI->>Sess: Switch(target)
  alt current foreground exists
    Sess->>Sess: Suspend(current)
  end
  Sess->>Sess: Load(target)
  Sess->>DB: set current_session_id
```

### session 清理

```text
ait session delete all         # 删除全部 sessions + session_state + sidecar + keyring 条目
ait session delete <id|alias>  # 删除单条；存在则确认提示（--force 跳过）
```

cobra 结构：`ait session delete` 为父命令，`all` 与 `<id|alias>` 为互斥目标（`all` 为固定子参数/子命令，需二次确认或 `--force`）。

- `delete all`：事务内清空两表；遍历数据目录删除 scrollback 文件；批量清除 keyring `ait/session/*`
- `delete <id|alias>`：`ON DELETE CASCADE` 清理 `session_state`；删对应 sidecar 与 keyring 项

### tmux 集成（`--tmux`）

```mermaid
sequenceDiagram
  participant CLI
  participant SSH
  participant Tmux as remote_tmux
  CLI->>SSH: login --tmux
  SSH->>Tmux: has-session ait-ID?
  alt exists
    Tmux-->>CLI: attach
  else
    CLI->>Tmux: new-session -s ait-ID
  end
  Note over CLI,Tmux: suspend 时 tmux detach-client
  CLI->>SSH: load --tmux
  SSH->>Tmux: attach-session -t ait-ID
```

- 远程无 `tmux`：`LookPath` 失败 → `ai.SuggestInstall`（apt/brew 等）
- 会话名确定性：`ait-<session_id>`，避免与用户手工 tmux 冲突
- Windows 本地客户端仍可 `load --tmux`，但远程主机通常为 Linux/macOS

### sync

1. `LookPath("rsync")` → 存在则构造 `rsync -e ssh ...`（会话的 host/user/identity 注入 `-e`）
2. 否则 `scp`；再否则 **内置 SFTP**（`ssh.NewSFTPClient` 或 `pkg/sftp`）
3. 三者皆不可用或执行失败且原因为「命令不存在」→ 调 `ai.SuggestInstall`

路径语义：`user@host:/path` 与本地路径，与 `scp` 一致；`--session` 复用已保存会话的连接参数。

### AI 助手（会话内）

与 [会话内 meta 协议](#会话内-meta-协议ait--) 一致；补充实现细节：

- 上下文：最近 N 行 scrollback（脱敏）、`GOOS`、当前 `remote_cwd`、用户问题
- 结果打印到本地终端（不自动执行，避免误操作）
- **安装建议**：`platform.LookPath(cmd)==err` 时，prompt 包含 OS 版本探测 → 返回 brew/apt/choco/winget 等建议

## 跨平台要点

| 平台 | PTY / 终端 | 数据目录 | 常见缺失命令 |
|------|------------|----------|----------------|
| Linux | creack/pty + raw mode | XDG data home | rsync 包名 `rsync` |
| macOS | 同上 | Application Support | `brew install rsync` |
| Windows | ConPTY；OpenSSH 内置 | LocalAppData | `winget` / OpenSSH 功能 |

构建标签：`//go:build linux || darwin` 与 `windows` 分离 PTY 实现；`make test` 在 Linux CI 全量，Windows 目标至少 `go build` 通过。

## 依赖（拟加入 go.mod）

- `github.com/spf13/cobra`
- `modernc.org/sqlite`
- `golang.org/x/crypto/ssh`
- `golang.org/x/term`
- `github.com/creack/pty`（unix）
- `github.com/99designs/keyring`
- `github.com/google/uuid`
- （可选）`github.com/pkg/sftp` — SFTP 回退
- （可选）`github.com/sashabaranov/go-openai` — Deepseek 兼容

遵循项目规则：**仅在需要时引入**，优先 stdlib + `x/` 扩展。

## 测试策略

- **单元测试**：`store` CRUD（`:memory:` SQLite）、`terminal` scrollback 截断/回放、`ai` mock HTTP
- **集成测试**：`sshclient` 用 `httptest` 或跳过标签；`//go:build integration` + docker testcontainers 可选后续
- 保持 [`internal/ait/app_test.go`](internal/ait/app_test.go)  smoke test 更新为 `Run([]string{"--help"})`

## 实施阶段（建议顺序）

1. **基础 CLI + 配置**：cobra、`platform`、`store` 迁移（含 `auth_expires_at`）、keyring、`session.auth_ttl` 默认 7d
2. **login + 交互 SSH**：sshclient；login 写 sessions + 刷新 TTL
3. **suspend / list / load / switch**：默认书签恢复 + TTL；`switch` = suspend 当前 + load 目标 + 交互选择器
4. **session delete**：`delete all` 与 `delete <id|alias>` 清理 DB、sidecar、keyring
5. **`--tmux` 集成**：`internal/pkg/tmux` + login/load/suspend 分支
6. **sync + AI 助手 + Windows**：rsync/sync、`!!` 助手、ConPTY、README 说明两种模式与 TTL

## 风险与缓解

- **Windows PTY 复杂度高**：第一期可先保证「连接 + 非全屏交互」可用，再完善 ConPTY 尺寸/信号；`go build` Windows 为门禁
- **scrollback 过大**：超过阈值（如 512KB）写入 sidecar 文件，DB 只存 path
- **安全**：不在日志打印 API Key/密码；`!!` 上下文脱敏；keyring 条目随 `delete`/`delete all` 清除
- **TTL 边界**：时钟回拨用单调比较 + 文档说明；`auth_expires_at` 可在 `config` 中覆盖（最小 1h，最大 30d 防误配）

## 交付物

- 可执行 `bin/ait` 支持上述子命令
- [`configs/ait.example.yaml`](configs/ait.example.yaml) 配置模板
- `make test` / `make vet` 通过
- 简短用户说明（可写在根 [`README.md`](README.md) 的「用法」一节，仅在你确认后添加）
