# aithost

`aithost` is a cross-platform host monitor for CPU, memory, disk, network, and top processes.

```bash
make build   # → bin/aithost
./bin/aithost --help
```

## Usage

```text
aithost                         # live dashboard (TTY required)
aithost top [flags]             # same dashboard
aithost snapshot [--json]       # one-shot all metrics
aithost cpu|mem|disk|net        # one category
aithost proc [--limit 10] [--sort cpu|mem|net|disk]
aithost sources [--json]
aithost version
```

With no arguments (or `top`), opens a multi-panel TUI: CPU, MEM, DISK, NET, and PROC (default top 10, sorted CPU → MEM → NET → DISK). Quit with `q`.

Without a TTY, use `snapshot` or `proc` instead of `top`.

## Flags (common)

Some flags apply only to specific subcommands:

* `top` supports `--interval`, `--limit`, `--sort` but hides `--json` / `--watch`.
* `sources` supports only `--json`.

| Flag | Meaning |
|------|---------|
| `--json` | Machine-readable JSON (snapshot, cpu/mem/disk/net/proc, sources; `top` hides it) |
| `--interval` | Refresh interval for `--watch` (cpu/mem/disk/net/proc) and `top` (default 1s) |
| `--watch` | Keep refreshing until Ctrl+C (cpu/mem/disk/net/proc) |
| `--limit` | Process rows (for `top` and `proc`; default 10, max 500) |
| `--sort` | Primary process sort key for `top` / `proc`: `cpu` (default), `mem`, `net`, `disk` |

## Examples

```bash
aithost snapshot
aithost snapshot --json
aithost proc --limit 5 --sort mem
aithost cpu --json
aithost sources
aithost top --interval 2s --limit 10
```

## Notes

- Per-process **network** rates are best-effort and may be `0` on some platforms.
- Disk IO and network rates need at least two samples for meaningful values (`--watch` / live `top`).
- Cross-platform: Linux, macOS, Windows (`CGO_ENABLED=0`).

Internals: [aithost-internals.md](aithost-internals.md).
