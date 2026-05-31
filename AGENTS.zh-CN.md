# AiT — Agent 导向

**语言：** [English](AGENTS.md) · 简体中文（本文）

**AiT**（AI Terminal）为 Go CLI：具备 SQLite 会话持久化、可选远程 tmux 保活、多 provider AI（Deepseek、MiniMax、可选 tiered Ollama）与 SFTP 文件 get/put 的 AI 增强 SSH 客户端。模块：`github.com/os-artificer/AiT`。目录布局遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout)。

## 构建与测试

```bash
make build       # → bin/ait
make test        # go test ./...
make vet         # go vet ./...
make fmt         # go fmt ./...
make test-cross  # windows + darwin go build
./bin/ait        # 运行二进制
```

需要 **Go 1.25+**（见 [`go.mod`](go.mod)）。Windows 交叉编译：`GOOS=windows GOARCH=amd64 go build -o bin/ait.exe ./cmd/ait`。

## 功能概览

- **SSH 会话**：`shell`、`login`、`session`（`--list`、`--load`、`--suspend`、`--switch`、`--delete`）
- **持久化**：SQLite（`internal/pkg/store`）；挂起时保存 scrollback/工作目录；认证 TTL 由 `session.auth_ttl` 配置（默认 7 天；密钥走 keyring）
- **可选 tmux**：在 `login`/`load`/`switch` 使用 `--tmux` 实现远程会话保活（`internal/pkg/tmux`）
- **AI**：多 provider 路由 + `tiered`；`ait ai chat` 单轮；`ait ai provider enable|disable`；`ait ai ollama`（setup/status/start/stop/disable；`models_dir`/`OLLAMA_MODELS`；安装前 [y/N]）；REPL `model` / `ait ai model list`；SSH `!ait ai chat` / `!! chat`
- **传输**：`ait get` / `ait put` — SFTP 单文件（`internal/pkg/transfer`）；`!ait get` / `!ait put` 复用当前 SSH（统一路由：优先 lrzsz/ZMODEM，直连无 lrzsz 回退 SFTP，链式缺失时可授权自动安装）
- **代理链支持**：`!ait chain [on|off]` — 链式模式下经 PTY 管道 ZMODEM 文件传输与 tmux（`internal/pkg/ssh/client/inject.go`、`transfer_zmodem.go`；`internal/pkg/tmux/remote_pty.go`）
- **OpenSSH 配置**：`login`/`load` 经 `ssh -G` 解析 `~/.ssh/config`（`internal/pkg/ssh/args`）；数据面为 Go `x/crypto/ssh`

完整设计：[`docs/design.zh-CN.md`](docs/design.zh-CN.md)。配置示例：[`configs/ait.example.yaml`](configs/ait.example.yaml)。Cursor 方案归档至 `docs/<kebab-case>.md`（见 [`.cursor/rules/docs-exec-plans.mdc`](.cursor/rules/docs-exec-plans.mdc)）。

## 代码位置

