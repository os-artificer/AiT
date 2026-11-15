# aihost internals

## Package layering

```text
cmd/host/main.go
  → internal/host (package aihost: Run, Execute, WriteRootHelp)
       → registry (wraps internal/pkg/subregistry)  ← OCP registration
       → top (default empty-argv route + registered)
       → cpu / mem / disk / net / proto / proc / snapshot / sources / version
       → shared (flags facade; not registered)
  → internal/pkg/hoststats   (Collector, Section, process sort/limit, ListConnections)
  → internal/pkg/hostdash    (Panel + sparkline; Bubble Tea-free primitives)
  → internal/pkg/buildinfo
```

Import direction: `cmd` → `internal/host` → `internal/pkg/*`.  
`hoststats` and `hostdash` must not import `internal/host` or `cobra`.

### Registry (OCP)

Same pattern as [ailog-internals](ailog-internals.md): wrap `pkg/subregistry`, no Deps.

New top-level command:

1. Add `internal/host/<cmd>/{register,cmd,help,run?}.go` with `registry.Register` in `init`
2. Blank-import from [`registry/all/imports.go`](../../internal/host/registry/all/imports.go)
3. Do **not** hand-edit `AddCommand` lists or add per-name branches in `cli_root.go`
4. Command packages must **not** import sibling command packages (AH15)

### Routing

| argv | Route |
|------|-------|
| `-h` / `--help` / `help` | `WriteRootHelp` |
| `-V` / `--version` | print `version.Version()` |
| empty | `top.Run` (live dashboard) |
| registered name | Cobra via `registry.AttachRoot` |
| else | errhint + Try |

### hoststats

- `Collector.Sample` fills `Host` (user, hostname, OS/arch, kernel, CPUs), system CPU/Mem/Disk/Net/Proto, and sorted `Processes`
- `FormatHostHeader` / `WrapHostHeader` build the dashboard and full-snapshot header (`USER` / `OS` / `CPU` / `MEM` / `KERNEL` + datetime; no hostname; wraps on narrow width)
- Default process sort: CPU% → Mem% → NetRate → DiskIORate; `--sort` sets primary key
- Default `--limit` 10 for `proc`/`snapshot` (cap 500); `top` defaults to `ProcessQuery.All` (every process, sorted) so `/` search covers the full table; optional `--limit` caps the sample (max 500); list mode still renders only window capacity rows; search/detail pass `KeepPIDs` so frozen rows (and the detail focus) keep refreshing outside the live top-N sort
- PROC table columns: PID, NAME, CPU%, MEM%, DISK/s, PEER (`:port×N` for listeners, else remote `host:port`), LOG (latest matching system-log line)
- Process logs come from `internal/pkg/logsrc` (same backends as `ailog`); matching is by PID, then exact process name; collection is refreshed in the background so `Sample` stays responsive. PID comes from journalctl `_PID` (Linux), `log show` `processID` (macOS), and wevtutil `Execution ProcessID` (Windows).
- Protocol sampling is OS-specific: Linux `/proc`, macOS `netstat -s` (plus `netstat -an` for current TCP ESTABLISHED) + `arp -an`, Windows `iphlpapi` (`GetTcpStatisticsEx`/`GetUdpStatisticsEx` IPv4+IPv6, `GetIcmpStatistics`, `GetIpNetTable`). ARP lookup rates exist on Linux only.
- `ProcStat` includes `PPID`, truncated `Cmdline`, `LogLine`, and `Logs` for match/detail
- Fuzzy match helpers: `MatchProc` / `FilterProcs`; tree helpers: `ProcAncestors` / `ProcChildren`
- `KillProc(pid, force)` sends SIGTERM or SIGKILL (Windows: Terminate / Kill)
- `Section` strategy: CPU/Mem/Disk/Net/Proto/Proc Human + JSON
- `ListConnections` / `ConnQuery` power `aithost net --conn` (gopsutil tcp/udp/inet; sort then limit; `--listen` is TCP LISTEN only; `--port` matches local/remote with default proto tcp). Linux AGE/CREATED prefer `/proc/<pid>/fd` Ctim; on failure may use nf_conntrack `delta-time` (CREATED ≈ now−delta). Status aliases (e.g. Windows `SYN_RECEIVED`) normalize to Linux canonical names; `DELETE` stays distinct from `CLOSE`. Not embedded in `Snapshot` or `top`.
- `shared.WatchEvery` / `shared.WatchSeparator` power section watch and `net --conn --watch` (human samples use `---`; JSON has no separator).

### hostdash

- `Panel` interface: Title / Update / View
- `ProcPanel` adds `/` search (frozen row order while searching; values still refresh), Enter detail overlay (resources, one full peer record per line with `… +N more` when the window is short, ancestors, children, matching system logs), `x` kill with SIGTERM confirm and optional SIGKILL on failure; Esc back; `HandleKey` is wired from `top.Update`
- `DefaultPanels()`: CPU, MEM, DISK, NET, TCP, UDP, ARP, ICMP, PROC
- Sparklines are Braille polylines with an `HH:MM:SS` X-axis (faint); Y ticks, legends, and matching values share the series color. CPU/MEM use a dedicated percent tint. DISK/NET/TCP/UDP/ARP/ICMP plot in (cyan) vs out (orange) with independent left/right Y axes (`r`/`w` or `↓`/`↑`, plus NET errs/drops in/out).

### Acceptance IDs

See change-acceptance.mdc: **AH1–AH81**.
