# AiT — Agent 导向

**语言：** [English](AGENTS.md) · 简体中文（本文）

**AiT**（AI Terminal）为 Go CLI：具备 SQLite 会话持久化、可选远程 tmux 保活、Deepseek 助手与 SFTP 文件 get/put 的 AI 增强 SSH 客户端。模块：`github.com/os-artificer/AiT`。目录布局遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout)。

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
- **传输**：`ait get` / `ait put` — SFTP 单文件（`internal/pkg/transfer`）；`!ait get` / `!ait put` 复用当前 SSH
- **OpenSSH 配置**：`login`/`load` 经 `ssh -G` 解析 `~/.ssh/config`（`internal/pkg/ssh/args`）；数据面为 Go `x/crypto/ssh`

完整设计：[`docs/design.zh-CN.md`](docs/design.zh-CN.md)。配置示例：[`configs/ait.example.yaml`](configs/ait.example.yaml)。Cursor 方案归档至 `docs/<kebab-case>.md`（见 [`.cursor/rules/docs-exec-plans.mdc`](.cursor/rules/docs-exec-plans.mdc)）。

## 代码位置

| 路径 | 职责 |
|------|------|
| `cmd/ait/main.go` | 极简 `main` — 委托至 `internal/ait.Run` |
| `internal/ait/` | CLI 入口（`Run`、`ExitCode`）；子包见下表 |
| `internal/ait/cli/` | Cobra 命令树、`Execute` / `ExecuteWriters` |
| `internal/ait/app/` | `NewManager`、`CommandContext`（store/config/ai 引导） |
| `internal/ait/login/` | OpenSSH login 解析/解析、帮助、退出码 |
| `internal/ait/shell/` | `ait shell` REPL（`AiT>`）、TTY I/O、行执行 |
| `internal/ait/meta/` | 会话内 `!ait <subcmd>` 委托 |
| `internal/pkg/ssh/args/` | OpenSSH 参数解析、`ssh -G` 解析、`Resolved` → `client.Options` |
| `internal/pkg/session/` | login、load、suspend、switch、delete、list |
| `internal/pkg/store/` | SQLite  schema、迁移、CRUD |
| `internal/pkg/ssh/client/` | 拨号、PTY、`!ait` meta、本地模式、主机密钥 |
| `internal/pkg/tmux/` | 远程 tmux 确保/附着/分离 |
| `internal/pkg/terminal/` | scrollback 缓冲区、回放、路径解析 |
| `internal/pkg/ai/` | 多 provider 路由、tiered、模型列表 |
| `internal/pkg/ollama/` | Ollama 自动配置、`FetchTags`、`EnsureLocal` |
| `internal/pkg/transfer/` | SFTP get/put（单文件） |
| `internal/pkg/platform/` | 数据/配置目录、`OpenSSHPath`、`LookPath` |
| `internal/pkg/config/` | YAML 配置 |
| `internal/pkg/creds/` | 操作系统密钥环 |
| `internal/pkg/pathsafe/` | 会话 ID 与路径校验 |
| `pkg/` | 对外可导入库（当前为空） |
| `configs/` | `ait.example.yaml` |
| `docs/` | [`design.zh-CN.md`](docs/design.zh-CN.md)、[`user-manual.zh-CN.md`](docs/user-manual.zh-CN.md) |
| `.cursor/rules/` | Cursor 规则；Go 结构：[`go-structure.mdc`](.cursor/rules/go-structure.mdc) |
| `scripts/`、`build/` | 构建与 CI 辅助 |

优先在 `internal/pkg/*` 与 `internal/ait` 中扩展，避免随意新增顶层目录。

### `internal/ait` 布局

| 包 | 文件（示例） |
|----|----------------|
| `ait`（根） | `root.go` — `Run`、`ExitCode` |
| `ait/app` | `manager.go`、`context.go` |
| `ait/cli` | `root.go`、`login.go`、`shell.go`、`load.go`、`session.go` 等 |
| `ait/login` | `run.go`、`helpers.go`、`help.go` |
| `ait/shell` | `repl.go`、`exec.go`、`parse.go`、`io.go`、`*_helpers.go` |
| `ait/aimode` | `scope.go`、`policy.go`、`repl.go`、`exec.go`、`run_local.go`、`prompt.go` — AI REPL（Runner 由调用方注入） |
| `ait/meta` | `exec.go` — `Execute`、`SetExecutorWriters` |