| 路径 | 职责 |
|------|------|
| `cmd/ait/main.go` | 极简 `main` — 委托至 `internal/ait.Run` |
| `internal/ait/` | CLI 入口（`Run`、`ExitCode`）；子包见下表 |
| `internal/ait/cli/` | Cobra 命令树 facade：`Execute` / `ExecuteWriters`、`newRootCmd`、manager 接线（子命令在下表的子包中实现） |
| `internal/ait/cli/ai/` | `ait ai` 命令（`chat`、`provider enable|disable`、本地 AI REPL Runner） |
| `internal/ait/cli/config/` | `ait config` 命令（`set`、`get`、`init`、`view`、`edit`、`validate`） |
| `internal/ait/cli/session/` | `ait session` 命令（`--list`、`--load`、`--suspend`、`--switch`、`--delete`） |
| `internal/ait/cli/transfer/` | `ait get` / `ait put` SFTP 命令 |
| `internal/ait/cli/ollama/` | `ait ai ollama` 子命令（`status`、`setup`、`start`、`stop`、`disable`） |
| `internal/ait/cli/login/` | `ait login` OpenSSH 兼容命令 |
| `internal/ait/helpdoc/` | login 样式 `--help`（`Write*Help`、OpenSSH 共享块）；`Bind` / `BindCustomArgv` Cobra 辅助 |
| `internal/ait/app/` | `NewManager`、`CommandContext`（store/config/ai 引导）、`Runner` |
| `internal/ait/login/` | OpenSSH login 解析/解析、帮助、退出码 |
| `internal/ait/shell/` | `ait shell` REPL（`AiT>`）、TTY I/O、行执行 |
| `internal/ait/meta/` | 会话内 `!ait <subcmd>` 委托 |
| `internal/pkg/ssh/args/` | OpenSSH 参数解析、`ssh -G` 解析、`Resolved` → `client.Options` |
| `internal/pkg/session/` | login、load、suspend、switch、delete、list（`manager.go`、`manager_login.go`、`manager_load.go`、`manager_suspend_switch.go`、`manager_list_pick.go`、`manager_helpers.go`） |
| `internal/pkg/store/` | SQLite schema、迁移、CRUD（`store.go`、`store_sessions.go`、`store_state.go`、`store_meta.go`、`store_scan.go`、`session_helpers.go`、`open_helpers.go`） |
| `internal/pkg/ssh/client/` | `Client`、`Options`、`Dial`、`DialChain`、`RunInteractive` facade、静默 PTY 注入（详见下方文件布局） |
| `internal/pkg/ssh/client/prompt/` | AiT 显示提示符字符串 + 远程 PS1 安装片段 |
| `internal/pkg/ssh/client/scrollback/` | 远程 PTY scrollback `Delta` 与捕获时序常量 |
| `internal/pkg/ssh/client/localmode/` | `!ait local` 本地 PTY `Shell`、环境变量、bash/zsh 启动脚本 |
| `internal/pkg/ssh/client/meta/` | `!ait` / `!!` 字节级状态机扫描器与 meta 常量 |
| `internal/pkg/ssh/client/dial/` | ProxyCommand stdio 连接、ProxyJump 解析、OpenSSH 私钥读取 |
| `internal/pkg/ssh/client/pty/` | TTY stdin 单读者、REPL 字节通道、跨平台 resize watcher |
| `internal/pkg/ssh/client/transfer/` | `!ait get|put` meta 参数解析与 usage 提示 |
| `internal/pkg/tmux/` | 远程 tmux（SFTP 方式 + 链式模式下的 PTY 注入方式） |
| `internal/pkg/terminal/` | terminal facade（`path.go`）；其余通过 type alias 指向子包 |
| `internal/pkg/terminal/scrollback/` | 环形 scrollback `Buffer`（写入/回放/文件 IO） |
| `internal/pkg/terminal/screen/` | ANSI 屏幕原语（清屏） |
| `internal/pkg/terminal/crlf/` | CRLF 规范化的 `io.Writer` |
| `internal/pkg/terminal/lineedit/` | `ReadLine`（Ctrl+A/E、历史导航、Ctrl+R 搜索） |
| `internal/pkg/terminal/history/` | `LineHistory`（内存 + 文件同步）与 `history` 命令 |
| `internal/pkg/ai/` | `Client`、`Chat`、多 provider 路由 facade |
| `internal/ait/aimode/risk/` | `Classify`、`Interactive`、`ClassifyAit` — 命令风险分类 |
| `internal/ait/aimode/approve/` | `Confirm`、`Policy`、`PatternKey`、`SessionTrust` — 确认提示与信任 |
| `internal/ait/aimode/agent/` | `Action`、`ParseAction`、`Run`、`CapturingRunner` — Agent 模式执行 |
| `internal/ait/aimode/chat/` | `RunOnce`、`ParseCommand` — AI chat 单轮 |
| `internal/pkg/ai/provider/` | `ProviderID` 常量与 sentinel 错误 |
| `internal/pkg/ai/router/` | 路由辅助（Strategy + Chain of Responsibility — failover 顺序、分类、API-key 提示） |
| `internal/pkg/ai/agent/` | function-call `Tools()` 与 tool-call → AiT Agent JSON 适配 |
| `internal/pkg/ai/models/` | `ait ai model list` 行格式化 + role / key 标签 |
| `internal/pkg/ollama/` | Ollama 自动配置、`FetchTags`、`EnsureLocal` |
| `internal/pkg/transfer/` | SFTP get/put（单文件） |
| `internal/pkg/platform/` | `AppRoot` / `DataDir` / `ConfigDir`（`AIT_HOME`）、`ExpandUserAbsPath`、路径、`OpenSSHPath` |
| `internal/pkg/config/` | YAML 配置（`Load`/`Save`/`EnsureFile`/`Validate`；`ait config init/view/edit/validate`） |
| `internal/pkg/creds/` | 操作系统密钥环 |
| `internal/pkg/pathsafe/` | 会话 ID 与路径校验 |
| `pkg/` | 对外可导入库；[`pkg/third-party/`](pkg/third-party/) 内嵌第三方 Go 源码（如 xiwh ZMODEM） |
| `configs/` | `ait.example.yaml` |
| `docs/` | [`design.zh-CN.md`](docs/design.zh-CN.md)、[`user-manual.zh-CN.md`](docs/user-manual.zh-CN.md) |
| `.cursor/rules/` | Cursor 规则；Go 结构：[`go-structure.mdc`](.cursor/rules/go-structure.mdc) |
| `scripts/`、`build/` | 构建与 CI 辅助 |

