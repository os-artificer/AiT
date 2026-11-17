# AiT — Agent 导读

**AiT** 是一套 Go **工具箱**（模块 `github.com/os-artificer/AiT`）。已发布二进制：**`ait`**（Agent 交互入口 + 工具箱 shell）、**`aitssh`**（SSH；插件 `ssh`；文档与包名 aissh；**无 AI**）、**`aitimg`** / **`aitlog`** / **`aithost`**，以及 **`aitkeep`**（插件 `keep`）。目录布局遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout)。

## 构建与测试

```bash
make build       # → bin/ait, bin/aitssh, bin/aitimg, bin/aitlog, bin/aithost, bin/aitkeep（+ bin/plugins/*）
make test        # go test ./...
make vet         # go vet ./...
make fmt         # go fmt ./...
make test-cross  # 五平台 go build ./...（linux/darwin/windows）
make test-e2e-smoke  # pytest @smoke，含仓内 testenv-mock（无需真实主机）
make test-e2e-mock   # GitHub CI 全量 e2e，对 testenv-mock（`AIT_TEST_USE_MOCK=1`）
./bin/ait           # Agent Loop（默认）或插件分发
./bin/ait shell     # 工具箱 REPL（install/update/upgrade）
./bin/aitssh       # 运行 SSH 工具
./bin/aitimg       # 运行图片元数据工具
./bin/aitlog       # 运行日志查看工具
./bin/aithost      # 运行主机监控
./bin/aitkeep      # 进程守护（插件命令：keep）
```

需要 **Go 1.25+**（见 `go.mod`）。Windows 交叉编译：`GOOS=windows GOARCH=amd64 go build -o bin/aitssh.exe ./cmd/ssh`。

## AiT 做什么

- **SSH 会话**：默认 OpenSSH 风格连接（`aitssh [flags] destination`）、`shell`、`session`（`--list`、`--load`、`--suspend`、`--switch`、`--delete`）
- **持久化**：SQLite（`internal/pkg/store`）；挂起时保存 scrollback/cwd；认证 TTL 由 `session.auth_ttl` 控制（默认 7d；密钥环存机密）；默认 `AppRoot` 为 `~/.local/share/ait`（Linux）、`~/Library/Application Support/ait`（macOS）、`%LOCALAPPDATA%\ait`（Windows）；可用 `AIT_HOME` 覆盖
- **可选 tmux**：在 connect/`load`/`switch` 上使用 `--tmux` 保持远端 shell（`internal/pkg/tmux`）；`--tmux` 隐含 `--save`，并存储 `login_destination` 以支持 ssh config Host 别名
- **Ops AI Agent（`ait` / `ait agent`）**：Agent Loop 内核为 `internal/pkg/agentcore`（`agentops`、`agentui`、`agentwire`）；管理 CLI 在 `internal/ait/agent`（`connect`、`model`、`provider`、`ollama`、`config`）。裸 **`ait`** 与 **`ait agent`** 启动交互式 Loop；另有 **`ait agent model list`**、**`ait agent connect user@host`** 等。多提供商路由经 `internal/pkg/ai`；`ai.*` 配置经 **`ait agent config`**（不是 `aitssh config`）。**`aitssh` 不含任何 AI**（无 `aitssh ai`，无会话内 AI meta）。
- **传输**：`aitssh get` / `aitssh put` — SFTP 单文件传输（`internal/pkg/transfer`）；`!ait get` / `!ait put` 复用当前 SSH（统一路由：优先 lrzsz/ZMODEM，直连跳回退 SFTP，链式缺 lrzsz 时可自动安装）
- **代理链支持**：`!ait chain [on|off]` — 在直连主机之外再 SSH 时，用 PTY 管道做 ZMODEM 传文件与 tmux（`internal/pkg/ssh/client/inject.go`、`transfer_zmodem.go`；`internal/pkg/tmux/remote_pty.go`）
- **OpenSSH 配置**：默认 connect 与 `load` 通过 `ssh -G` 解析 `~/.ssh/config`（argv 解析：`internal/pkg/ssh/openssh`；`ssh -G`：`internal/pkg/ssh/resolve`）；数据面为 Go `x/crypto/ssh`
- **图片元数据（`aitimg` / 插件 `img`）**：经 `internal/pkg/imgmeta` 提取 EXIF；经 `internal/pkg/geocode`（Nominatim）逆地理编码；指南见 [docs/guide/aiimg.md](docs/guide/aiimg.md)
- **系统日志（`aitlog` / 插件 `log`）**：经 `internal/pkg/logsrc` 查看/搜索（原生工具 + 内置兜底），经 `internal/pkg/pager` 做类似 less 的分页；指南见 [docs/guide/ailog.md](docs/guide/ailog.md)
- **主机监控（`aithost` / 插件 `host`）**：基于 Bubble Tea 的实时多面板仪表盘（页眉：USER/OS/CPU/MEM/KERNEL + 日期时间；窄终端换行；面板：居中标题、淡色百分比、Braille 曲线（DISK/NET 双速率）、`HH:MM:SS` 横轴；TCP/UDP/ARP/ICMP 协议行；PROC 默认采样全部进程并按窗口高度填满，含 PEER 监听/远端与 LOG（最新系统日志行）列；`/` 模糊搜索 + Enter 详情浮层（匹配日志与完整 peer 列表）；`x` 杀进程需 SIGTERM 确认 / 可选 SIGKILL）；指标经 `internal/pkg/hoststats`，面板在 `internal/pkg/hostdash`。支持 Linux、macOS、Windows（`CGO_ENABLED=0`）；协议/日志后端按操作系统区分（见 [docs/guide/aihost.md](docs/guide/aihost.md)）。
- **进程守护（`aitkeep` / `ait keep`）**：`GET /ping` 返回 `pong`；`serve` 按 `keepalive.yaml` 中的 unit 做 CPU、内存、磁盘空间与磁盘 IO 限制（见 [docs/guide/aitkeep.md](docs/guide/aitkeep.md)）。CLI 与守护进程共用 `internal/pkg/keepalive`。

