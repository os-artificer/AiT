# aihost

> **名称：** 二进制 `aithost`，插件命令 `host`（`ait host …`），Go 包名 `aihost`。

主机监视：CPU、内存、磁盘、网络、协议与进程。二进制 `aithost`，插件 `host`。

```bash
make build   # → bin/aithost
./bin/aithost --help
```

## 用法

```text
aithost                         # 实时仪表盘（需要 TTY）
aithost top [flags]             # 同一仪表盘
aithost snapshot [--json] [--watch]  # 一次性全量指标；可持续刷新
aithost cpu|mem|disk|net|proto  # 单个类别
aithost net --conn [filters]    # 连接表（不是接口流量）
aithost proc [--limit 10] [--sort cpu|mem|net|disk]
aithost sources [--json]
aithost version
```

无参数（或 `top`）时打开多面板 TUI：CPU、MEM、DISK、NET、TCP / UDP / ARP / ICMP，以及 PROC（排序为 CPU → MEM → NET → DISK）。PROC 默认采样**全部**进程，并按窗口高度填满面板（`--limit` 可限制采样数量）。

PROC 按键按状态分层：

- 主列表只处理 `/`，按名称、命令行或 PID 做模糊搜索。进入搜索后行顺序冻结，CPU / MEM 等数值仍会刷新。在主列表按 `j`、`k` 或 `x` 没有作用。
- 进入搜索之后，可用 ↑↓ / `j` / `k` 移动光标，Enter 打开详情浮层，`x` 进入杀进程确认。`q` 与 Esc 退出搜索，不退出整个界面；若搜索框里还有查询，会先清空查询。
- 详情浮层展示资源、命令行、每条 peer 一行、祖先链、子进程，以及匹配的系统日志；面板过短时多余 peer 收成 `… +N more`。详情层可按 `x`。详情层没有被面板处理的 `q` 会退出界面；主列表上的 `q` 也会退出界面。

杀进程时先确认后发 SIGTERM；失败后再询问是否强制 SIGKILL。PROC 表含 `PEER`（监听写成 `:port×conns`，否则写成远端 `host:port`）与 `LOG`（该进程最新一条系统日志）。页眉打印 `USER`、`OS`、`CPU`（核数）、`MEM`（总量）、`KERNEL` 与日期时间；窄屏时左对齐换行。指标面板用居中的 `TYPE   value` 标题行，以及带 `HH:MM:SS` X 轴的 Braille 曲线。Y 刻度与面板数值和曲线同色：CPU 与 MEM 为百分比；DISK 为 `r` / `w` 速率（标题仍显示使用率）；NET 与 TCP / UDP / ARP / ICMP 为 `↓` / `↑`，左右独立 Y 轴（NET 还会给 errs / drops 上色）。

无 TTY 时用 `snapshot` 或 `proc`，不要用 `top`。

## 常用旗标

部分旗标只适用于特定子命令：

* `top` 支持 `--interval`、`--limit`、`--sort`，但隐藏 `--json` / `--watch`。
* `sources` 只支持 `--json`。

| 旗标 | 含义 |
|------|------|
| `--json` | 机器可读 JSON（snapshot、cpu / mem / disk / net / proto / proc、sources；`top` 隐藏该旗标） |
| `--interval` | `--watch`（cpu / mem / disk / net / proto / proc / snapshot）与 `top` 的刷新间隔（默认 1s） |
| `--watch` | 持续刷新直到 Ctrl+C（cpu / mem / disk / net / proto / proc / snapshot） |
| `--limit` | 进程行数（`proc` / `snapshot` 默认 10，最大 500；`top` 默认全部进程，可选上限最大 500）。对 `net --conn` 限制连接行数（默认 100，最大 5000） |
| `--sort` | `top` / `proc` 的主排序键：`cpu`（默认）、`mem`、`net`、`disk` |
| `--conn` | 仅 `net`：列出连接，而不是接口流量 |
| `--proto` | 仅 `net --conn`：`tcp`（默认）、`udp`，或 `all`（inet 套接字，不含 unix）。与 `aithost proto` 无关 |
| `--state` | 仅 `net --conn`：逗号分隔的状态（与 `--listen` 互斥） |
| `--listen` | 仅 `net --conn`：TCP LISTEN 套接字（不含 UDP 绑定） |
| `--port` | 仅 `net --conn`：匹配本地或远端端口（1～65535）；默认 `--proto` 为 tcp（非 TCP 用 `udp` / `all`） |
| `--pid` | 仅 `net --conn`：按进程 id 过滤 |

## 示例

```bash
aithost snapshot
aithost snapshot --json
aithost snapshot --watch
aithost proc --limit 5 --sort mem
aithost proto --json
aithost net --conn
aithost net --conn --port 8080
aithost net --conn --listen --json
aithost cpu --json
aithost sources
aithost top --interval 2s --limit 10   # --limit 可选封顶；省略则采样全部进程
```

## 说明

- 单进程网络速率在部分平台上可能为 `0`。
- 磁盘 IO、网络与协议速率至少两次采样才有意义（`--watch` / `top`）。
- 协议计数（TCP / UDP / ARP / ICMP）在部分 OS 上可能不完整。
- `net --conn` 的 AGE / CREATED：Linux 优先 `/proc/<pid>/fd` 的 fd 寿命；Lstat 失败可回退 nf_conntrack `delta-time`（CREATED ≈ now−delta）。非 Linux、不可读 fd 或无匹配时，人类为 `-`，JSON 的 `age_seconds` / `created_at` 为 null。`--port` 默认 TCP（`--proto udp|all` 改范围）。
- `CGO_ENABLED=0`：Linux / macOS / Windows 共用 CLI 与仪表盘；采集因 OS 而异：

| 区域 | Linux | macOS | Windows |
|------|-------|-------|---------|
| CPU / mem / disk / net | gopsutil | gopsutil | gopsutil（磁盘挂载 `C:\`） |
| TCP / UDP / ICMP | `/proc` 协议计数 | `netstat -s`（单复数行）；TCP `estab` 来自 `netstat -an` | `iphlpapi` IPv4+IPv6 |
| ARP | `/proc/net/arp`（有数据时含 lookup 速率） | `arp -an`（仅表） | `GetIpNetTable`（仅 IPv4 表） |
| `net --conn` | gopsutil + fd AGE | gopsutil；AGE `-` | gopsutil；AGE `-` |
| PROC PEER | `/proc` TCP | gopsutil | iphlpapi |
| PROC LOG | `journalctl` `_PID` | `log show` `processID` | `wevtutil` `Execution ProcessID` |
| Kill | SIGTERM / SIGKILL | SIGTERM / SIGKILL | Terminate / Kill |

实现说明：[../architecture/aihost.md](../architecture/aihost.md)。