优先在 `internal/pkg/*` 与 `internal/ait` 中扩展，避免随意新增顶层目录。

### `internal/ait` 布局

| 包 | 文件（示例） |
|----|----------------|
| `ait`（根） | `root.go` — `Run`、`ExitCode` |
| `ait/app` | `manager.go`、`context.go`、`run.go`（`Runner`） |
| `ait/cli` | `root.go`、`manager.go`、集成测试；各子命令位于 `cli/{ai,config,session,transfer,ollama,login}` |
| `ait/login` | `run.go`、`helpers.go`、`help.go` |
| `ait/shell` | `repl.go`、`exec.go`、`parse.go`、`io.go`、`*_helpers.go` |
| `ait/aimode` | `scope.go`、`mode.go`、`repl.go`、`exec.go`、`run_local.go`、`agent.go` — AI REPL（子包 `risk`、`approve`、`agent`、`chat`） |
| `ait/meta` | `exec.go` — `Execute`、`SetExecutorWriters` |

### `internal/pkg/ssh/client` 文件布局

`client` 包仅保留 `Client` 方法和对外 facade（`Dial`/`DialChain`/`RunInteractive`/`RunPTYCapture` 以及各子包符号的 alias 文件）；无 receiver 的辅助函数迁至上表的子包内。

| 文件 | 职责 |
|------|------|
| `client.go` | `Client`、`Options`、`RunInteractive`、meta 派发、本地模式钩子、`chainMode` 切换 |
| `client_dial.go`、`client_dial_config.go` | `Dial`、`DetectAuthMethod`、`buildAuth`、`buildSSHClientConfig` |
| `dialchain.go`、`dialchain_helpers.go` | `DialChain`、ProxyJump 调用方、`RunRemoteCommand` |
| `hostkeys.go` | known_hosts 回调（TTY 确认 / accept-new） |
| `client_interactive_helpers.go` | meta 派发辅助、刷新远程提示符、AI / transfer 入口、`handleChainMeta` |
| `client_init_helpers.go`、`client_setup_filter_helpers.go` | PTY 初始化输出过滤（抑制 PS1 泄漏） |
| `client_transfer.go`、`client_transfer_helpers.go` | `Client.handleTransfer*`（根据 chainMode 走 SFTP 或 PTY）、transfer hint 包装 |
| `inject.go`、`inject_helpers.go` | `Client.Inject` / `Client.InjectCapture` —— 静默 PTY 命令注入，含历史/回显抑制与捕获定界符 |
| `transfer_route.go`、`internal/pkg/lrzsz/`、`internal/pkg/transfer/route/` | 会话内 get/put 统一路由（lrzsz 探测、SFTP 回退、链式安装） |
| `transfer_zmodem.go`、`zmodem_session.go`、`internal/pkg/transfer/zmodem/` | `ZModemTransfer` — lrzsz 兼容 ZMODEM PTY 文件传输 |
| `scrollback_capture.go` | `Client.RunPTYCapture` 驱动 `scrollback.Delta`（使用 `Inject` 进行历史抑制） |
| `resize.go` | `sessionResize` 适配 `pty.Watch` |
| `ai_mode.go` | 远程 AI REPL 入口的 `RemoteAIParams` / `RemoteAIRunner` 类型 |
| `prompt_alias.go`、`meta_alias.go`、`pty_alias.go` | facade alias，将子包符号以旧名暴露给原调用方 |