项目结构：[`docs/architecture/overview.md`](docs/architecture/overview.md)。ait：[`docs/guide/ait.md`](docs/guide/ait.md)、[`docs/architecture/ait.md`](docs/architecture/ait.md)。aissh：[`docs/guide/aissh.md`](docs/guide/aissh.md)、[`docs/architecture/aissh.md`](docs/architecture/aissh.md)。aiimg：[`docs/guide/aiimg.md`](docs/guide/aiimg.md)、[`docs/architecture/aiimg.md`](docs/architecture/aiimg.md)。ailog：[`docs/guide/ailog.md`](docs/guide/ailog.md)、[`docs/architecture/ailog.md`](docs/architecture/ailog.md)。aihost：[`docs/guide/aihost.md`](docs/guide/aihost.md)、[`docs/architecture/aihost.md`](docs/architecture/aihost.md)。aitkeep：[`docs/guide/aitkeep.md`](docs/guide/aitkeep.md)、[`docs/architecture/aitkeep.md`](docs/architecture/aitkeep.md)。示例配置：[`configs/ait.example.yaml`](configs/ait.example.yaml)。按 [`docs-exec-plans.mdc`](.cursor/rules/docs-exec-plans.mdc) 将 Cursor 方案归档到 `docs/<kebab-case>.md`。

## 代码所在位置

