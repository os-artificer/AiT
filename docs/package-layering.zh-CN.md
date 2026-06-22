# 包分层规范（AiT 工具箱 — ai-ssh）

**语言：** [English](package-layering.md) · 简体中文（本文）

Cursor Agent 权威副本：[`.cursor/rules/package-layering.mdc`](../.cursor/rules/package-layering.mdc)（约定变更时请与本页同步）。

本文规定 AiT 模块内**代码应放在哪一层**。**AiT** 为工具箱；**`ai-ssh`** 为首个工具（二进制 `bin/ai-ssh`，入口 `cmd/ai-ssh`）。与 [go-structure](../.cursor/rules/go-structure.mdc)（包内文件布局）、[code-reuse](../.cursor/rules/code-reuse.mdc)（何时抽取共用逻辑）配合使用。

---

## 调用链

```text
cmd/ai-ssh/main.go          → 仅极简 main；委托 aish.Run
internal/ai-ssh             → 包 aish：Run、ExitCode、Execute、NewManager、Cobra 根
internal/ai-ssh/<cmd>       → 每个顶层子命令一包（设计闭环）
internal/pkg/*              → 可复用库（不承载某一子命令的产品语义）
pkg/*                       → 供外部模块导入；与 AiT 业务无关的通用能力
```

**目录与 Go 包名：** 应用根在 **`internal/ai-ssh/`**，声明 **`package aish`**（`ai-ssh` 不是合法 Go 标识符）。子命令目录仍使用与 Cobra 命令一致的包名（`login`、`shell`、`ai` 等）。

---

## 子命令包（设计闭环）

每个顶层 **`ai-ssh <子命令>`**（如 `login`、`get`、`put`、`shell`、`ai`、`session`、`config`）放在 **`internal/ai-ssh/<子命令>/`**，并在**同一包内**完成：

| 职责 | 典型文件 | 说明 |
|------|----------|------|
| 子命令名 | `cmd.go` 或 `name.go` | 如 `const Name = "login"`，供 Cobra `Use`、解析错误、meta 对齐共用 |
| 支持的 flags | `flags.go` | flag 令牌及本命令的 `Apply*` |
| 参数解析 | `parse.go` | 仅本命令的 `Parse(argv)` |
| Cobra | `cmd.go` | `New(runWithMgr) *cobra.Command`，放在此包而非 `cli/<cmd>/` |
| 帮助 | `help.go` | `--help` 正文；flag 行使用本包常量 |
| 执行 | `run.go` | `Run(ctx, mgr, argv)` 等 |

**禁止**

- 新建仅存放子命令名称的包（如 `internal/pkg/aitcmd`）。
- 长期把同一子命令的 argv 解析留在 `internal/pkg/ssh/parse`、业务却在 `internal/ai-ssh/login` 分裂。OpenSSH 共用机制进 `internal/pkg`；login 专有 flag 留在 `internal/ai-ssh/login`。

---

## `internal/ai-ssh`（根包 `aish`）

- **入口与根接线**：`Run`、`ExitCode`、`newRootCmd`、`Execute` / `ExecuteWriters`、`NewManager`、`runWithManager`；`init_hooks.go` 注入 `meta` / `shell`。
- **Go 包名：** `aish`（目录 `internal/ai-ssh`）。
- 通过 **`registry.AttachRoot`** 挂接子命令；各 **`internal/ai-ssh/<cmd>`** 在 `register.go` 实现 `registry.TopLevel`。
- 避免仅包一层 Cobra、parse/flags 却在别处的薄包装结构。

---

## `internal/pkg/*`

当 **两个及以上** 子命令需要相同的、**与子命令无关** 的能力时使用，例如：

- OpenSSH 兼容 flags（`-p`、`-l`、`-i`、`-F`、`-J`、`-o` 及 noop/unsupported 表）
- 目标地址解析（`user@host`、`ssh://`、方括号主机等）
- `store`、`session`、`ssh/client`、SFTP `transfer`、`config`、`creds`、`ai` 路由等

**依赖规则：** `internal/pkg` **不得** import `internal/ai-ssh`，依赖只能自上而下。

若 `internal/pkg` 需识别子命令名（如 meta `!ai-ssh get`）：

- 由 `internal/ai-ssh` 在 `init` 中注册回调，或
- 在 `internal/pkg` 保留协议级常量，并用测试保证与 `internal/ai-ssh/<cmd>.Name` 一致。

---

## 根目录 `pkg/*`

- 供**外部模块**导入、与 AiT 产品语义无关的库。
- 不要把 `ai-ssh login` 等业务逻辑放在此处。
- 模块内共用代码优先 `internal/pkg`。

---

## `helpdoc`

- **`internal/pkg/helpdoc`**：共用排版（`Body`、OpenSSH 块、`Bind`、`WriteTransferHelp`）。
- 各命令 flags 与长 `--help` 在**命令包**（`internal/ai-ssh/<cmd>/help.go`；嵌套见 `<parent>/<sub>/help.go`）。
- 根 `ai-ssh --help` 用 `registry.RootHelpEntries()`；嵌套父命令用 **`subregistry.Entries()`**。

---

## 包名 = 命令名