### `internal/pkg/ssh/client` 文件布局

| 文件 | 职责 |
|------|------|
| `client.go` | `Client`、`Options`、`RunInteractive`、meta 处理 |
| `client_dial.go` | `Dial`、`DetectAuthMethod`、`buildAuth`、`loadSigner` |
| `client_interactive_helpers.go` | stdin 读取、终端输出、密码捕获 |
| `dialchain.go`、`dialchain_helpers.go` | `DialChain`、代理跳转、`RunRemoteCommand` |
| `hostkeys.go` | known_hosts 回调 |
| `metascanner.go`、`metascanner_helpers.go` | `!ait` / `!!` 行检测 |
| `localmode.go`、`localmode_init.go`、`localmode_init_helpers.go` | `!ait local` PTY |
| `prompt.go` | 远程/本地提示符字符串 |
| `resize.go`、`resize_unix.go`、`resize_windows.go` | 窗口尺寸调整 |
| `proxy_conn.go`、`proxy_unix.go`、`proxy_windows.go` | ProxyCommand stdio 连接 |

包级辅助函数置于 `*_helpers.go`（不与 struct 方法混于同一文件）。

## 约定

- **人类贡献者**：[CONTRIBUTING.zh-CN.md](CONTRIBUTING.zh-CN.md) — 开发流程与 Go 编码规范。
- **Go 编码规范**：生成或重构 Go 代码时遵循 [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc)。
- **复用**：遵循 [`.cursor/rules/code-reuse.mdc`](.cursor/rules/code-reuse.mdc)。
- **CLI 帮助**：`ait <cmd> --help` 使用 [`internal/ait/helpdoc`](internal/ait/helpdoc) 与 [`cli/help_bind.go`](internal/ait/cli/help_bind.go)；见 [`.cursor/rules/cli-help.mdc`](.cursor/rules/cli-help.mdc)。
- **格式化**：Go 使用 tab；完成修改后运行 `make fmt`。
- **测试**：`*_test.go` 与实现同目录；`make test` 必须通过；store 测试使用 `:memory:` SQLite。
- **`internal/`**：模块私有，禁止被本模块外代码 import。
- **依赖**：按需添加；见 `go.mod`。
- **密钥**：API Key 与 SSH 密码存于操作系统密钥环（`creds`）；禁止提交或记录于日志；禁止将密码写入 SQLite。

## CLI 速查

```text
ait shell
ait login [ssh-options] destination [cmd ...] [--save] [--alias] [--tmux] [--password]
ait ai [--model deepseek|minimax]            # 本机 AI REPL（AiT AI>）；tiered 可先走 Ollama
ait ai model list
ait ai chat [message...]
ait ai provider enable|disable <provider>
ait ai ollama status|setup|start|stop|disable  # 本地 Ollama（setup 安装前 [y/N]）
ait session --list
ait session --load <id|alias> [--tmux] [--password]
ait session --switch [id|alias] [--tmux] [--password] [--force]
ait session --delete <id|alias|all> [--force]
ait get [ssh-flags] <paths...>   # ParseTransfer + Resolve + DialChain；flag/路径顺序任意
ait put [ssh-flags] <paths...>
ait config set|get <key> [<value>]
```

提示符：`AiT>`（`ait shell`）、`AiT AI>`（`ait ai`）、`AiT AI user@host>`（SSH AI）、`AiT user@host>`（SSH）、`AiT Local>`（`!ait local`）。SSH 内：`!ait help`；`!ait <subcmd>`（不含 `ai`）；`!ait ai`/`!!` 远程 AI；`!ait local`/`remote`；`!ait suspend`。合并门禁见 [change-acceptance.mdc](.cursor/rules/change-acceptance.mdc)。

## 注意事项

- 保持 `cmd/ait` 精简，逻辑位于 `internal/ait` 与 `internal/pkg`。
- `login` 经 `ssh -G` 解析配置；数据面为 Go `x/crypto/ssh`。仅开发环境使用 `AIT_SSH_INSECURE=1`。
- **默认模式**不保证远程进程保活；仅 `--tmux` 提供远程会话保活。
- 从本机执行 `ait session --switch` 时，若仍处于 SSH 会话中，须先 `!ait suspend`，或使用 `--force`（不保存当前 scrollback/cwd）。
- `third_party/` 用于 vendored 工具，非主应用代码。
- 除用户要求或更新产品文档外，不随意新增 Markdown 文件。