| 路径 | 职责 |
|------|------|
| `cmd/ait/main.go` | 薄 `main` — 委托给 `ait.Run` |
| `cmd/ssh/main.go` | 薄 `main` — 委托给 `aissh.Run`（二进制 `aitssh`） |
| `cmd/img/main.go` | 薄 `main` — 委托给 `aiimg.Run`（二进制 `aitimg`） |
| `cmd/log/main.go` | 薄 `main` — 委托给 `ailog.Run`（二进制 `aitlog`） |
| `cmd/host/main.go` | 薄 `main` — 委托给 `aihost.Run`（二进制 `aithost`） |
| `cmd/keep/main.go` | 薄 `main` — 委托给 `keepalive.Run`（二进制 `aitkeep`） |
| `internal/ait/` | `ait` 入口（包 **`ait`**：`Run`、`Execute`、router）；**`ait/agent`** Loop + 管理 CLI |
| `internal/pkg/aitplugin/` | 插件 Manifest v1、Discover（`plugins/<plugin-command>/`，如 `plugins/ssh/aitssh`）、内嵌 SSOT manifest |
| `internal/pkg/aitrelease/` | 为 REPL `install` / `update` / `upgrade` 拉取/解压/应用 GitHub Release |
| `mock/testrelease/` | httptest GitHub Releases 夹具（`mock/aitrelease-mock`） |
| `mock/testenv-mock/` | 仅测试用的回环 SSH+HTTP mock 守护进程（JSON banner、SIGINT） |
| `mock/testssh/` | 回环 mock sshd + SFTP，用于自动化验收 |
| `mock/testhttp/` | 回环 mock OpenAI / Ollama `/api/tags` / Nominatim |
| `internal/pkg/pluginbin/` | 已发现插件的 `ExecPath` + 可选 Runner（TTY Restore） |
| `internal/img/` | `aiimg` CLI（包 **`aiimg`**：`Run`、`Execute`、help、version ldflags） |
| `internal/log/` | `ailog` CLI（包 **`ailog`**：`Run`、`Execute`、view/search/sources/version） |
| `internal/host/` | `aihost` CLI（包 **`aihost`**：`Run`、`Execute`、top/cpu/mem/disk/net/proto/proc/snapshot） |
| `internal/pkg/logsrc/` | 系统日志源、过滤器、原生/内置后端、链式解析 |
| `internal/pkg/hoststats/` | 主机 CPU/mem/disk/net/进程采样与 Section 格式化 |
| `internal/pkg/hostdash/` | aihost 仪表盘 Panel 原语（sparkline） |
| `internal/pkg/pager/` | 类似 less 的交互式分页 |
| `internal/pkg/buildinfo/` | 共享版本字符串格式化（`ait`/`aitssh`/`aitimg`/`aitlog`/`aithost`/`aitkeep`） |
| `internal/pkg/keepalive/` | `aitkeep` 守护进程与 CLI（包 **`keepalive`**：ping、supervisor、`serve`/`add`/`status`） |
| `internal/pkg/imgmeta/` | 图片 EXIF 提取 + 人类可读/JSON 格式化 |
| `internal/pkg/geocode/` | Nominatim 逆地理编码 HTTP 客户端 |
| `internal/ssh/` | CLI 入口（Go 包 **`aissh`**：`Run`、`ExitCode`、`Execute`、`NewManager`、Cobra root）；子包见下 |
| `internal/ssh/registry/` | 顶层命令注册表（OCP）；根 help 经 `registry/root_help.go` + `internal/pkg/helpdoc` |
| `internal/pkg/helpdoc/` | 共享 `--help` 版式（`Body`、`Bind`、OpenSSH 块、`WriteTransferHelp`、TTY 按键） |
| `internal/ssh/connect/` | SSH connect 解析/运行（默认顶层路由；非注册子命令） |
| `internal/ssh/get/` / `internal/ssh/put/` | `aissh get` / `aissh put` SFTP |
| `internal/ssh/shell/` | `aissh shell` REPL（`AiT>`）与顶层 `aissh shell` Cobra |
| `internal/ssh/session/` | `aissh session`（flags：`--list`、`--load`、…）；引用 `internal/pkg/session` 时用别名 `pkgsession` |
| `internal/pkg/agentcore/` | `ait agent` / `ait agent connect` 的 Agent Loop 内核（`risk`、`approve`、`runner`、`RunLoop`） |
| `internal/pkg/agentops/` | Ops 工具箱 `DispatchToolbox` + 首轮上下文快照 |
| `internal/pkg/agentui/` | Agent Loop 的 Plain / TUI 发射器 |
| `internal/ssh/config/` | `aissh config` 父命令 + `config/subregistry`；嵌套 `set`、`get`、`init`、`view`、`edit`、`validate`（仅管理 `session.*`；`set`/`get` 拒绝 `ai.*`；AI 键使用 `ait agent config`） |
| `internal/ssh/version/` | `aissh version` |
| `internal/ssh/meta/` | 会话内 `!ait <subcmd>` 委托 |
| `internal/pkg/ssh/openssh/` | OpenSSH argv 解析 |
| `internal/pkg/ssh/resolve/` | `ssh -G` 解析；`Resolved` → `client.Options` |
| `internal/pkg/session/` | Login、load、suspend、switch、delete、list（`manager.go`、`manager_login.go`、`manager_load.go`、`manager_suspend_switch.go`、`manager_list_pick.go`、`manager_helpers.go`） |
| `internal/pkg/store/` | SQLite schema、迁移、CRUD（`store.go`、`store_sessions.go`、`store_state.go`、`store_meta.go`、`store_scan.go`、`session_helpers.go`、`open_helpers.go`） |
| `internal/pkg/ssh/client/` | `Client`、`Options`、`Dial`、`DialChain`、`RunInteractive` 门面、静默 PTY 注入（文件布局见下） |
| `internal/pkg/ssh/client/prompt/` | AiT 显示提示符字符串 + 远端 PS1 设置片段 |
| `internal/pkg/ssh/client/scrollback/` | 远端 PTY scrollback `Delta` + 捕获时序常量 |
| `internal/pkg/ssh/client/localmode/` | `!ait local` PTY `Shell`、环境、bash/zsh 初始化脚本 |
| `internal/pkg/ssh/client/meta/` | `!ait` 字节级扫描器 + meta 行常量（`!!` 不是 AI 入口） |
| `internal/pkg/ssh/client/dial/` | ProxyCommand stdio 连接、跳板解析、OpenSSH 密钥签名器 |
| `internal/pkg/ssh/client/pty/` | TTY stdin 读取器、REPL 字节通道、跨平台 resize 监视 |
| `internal/pkg/ssh/client/transfer/` | `!ait get|put` meta argv 解析 + 用法提示 |
| `internal/pkg/tmux/` | 远端 tmux（基于 SFTP + 链式模式下的 PTY 注入） |
| `internal/pkg/terminal/` | 终端门面（`path.go`），类型别名指向子包 |
| `internal/pkg/terminal/scrollback/` | 环形缓冲 scrollback `Buffer`（写入/回放、文件 IO） |
| `internal/pkg/terminal/screen/` | ANSI 屏幕原语（清屏） |
| `internal/pkg/terminal/crlf/` | CRLF 规范化 `io.Writer` |
| `internal/pkg/terminal/lineedit/` | `ReadLine`（Ctrl+A/E、历史导航、Ctrl+R 搜索） |
| `internal/pkg/terminal/history/` | `LineHistory`（内存 + 文件同步）与 `history` 命令 |
| `internal/pkg/ai/` | `Client`、`Chat`、多提供商路由门面 |
| `internal/pkg/agentcore/risk/` | `Classify`、`Interactive`、`ClassifyAit` — 命令风险分级 |
| `internal/pkg/agentcore/approve/` | `Confirm`、`Policy`、`PatternKey`、`SessionTrust` — 审批提示与信任 |
| `internal/pkg/agentcore/runner/` | `Action`、`ParseAction`、`Run`、捕获 runner — Agent 步骤执行 |
| `internal/pkg/ai/provider/` | `ProviderID` 常量与哨兵错误 |
| `internal/pkg/ai/router/` | 路由辅助（Strategy + Chain of Responsibility — 故障转移顺序、分类、API key 提示） |
| `internal/pkg/ai/agent/` | Function-call `Tools()` + tool-call → AiT Agent JSON 适配器 |
| `internal/pkg/ai/models/` | `ait agent model list` 行格式化 + 角色 / 密钥标签 |
| `internal/pkg/ollama/` | Ollama 安装、`FetchTags`、`EnsureLocal`、`ManualSetupInstructions` |
| `internal/pkg/transfer/` | SFTP get/put（单文件） |
| `internal/pkg/platform/` | `AppRoot` / `DataDir` / `ConfigDir`（`AIT_HOME`）、`ExpandUserAbsPath`、路径、OpenSSH |
| `internal/pkg/config/` | YAML 配置（`Load`/`Save`/`EnsureFile`/`Validate`；`aissh config init/view/edit/validate`） |
| `internal/pkg/creds/` | OS 密钥环（API key、会话解锁） |
| `internal/pkg/pathsafe/` | 会话 ID 与路径校验（scrollback 侧车） |
| `pkg/` | 可对外导入的库；[`pkg/third-party/`](pkg/third-party/) 为上游 Go 源码 vendor（如 xiwh ZMODEM） |
| `mock/` | 仅测试验收 mock：[`mock/testssh`](mock/testssh)、[`mock/testhttp`](mock/testhttp)、[`mock/testrelease`](mock/testrelease)，守护进程在 `mock/testenv-mock` / `mock/aitrelease-mock` |
| `configs/` | `ait.example.yaml` |
| `docs/` | [`README.md`](docs/README.md)、[`guide/`](docs/guide/)、[`architecture/`](docs/architecture/)、[`changelog/`](docs/changelog/) |
| `.cursor/rules/` | Cursor 规则；Go 结构：[`go-structure.mdc`](.cursor/rules/go-structure.mdc) |
| `scripts/` | 构建与 CI 辅助脚本 |

