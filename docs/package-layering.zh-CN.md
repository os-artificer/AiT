# 包分层规范（AiT CLI）

**语言：** [English](package-layering.md) · 简体中文（本文）

Cursor Agent 权威副本：[`.cursor/rules/package-layering.mdc`](../.cursor/rules/package-layering.mdc)（约定变更时请与本页同步）。

本文规定 AiT 模块内**代码应放在哪一层**，与 [go-structure](../.cursor/rules/go-structure.mdc)（包内文件布局）、[code-reuse](../.cursor/rules/code-reuse.mdc)（何时抽取共用逻辑）配合使用。

---

## 调用链

```text
cmd/ait/main.go          → 仅极简 main
internal/ait             → Run、ExitCode、应用装配
internal/ait/cli         → Cobra 根与子命令注册（薄层）
internal/ait/<cmd>       → 每个顶层子命令一包（设计闭环）
internal/pkg/*           → 可复用库（不承载某一子命令的产品语义）
pkg/*                    → 供外部模块导入；与 AiT 业务无关的通用能力
```

---

## 子命令包（设计闭环）

每个顶层 **`ait <子命令>`**（如 `login`、`get`、`put`、`shell`、`ai`、`session`、`config`）放在 **`internal/ait/<子命令>/`**，并在**同一包内**完成：

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
- 长期把同一子命令的 argv 解析留在 `internal/pkg/ssh/parse`、业务却在 `internal/ait/login` 分裂。OpenSSH 共用机制进 `internal/pkg`；login 专有 flag 留在 `internal/ait/login`。

---

## `internal/ait/cli`

- **只做根接线**：`newRootCmd`、`Execute` / `ExecuteWriters`、`runWithManager`、`meta` / `shell` 等 init。
- 通过 **`internal/ait/<cmd>`** 的 `login.New`、`get.New` 等注册子命令。
- 避免 `cli/<cmd>/` 只包一层 Cobra，而 parse/flags 在别处的结构。

---

## `internal/pkg/*`

当 **两个及以上** 子命令需要相同的、**与子命令无关** 的能力时使用，例如：

- OpenSSH 兼容 flags（`-p`、`-l`、`-i`、`-F`、`-J`、`-o` 及 noop/unsupported 表）
- 目标地址解析（`user@host`、`ssh://`、方括号主机等）
- `store`、`session`、`ssh/client`、SFTP `transfer`、`config`、`creds`、`ai` 路由等

**依赖规则：** `internal/pkg` **不得** import `internal/ait`，依赖只能自上而下。

若 `internal/pkg` 需识别子命令名（如 meta `!ait get`）：

- 由 `internal/ait` 在 `init` 中注册回调，或
- 在 `internal/pkg` 保留协议级常量，并用测试保证与 `internal/ait/<cmd>.Name` 一致。

---

## 根目录 `pkg/*`

- 供**外部模块**导入、与 AiT 产品语义无关的库。
- 不要把 `ait login` 等业务逻辑放在此处。
- 模块内共用代码优先 `internal/pkg`。

---

## `helpdoc`

- **`internal/pkg/helpdoc`**：共用排版（`Body`、OpenSSH 块、`Bind`、`WriteTransferHelp`）。
- 各命令 flags 与长 `--help` 在**命令包**（`internal/ait/<cmd>/help.go`；嵌套见 `<parent>/<sub>/help.go`）。
- 根 `ait --help` 用 `registry.RootHelpEntries()`；嵌套父命令用 **`subregistry.Entries()`**。

---

## 包名 = 命令名

- 目录/包名与 Cobra 命令一致；禁止 `*cmd` 包后缀。
- `ait ai` REPL：`internal/ait/ai/mode`；单轮：`ai/ask`、`ai/agent`。
- CLI `session`：`internal/ait/session`；`internal/pkg/session` 用 `pkgsession` 别名。

---

## 嵌套子命令

子命令在父命令包树下实现；父级用 `subregistry` + `registry/all/imports.go` 挂接，勿在父 `cmd.go` 手写 growing `AddCommand`。`ait session` 无 Cobra 子节点，在父包内按 flag 分文件。详见英文 [package-layering.md](package-layering.md) 表格。

---

## 抽取共用逻辑示例

| 包 | 用途 |
|----|------|
| `internal/pkg/ssh/openssh` | `login` / `get` / `put` 共用的连接类 argv flags |
| `internal/pkg/ssh/transferparse` | 仅当 get/put 会大量重复路径分类逻辑时再抽 |

login 专有（`--alias`、`--save`、`--tmux`）与 get/put 专有（`--session`）分别放在对应 `internal/ait/<cmd>/flags.go`。

---

## 扩展顶层子命令（开闭原则）

使用 **`internal/ait/registry`**：各顶层命令实现 `registry.TopLevel`，在 `register.go` 的 `init()` 中调用 `registry.Register`。

- **不要**在 `cli/root.go` 的 `AddCommand` 手写命令清单；根 `--help` 由 `registry.RootHelpEntries()` / [`registry/root_help.go`](../internal/ait/registry/root_help.go) 聚合。
- **要**新增 `internal/ait/<cmd>/` 包（含 `register.go` 与业务文件）。
- **链接**：在 `registry/imports.go` 增加 blank import，或启用 `go generate` 自动生成（见执行计划）。

新增命令不得要求修改其他子命令包。

架构重构（registry、parse 搬迁）须保持存量行为：全量 `make test`，**R1–R19**、**A20–A22** 不回退（见 [change-acceptance.mdc](../.cursor/rules/change-acceptance.mdc)）；未改动路径的错误文案、退出码、测试断言不得削弱。

## 检查清单（新增或重构子命令）

- [ ] `internal/ait/<cmd>/` 含 Name、flags、Parse、Cobra `New`、Help、Run、**`register.go`**
- [ ] 已在 `init()` 中 `registry.Register`
- [ ] 未修改其他子命令包或 `cli/root.go` 命令列表
- [ ] 不在 `internal/pkg` 下新增「仅名称」包
- [ ] `internal/pkg` 不 import `internal/ait`
- [ ] `cmd/ait` 保持极简 main
- [ ] 跨子命令复用放在 `internal/pkg`，不在多个 `ait` 包内复制粘贴
