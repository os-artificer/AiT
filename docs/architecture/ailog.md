# ailog 实现说明

## 包分层

```text
cmd/log/main.go
  → internal/log（包名 ailog：Run、Execute、WriteRootHelp）
       → registry（包装 internal/pkg/subregistry）← OCP 注册
       → view（默认路由，不注册）
       → search / sources / version（已注册）
       → queryflags（共享旗标绑定与 Query 构建）
  → internal/pkg/logsrc   （Cursor / Source / Resolver，原生 + 内置后端）
  → internal/pkg/pager    （类似 less 的查看器）
  → internal/pkg/buildinfo （共享版本格式化）
```

导入方向：`cmd` → `internal/log` → `internal/pkg/*`。`logsrc` 与 `pager` 不得导入 `internal/log` 或 `cobra`。

### 与 aiimg 的注册差异

`ailog` 包装 [`internal/pkg/subregistry`](../../internal/pkg/subregistry/subregistry.go)（不像 `aiimg` 手写 registry），无注入 `Deps`。新增顶层命令：加 `internal/log/<cmd>/register.go`，并在 [`registry/all/imports.go`](../../internal/log/registry/all/imports.go) blank-import。

### 路由

| argv | 路由 |
|------|------|
| 空 / `-h` / `--help` / `help` | `WriteRootHelp` |
| 已注册名称（`search`、`sources`、`version`） | 经 `registry.AttachRoot` 走 Cobra |
| 其他（含顶层 `-V` / `--version` 旗标） | `view.Run`（默认路由；`-V` 在 view 内打印版本，避免与 `search version` 关键字冲突） |

## logsrc

- **Cursor** `{ Next, Err, Close }` — 不是 channel。
- **Filter** — 关键字 AND、正则、`--expr`（AND / OR / NOT）、级别、时间；所有后端共用一套 Go 实现。
- **Limit** — 最近 N 条*匹配*（不是扫描 N 条）。支持反向的后端在拿到 N 条最新匹配后停止。`-f` 忽略 `-n` / `--head`。
- **Resolver** — 可注入 `LookPath` / `Notice` / `Now` 供测试（`--backend builtin` → 零 LookPath）。
- **Chains** — `chains_linux.go` / `chains_darwin.go` / `chains_windows.go` / 桩 `chains_other.go`。
- **macOS `log show`**：`WindowedSource` / `WindowedSourceFor` 无过滤 last-N 按 `5m→1h→6h→24h→7d` 扩窗；关键字 / 正则 / `--expr` 进 `--predicate`，只扩到 `1h`（更宽用 `--since`）。`journalctl` 仅对关键字 / 正则加 `-g`。`chains_darwin.go` 对 `system` 先试 ndjson `log show`，再试 syslog 风格，最后 `/var/log/system.log`。
- **`--expr`** — `ParseExpr` / `Expr.Eval` 在 `expr.go` + `expr_helpers.go`；与关键字 / `-e` 的互斥在 `queryflags`。
- Stage B 二进制读取器（`.journal` / `.asl` / `.evtx`）列为 `Available: false, Reason: "not implemented"`。

## pager

交互式 raw-TTY 查看器，含 progressive 与 bounded 模式。非 TTY / `--no-pager` 直通。raw 写入经 `terminal/crlf`。

## 验收 ID

见 change-acceptance.mdc：**LG1–LG22**、**PG1–PG3**、**FB1–FB4**、**BI1–BI4**。
