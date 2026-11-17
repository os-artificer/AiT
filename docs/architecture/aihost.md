# aihost 实现说明

## 包分层

```text
cmd/host/main.go
  → internal/host（包名 aihost：Run、Execute、WriteRootHelp）
       → registry（包装 internal/pkg/subregistry）← OCP 注册
       → top（默认空 argv 路由，且已注册）
       → cpu / mem / disk / net / proto / proc / snapshot / sources / version
       → shared（旗标门面；不注册）
  → internal/pkg/hoststats   （Collector、Section、进程排序 / 限制、ListConnections）
  → internal/pkg/hostdash    （Panel + sparkline；不依赖 Bubble Tea 的原语）
  → internal/pkg/buildinfo
```

导入方向：`cmd` → `internal/host` → `internal/pkg/*`。  
`hoststats` 与 `hostdash` 不得导入 `internal/host` 或 `cobra`。

### 注册（OCP）

与 [ailog](ailog.md) 相同：包装 `internal/pkg/subregistry`，没有 Deps。

新增顶层命令：

1. 增加 `internal/host/<cmd>/{register,cmd,help,run?}.go`，在 `init` 里调用 `registry.Register`
2. 在 [`registry/all/imports.go`](../../internal/host/registry/all/imports.go) 中 blank-import
3. **不要**手改 `AddCommand` 列表，也不要在 `cli_root.go` 里按名称分支
4. 命令包**不得**导入兄弟命令包（AH15）

### 路由

| argv | 路由 |
|------|------|
| `-h` / `--help` / `help` | `WriteRootHelp` |
| `-V` / `--version` | 打印 `version.Version()` |
| 空 | `top.Run`（实时仪表盘） |
| 已注册名称 | 经 `registry.AttachRoot` 走 Cobra |
| 其他 | errhint + Try |

### hoststats

- `Collector.Sample` 填充 `Host`（用户、主机名、OS / arch、内核、CPU 数）、系统 CPU / Mem / Disk / Net / Proto，以及排序后的 `Processes`。
- `FormatHostHeader` / `WrapHostHeader` 构建仪表盘与全量快照页眉（`USER` / `OS` / `CPU` / `MEM` / `KERNEL` + 日期时间；不含主机名；窄屏换行）。
- 默认进程排序：CPU% → Mem% → NetRate → DiskIORate；`--sort` 设置主键。
- `proc` / `snapshot` 的默认 `--limit` 为 10（上限 500）；`top` 默认 `ProcessQuery.All`（全部进程排序），使 `/` 搜索覆盖全表；可选 `--limit` 限制采样（最大 500）。列表模式仍只渲染窗口能容纳的行；搜索 / 详情传 `KeepPIDs`，使冻结行与详情焦点在 live top-N 排序之外仍刷新。
- PROC 表列：PID、NAME、CPU%、MEM%、DISK/s、PEER（监听为 `:port×N`，否则为远端 `host:port`）、LOG（最新匹配的系统日志行）。
- 进程日志来自 `internal/pkg/logsrc`（与 `ailog` 相同后端）；先按 PID 匹配，再按精确进程名；采集在后台刷新，以免拖慢 `Sample`。PID 分别来自 journalctl `_PID`（Linux）、`log show` 的 `processID`（macOS）、wevtutil 的 `Execution ProcessID`（Windows）。
- 协议采样因操作系统而异：Linux `/proc`；macOS `netstat -s`（加上 `netstat -an` 取当前 TCP ESTABLISHED）与 `arp -an`；Windows `iphlpapi`（`GetTcpStatisticsEx` / `GetUdpStatisticsEx` IPv4+IPv6、`GetIcmpStatistics`、`GetIpNetTable`）。ARP lookup 速率仅 Linux 有。
- `ProcStat` 含 `PPID`、截断后的 `Cmdline`、`LogLine`、用于匹配 / 详情的 `Logs`。
- 模糊匹配：`MatchProc` / `FilterProcs`；树：`ProcAncestors` / `ProcChildren`。
- `KillProc(pid, force)` 发送 SIGTERM 或 SIGKILL（Windows：Terminate / Kill）。
- `Section` 策略：CPU / Mem / Disk / Net / Proto / Proc 的 Human + JSON。
- `ListConnections` / `ConnQuery` 支撑 `aithost net --conn`（gopsutil 的 tcp / udp / inet；先排序再限制；`--listen` 仅为 TCP LISTEN；`--port` 匹配本地或远端，默认 proto 为 tcp）。Linux 的 AGE / CREATED 优先用 `/proc/<pid>/fd` 的 Ctim；失败时可用 nf_conntrack 的 `delta-time`（CREATED ≈ now−delta）。状态别名（例如 Windows 的 `SYN_RECEIVED`）归一到 Linux 规范名；`DELETE` 与 `CLOSE` 保持区分。不嵌入 `Snapshot` 或 `top`。
- `shared.WatchEvery` / `shared.WatchSeparator` 支撑分段 watch 与 `net --conn --watch`（人类可读样本用 `---`；JSON 无分隔符）。`snapshot` 也绑定了 `--watch` / `--interval`。

### hostdash

- `Panel` 接口：Title / Update / View。
- `ProcPanel`：主列表只处理 `/` 进入搜索（搜索时行顺序冻结，数值仍刷新）。搜索态可用方向键 / `j` / `k`、Enter 打开详情（资源、每条完整 peer 一行，窗口不够时 `… +N more`、祖先、子进程、匹配系统日志）、`x` 杀进程（先 SIGTERM 确认，失败后再问 SIGKILL）。搜索态的 `q` 与 Esc 退出搜索。`HandleKey` 由 `top.Update` 接入。
- `DefaultPanels()`：CPU、MEM、DISK、NET、TCP、UDP、ARP、ICMP、PROC。
- Sparkline 是带淡色 `HH:MM:SS` X 轴的 Braille 折线；Y 刻度、图例与同色数值共享系列色。CPU / MEM 用专用百分比色。DISK / NET / TCP / UDP / ARP / ICMP 用青色入、橙色出，左右独立 Y 轴（`r` / `w` 或 `↓` / `↑`，NET 另含 errs / drops）。

### 验收 ID

见 change-acceptance.mdc：**AH1–AH81**。