包级辅助函数置于 `*_helpers.go`（不与 struct 方法混于同一文件）。

## 约定

- **人类贡献者**：[CONTRIBUTING.zh-CN.md](CONTRIBUTING.zh-CN.md) — 开发流程与 Go 编码规范。
- **Go 编码规范**：生成或重构 Go 代码时遵循 [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc)。
- **复用**：遵循 [`.cursor/rules/code-reuse.mdc`](.cursor/rules/code-reuse.mdc)。
- **CLI 帮助**：`ait <cmd> --help` 使用 [`internal/ait/helpdoc`](internal/ait/helpdoc)（`helpdoc.Bind` / `helpdoc.BindCustomArgv`）；见 [`.cursor/rules/cli-help.mdc`](.cursor/rules/cli-help.mdc)。
- **格式化**：Go 使用 tab；完成修改后运行 `make fmt`。
- **测试**：`*_test.go` 与实现同目录；`make test` 必须通过；store 测试使用 `:memory:` SQLite。
- **`internal/`**：模块私有，禁止被本模块外代码 import。
- **依赖**：按需添加；见 `go.mod`。
- **密钥**：API Key 与 SSH 密码存于操作系统密钥环（`creds`）；禁止提交或记录于日志；禁止将密码写入 SQLite。

## CLI 速查

```text
ait shell                              # AiT> REPL：login、session --list/--load、history、help、exit
ait login [ssh-options] destination [cmd ...] [--save] [--alias] [--tmux] [--password]
ait ai [--model deepseek|minimax]            # 本机 AI REPL（AiT AI>）；tiered 可先走 Ollama
ait ai model list
ait ai chat [message...]
ait ai provider enable|disable <provider>
ait ai ollama status|setup|start|stop|disable  # 本地 Ollama（setup 安装前 [y/N]）
ait session --list [-v]
ait session --load [<id|alias>] [--tmux] [--password]
ait session --switch [id|alias] [--tmux] [--password] [--force]
ait session --delete <id|alias|all> [--force]
ait get [ssh-flags] <paths...>   # ParseTransfer + Resolve + DialChain；flag/路径顺序任意
ait put [ssh-flags] <paths...>
ait config init|view|edit|validate|set|get <key> [<value>]
```

提示符：`AiT>`（`ait shell`）、`AiT AI>`（`ait ai`）、`AiT AI user@host>`（SSH AI）、`AiT user@host>`（SSH）、`AiT Local>`（`!ait local`）。SSH 内：`!ait help`；`!ait <subcmd>`（不含 `ai`）；`!ait ai`/`!!` 远程 AI；`!ait local`/`remote`；`!ait suspend`；`!ait chain [on|off]` 用于多跳 SSH 的 PTY 管道注入。合并门禁见 [change-acceptance.mdc](.cursor/rules/change-acceptance.mdc)（**R1–R18**、**D\***、**S\***、**B1–B7**、**H0a–H4d**）。

## 注意事项

- 保持 `cmd/ait` 精简，逻辑位于 `internal/ait` 与 `internal/pkg`。
- `login` 经 `ssh -G` 解析配置；数据面为 Go `x/crypto/ssh`。仅开发环境使用 `AIT_SSH_INSECURE=1`。
- **默认模式**不保证远程进程保活；仅 `--tmux` 提供远程会话保活。
- 从本机执行 `ait session --switch` 时，若仍处于 SSH 会话中，须先 `!ait suspend`，或使用 `--force`（不保存当前 scrollback/cwd）。
- `third_party/` 用于 vendored 工具，非主应用代码。
- `cli` 子包通过各自的 `New(...)` 工厂在 [`cli/root.go`](internal/ait/cli/root.go) 注册；新增命令请在子包中实现并经由 `root.AddCommand(...)` 接入。
- 除用户要求或更新产品文档外，不随意新增 Markdown 文件。