优先扩展 `internal/pkg/*` 与 `internal/ssh`，避免新建顶层目录。

### `internal/ssh` 布局（根 Go 包：`aissh`）

| 包 | 文件（示例） |
|---------|------------------|
| `aissh`（根） | `root.go`（`Run`、`ExitCode`）；`cli_root.go`（`Execute`、`newRootCmd`）；`manager.go`（`NewManager`、`runWithManager`）；`context.go`、`runner.go`；`init_hooks.go`（meta/shell 注入；**无 AI**）；集成测试 |
| `registry` | `registry.go`、`top_level.go`、`deps.go`、`root_help.go`、`all/imports.go` |
| `get`、`put`、`shell`、`session`、`version` | 顶层命令：`register.go`、`cmd.go`、`help.go`、`run.go` / flags；SSH connect 走默认路由（`internal/ssh/connect`） |
| `config` | `cmd.go`、`help.go`、`subregistry/`、`shared/`；嵌套 `set`、`get`、`init`、`view`、`edit`、`validate` |
| `meta` | `exec.go` — `Execute`、`SetExecutorWriters`（由 `init_hooks.go` 接线；SSH 内 CRLF stdout） |

Ops AI 管理命令位于 **`internal/ait/agent/`**（不在 `internal/ssh`）。Agent Loop 内核：**`internal/pkg/agentcore`**；交互入口：**`ait`** / **`ait agent`**，经 **`internal/ait/router`** + **`internal/ait/loop`**。