- **例外 — 工具箱根包：** 目录 `internal/ai-ssh/`，Go 包 **`aish`**（对应 `ai-ssh` 二进制名；带连字符的名称不能作为 Go 包名）。
- 其余情况目录/包名与 Cobra 命令一致；禁止 `*cmd` 包后缀。
- `ai-ssh ai` REPL：`internal/ai-ssh/ai/mode`；单轮：`ai/ask`、`ai/agent`。
- CLI `session`：`internal/ai-ssh/session`；`internal/pkg/session` 用 `pkgsession` 别名。
- **`ai-ssh shell` REPL 内置命令**（`exit`、`help`、`ll`、`login` 等）在 **`internal/ai-ssh/shell/builtin/`**（package `builtin`）；父包 `shell` 负责 REPL I/O 与非 builtin 四层分发（`exec_dispatch.go`）。

---

## 嵌套子命令

子命令在父命令包树下实现；父级用 `subregistry` + `registry/all/imports.go` 挂接，勿在父 `cmd.go` 手写 growing `AddCommand`。`ai-ssh session` 无 Cobra 子节点，在父包内按 flag 分文件。详见英文 [package-layering.md](package-layering.md) 表格。

---

## 抽取共用逻辑示例

| 包 | 用途 |
|----|------|
| `internal/pkg/ssh/openssh` | `login` / `get` / `put` 共用的连接类 argv flags |
| `internal/pkg/ssh/transferparse` | 仅当 get/put 会大量重复路径分类逻辑时再抽 |

login 专有（`--alias`、`--save`、`--tmux`）与 get/put 专有（`--session`）分别放在对应 `internal/ai-ssh/<cmd>/flags.go`。

---

## 命令入口（三入口模型）

AiT 有三条**独立**的命令注册路径。新增用户可见动词须落在正确层级，三者不可混用。

| 入口 | 机制 | 扩展方式 |
|------|------|----------|
| **Cobra 子命令** | `registry` / `subregistry` + 各包 `register.go` | 在 `internal/ai-ssh/<cmd>/`（或嵌套）新建包 + 在 `registry/all/imports.go` 增加 blank import |
| **`ai-ssh shell` 内置** | [`shell/builtin/registry.go`](../internal/ai-ssh/shell/builtin/registry.go) map | 在 `builtin/constants.go` 增名 + 在 registry map 注册处理函数 |
| **SSH meta / AI REPL** | [`meta/exec.go`](../internal/ai-ssh/meta/exec.go)、[`ai/mode/exec.go`](../internal/ai-ssh/ai/mode/exec.go) | Meta：委托 Cobra 或阻断；AI REPL：在 `mode` 增加 meta 动词 |

**特例 — `ai-ssh ai model list`：** 作为**父命令 positional** 进入 AI REPL meta 路径（R15），**不是** `ai/model` 下的 Cobra 子树。见 [design.zh-CN.md](design.zh-CN.md) AI CLI 小节。

Shell 内置（`list`、`load`、`login` 等）**不是** `ai-ssh shell` 的 Cobra 子命令；SSH 内 `!ai-ssh <subcmd>` 经 [`meta.Execute`](../internal/ai-ssh/meta/exec.go) 复用顶层 Cobra（`ai` / `!!` 除外，进入 AI 模式）。

**Try 块 SSOT：** errhint 中 `ai-ssh session …` / `ai-ssh login …` 示例行集中在 [`internal/pkg/clihints/try_lines.go`](../internal/pkg/clihints/try_lines.go)；`internal/pkg/session` 再导出以便引用。

---

## 扩展顶层子命令（开闭原则）

使用 **`internal/ai-ssh/registry`**：各顶层命令实现 `registry.TopLevel`，在 `register.go` 的 `init()` 中调用 `registry.Register`。

- **不要**在 `cli/root.go` 的 `AddCommand` 手写命令清单；根 `--help` 由 `registry.RootHelpEntries()` / [`registry/root_help.go`](../internal/ai-ssh/registry/root_help.go) 聚合。
- **要**新增 `internal/ai-ssh/<cmd>/` 包（含 `register.go` 与业务文件）。
- **链接**：在 [`registry/all/imports.go`](../internal/ai-ssh/registry/all/imports.go) **手动**增加 blank import，或运行 `go generate ./internal/ai-ssh/registry/...` 重新生成顶层 import（嵌套 subregistry 的 imports 仍手动维护）。

新增命令不得要求修改其他子命令包。

架构重构（registry、parse 搬迁）须保持存量行为：全量 `make test`，**R1–R19**、**A20–A22** 不回退（见 [change-acceptance.mdc](../.cursor/rules/change-acceptance.mdc)）；未改动路径的错误文案、退出码、测试断言不得削弱。

## 检查清单（新增或重构子命令）

- [ ] `internal/ai-ssh/<cmd>/` 含 Name、flags、Parse、Cobra `New`、Help、Run、**`register.go`**
- [ ] 已在 `init()` 中 `registry.Register`
- [ ] 未修改其他子命令包或 `cli/root.go` 命令列表
- [ ] 不在 `internal/pkg` 下新增「仅名称」包
- [ ] `internal/pkg` 不 import `internal/ai-ssh`
- [ ] `cmd/ai-ssh` 保持极简 main
- [ ] 跨子命令复用放在 `internal/pkg`，不在多个 `ai-ssh` 包内复制粘贴
