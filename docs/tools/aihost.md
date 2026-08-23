# aihost

`aihost` is a cross-platform host monitor for CPU, memory, disk, network, protocols, and top processes.

```bash
make build   # → bin/aihost
./bin/aihost --help
```

## Usage

```text
aihost                         # live dashboard (TTY required)
aihost top [flags]             # same dashboard
aihost snapshot [--json]       # one-shot all metrics
aihost cpu|mem|disk|net|proto  # one category
aihost proc [--limit 10] [--sort cpu|mem|net|disk]
aihost sources [--json]
aihost version
```

With no arguments (or `top`), opens a multi-panel TUI: CPU, MEM, DISK, NET, TCP/UDP/ARP/ICMP, and PROC (sorted CPU → MEM → NET → DISK). PROC samples **all** processes by default and fills the panel from the window height (`--limit` optionally caps the sample). Press `/` to fuzzy-search by name, cmdline, or PID (row order freezes while searching; CPU/MEM/etc. still refresh); ↑↓/`j`/`k` to move; Enter for a PROC-area detail overlay (resources, cmdline, one peer record per line, ancestor chain, children, and matching system-log records; extra peers collapse to `… +N more` when the panel is short); `x` to kill the selected process (SIGTERM after `y` confirm; on failure, ask whether to force SIGKILL); Esc backs out; `q` quits. The PROC table shows `PEER` (listening servers as `:port×conns`, otherwise remote `host:port`) and `LOG` (latest system-log line for that process). The header prints `USER`, `OS`, `CPU` (count), `MEM` (total), `KERNEL`, and datetime; it wraps left-aligned on narrow terminals. Metric panels use a centered `TYPE   value` title line and Braille trend curves with an `HH:MM:SS` X-axis. Y ticks and on-panel values use the same color as their curve: CPU and MEM percents; DISK `r`/`w` rates (title still shows usage %); NET and TCP/UDP/ARP/ICMP `↓`/`↑` in/out with left/right Y axes (NET also tints errs/drops in/out).

Without a TTY, use `snapshot` or `proc` instead of `top`.

## Flags (common)

Some flags apply only to specific subcommands:

* `top` supports `--interval`, `--limit`, `--sort` but hides `--json` / `--watch`.
* `sources` supports only `--json`.

| Flag | Meaning |
|------|---------|
| `--json` | Machine-readable JSON (snapshot, cpu/mem/disk/net/proto/proc, sources; `top` hides it) |
| `--interval` | Refresh interval for `--watch` (cpu/mem/disk/net/proto/proc) and `top` (default 1s) |
| `--watch` | Keep refreshing until Ctrl+C (cpu/mem/disk/net/proto/proc) |
| `--limit` | Process rows (`proc`/`snapshot` default 10, max 500; `top` default: all processes; optional cap, max 500) |
| `--sort` | Primary process sort key for `top` / `proc`: `cpu` (default), `mem`, `net`, `disk` |

## Examples

```bash
aihost snapshot
aihost snapshot --json
aihost proc --limit 5 --sort mem
aihost proto --json
aihost cpu --json
aihost sources
aihost top --interval 2s --limit 10
```

## Notes

- Per-process **network** rates are best-effort and may be `0` on some platforms.
- Disk IO, network, and protocol rates need at least two samples for meaningful values (`--watch` / live `top`).
- Protocol counters (TCP/UDP/ARP/ICMP) are best-effort and may be partial on some OSes.
- Cross-platform (`CGO_ENABLED=0`): Linux, macOS, and Windows share the same CLI and dashboard. Collectors differ by OS:

| Area | Linux | macOS | Windows |
|------|-------|-------|---------|
| CPU / mem / disk / net | gopsutil | gopsutil | gopsutil (disk mount `C:\`) |
| TCP/UDP/ICMP | `/proc` proto counters | `netstat -s` (singular/plural); TCP `estab` from `netstat -an` | `iphlpapi` IPv4+IPv6 |
| ARP | `/proc/net/arp` (+ lookup rates when present) | `arp -an` (table only) | `GetIpNetTable` (IPv4 table only) |
| PROC PEER | `/proc` TCP | sysctl / gopsutil | iphlpapi |
| PROC LOG | `journalctl` `_PID` | `log show` `processID` | `wevtutil` `Execution ProcessID` |
| Kill | SIGTERM / SIGKILL | SIGTERM / SIGKILL | Terminate / Kill |

Internals: [aihost-internals.md](aihost-internals.md).