### `internal/pkg/ssh/client` 文件布局

该包仅保留带 `Client` 方法的实现与用户可见门面（`Dial`、`DialChain`、`RunInteractive`、`RunPTYCapture`，以及各子包导出符号的别名文件）。无 receiver 的辅助函数放在上述子包中。

| 文件 | 职责 |
|-------|------|
| `client.go` | `Client`、`Options`、`RunInteractive`、meta 分发、local-mode 钩子、`chainMode` 开关 |
| `client_dial.go`、`client_dial_config.go` | `Dial`、`DetectAuthMethod`、`buildAuth`、`buildSSHClientConfig` |
| `dialchain.go`、`dialchain_helpers.go` | `DialChain`、proxy jump、`RunRemoteCommand` |
| `hostkeys.go` | known_hosts 回调（TTY 确认 / accept-new） |
| `client_interactive_helpers.go` | meta 分发辅助、刷新远端提示符、传输入口、`handleChainMeta` |
| `client_init_helpers.go`、`client_setup_filter_helpers.go` | PTY setup 输出过滤（抑制 PS1 泄漏） |
| `client_transfer.go`、`client_transfer_helpers.go` | `Client.handleTransfer*`（按 chainMode 走 SFTP 或 PTY）、传输提示包装 |
| `inject.go`、`inject_helpers.go` | `Client.Inject` / `Client.InjectCapture` — 静默 PTY 命令注入（抑制历史/回显与捕获定界符） |
| `transfer_route.go`、`internal/pkg/lrzsz/`、`internal/pkg/transfer/route/` | 统一会话内 get/put 路由（lrzsz 探测、SFTP 回退、链式安装） |
| `transfer_zmodem.go`、`zmodem_session.go`、`internal/pkg/transfer/zmodem/` | `ZModemTransfer` — 兼容 lrzsz 的 PTY 上 ZMODEM |
| `scrollback_capture.go` | `Client.RunPTYCapture` 驱动 `scrollback.Delta`（用 `Inject` 抑制历史） |
| `resize.go` | 接入 `pty.Watch` 的 `sessionResize` 适配器 |
| `prompt_alias.go`、`meta_alias.go`、`pty_alias.go` | 门面别名，按旧名再导出子包符号 |

包级辅助函数放在 `*_helpers.go`（不与同一文件中的 struct 方法混放）。

## 约定

- **人工贡献者**：[CONTRIBUTING.md](CONTRIBUTING.md) — 开发流程与 Go 编码标准。
- **Go 编码标准**：生成或重构 Go 时遵循 [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) — import 分组（标准库 / 模块 / 第三方）、每个文件至多一个带方法的 struct、包级函数放在独立 `*_helpers.go`、行宽 150（`.editorconfig`）、嵌套 ≤2、函数内职责块之间空一行（§11）、导出项完整 godoc。
- **Package 分层**：`cmd/ssh` → `internal/ssh`（包 `aissh`）/ `<subcommand>` → 共享代码在 `internal/pkg`（不得 import `internal/ssh`）；根 `pkg/` 放对外、非业务库。见 [`docs/architecture/aissh.md#package-layering`](docs/architecture/aissh.md#package-layering) 与 [`.cursor/rules/package-layering.mdc`](.cursor/rules/package-layering.mdc)。
- **复用**：遵循 [`.cursor/rules/code-reuse.mdc`](.cursor/rules/code-reuse.mdc) — 在复制逻辑前先搜已有辅助与共享包。
- **CLI help**：共享版式在 [`internal/pkg/helpdoc`](internal/pkg/helpdoc)（`pkghelpdoc.Bind`）；各命令正文在 `internal/ssh/<cmd>/help.go`（嵌套：`<parent>/<sub>/help.go`）；见 [`.cursor/rules/cli-help.mdc`](.cursor/rules/cli-help.mdc)。
- **格式化**：Go 用 tab；完成 Go 改动前运行 `make fmt`。
- **测试**：`*_test.go` 与实现同目录；`make test` 必须通过。Store 测试使用 `:memory:` SQLite。
- **`internal/`**：模块私有 — 不得从本模块外导入。
- **依赖**：按需添加；见 `go.mod`（cobra、modernc sqlite、x/crypto/ssh、x/term、keyring、uuid、sftp、cloudwego/eino、yaml）。
- **机密**：API key 与 SSH 密码放在 OS 密钥环（`creds`）；切勿提交或记录。不要把密码存进 SQLite。
- **提交**：仅在用户明确要求时提交；切勿改 git 配置或提交机密（`.env`）。

## CLI 速查

```text
ait                                      # Agent Loop；未配置 AI 时退出并提示 Try: ait agent config set
ait agent [--plain] [goal...]            # 交互式 Agent Loop（TTY 默认 TUI）
ait shell                                # ait> 工具箱 REPL（install/update/upgrade）
ait> install [plugin...]                 # GitHub Release → 安装目录（无参则装全部官方插件）
ait> update [plugin...]                  # 仅更新已安装的官方插件
ait> upgrade                             # 更新 ait + 已安装官方插件（不接受额外参数）
aitssh shell                             # AiT> REPL：ssh 风格连接、session --list/--load、history、help、exit
aitssh [ssh-options] destination [cmd ...] [--save] [--alias] [--tmux] [--password]
# ssh-options: -p -l -i -F -J -o ... ；仍支持旧式 --host/--user
# 各平台均需 PATH 上有 OpenSSH 客户端（ssh），用于 ssh -G
# aitssh 无 AI 子树（请用 ait agent）；插件分发：ait ssh …
ait agent model list
ait agent connect [ssh-options] destination  # 在 SSH 主机上跑远端 Agent Loop
ait agent provider enable|disable <provider>
ait agent ollama status|setup|start|stop|disable
aitssh session --list [-v]
aitssh session --load [<id|alias>] [--tmux] [--password]
aitssh session --switch [id|alias] [--tmux] [--password] [--force]
aitssh session --delete <id|alias|all> [--force]
aitssh get [ssh-flags] <paths...>   # ParseTransfer + Resolve + DialChain；flags/paths 任意顺序
aitssh put [ssh-flags] <paths...>
aitssh config init|view|edit|validate|set|get <key> [<value>]
```

提示符：`ait>`（`ait shell`）；Agent Loop 经裸 `ait` / `ait agent` / `ait agent connect`（plain 或 TUI；确认 `y`/`c`/`s`/`n`；只读自动执行）；`AiT>`（`aitssh shell`）。SSH 远端交互用原生 shell PS1；`AiT Local>`（`!ait local`）。在 SSH（`aitssh`）内：`!ait help`；`!ait <subcmd>` 走本地 CLI；**没有** `!ait ai` / `!!` AI 入口 — 请用 `ait agent connect` 或 `ait agent`。另有 `!ait local` / `remote`；`!ait suspend`；`!ait chain [on|off]` 用于多跳 SSH 上的 PTY 管道注入。嵌套 connect/`load` 用 `aitssh shell`（远端 `exit` 后回到 `AiT>`）。合并门禁：[change-acceptance.mdc](.cursor/rules/change-acceptance.mdc)（**OA\***、**AG\***、**AC-V4\***、**R1–R14**、**D\***、**S\***、**B1–B7**、**H0a–H4d**）。

## 注意点

- 保持 `cmd/ssh` 薄 — 全部逻辑在 `internal/ssh` 与 `internal/pkg`。
- 默认 SSH connect（以及 `get`/`put`）经 `ssh -G` 解析配置（argv：`internal/pkg/ssh/openssh`；解析：`internal/pkg/ssh/resolve`）；数据面为 Go `x/crypto/ssh`。仅开发环境可设 `AIT_SSH_INSECURE=1`。
- **默认模式**不会保持远端进程存活；只有 `--tmux` 会。
- **从本地 shell 执行 `session --switch`** 无法挂起仍在进行的 SSH；用户须先在会话内 `!ait suspend`，或使用 `--force`（不保存当前 scrollback/cwd）。
- `pkg/third-party/` 用于 vendored 工具，不是主应用代码。

- 顶层命令：新增 `internal/ssh/<cmd>/`，含 `register.go`（`registry.Register`），并在 [`registry/all/imports.go`](internal/ssh/registry/all/imports.go) 加 blank import。嵌套命令：父级 `subregistry` + `<parent>/registry/all/imports.go`（见 [`docs/architecture/aissh.md#package-layering`](docs/architecture/aissh.md#package-layering)）。
- 除非用户要求，否则不要新增 markdown 文件（文档化产品时更新 `README.md` / `AGENTS.md` / `docs/` 除外）。
