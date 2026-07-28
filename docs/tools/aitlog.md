# aitlog

View and search system logs from the command line on Linux, macOS, and Windows.

## Build

```bash
make build   # → bin/aitlog
./bin/aitlog --help
```

## Usage

```text
aitlog [flags] [keyword...]
aitlog search [flags] <keyword...>
aitlog sources [--json] [--backend auto|native|builtin]
aitlog version
```

Default route views or searches logs. Keywords are case-insensitive **AND** matches against the message. Use `aitlog search …` when a keyword would collide with a subcommand name (for example `version`).

On macOS, keyword/regex queries use `log show --predicate` and expand at most through the last hour (stderr notes the window). Pass `--since` to search a wider explicit range.

### Common flags

| Flag | Meaning |
|------|---------|
| `-n, --lines` | Last N **matching** records (default 1000) |
| `--head` | First N matches instead of last N |
| `-e, --regex` | Regular expression match |
| `--case-sensitive` | Case-sensitive keywords/regex |
| `--since` / `--until` | Time window (`15m`, `2h`, `today`, RFC3339) |
| `-p, --priority` | Minimum priority (`emerg`…`debug`) |
| `-u, --unit` | Filter by unit/service |
| `-S, --source` | Logical source name or file path/glob |
| `-f, --follow` | Follow new entries (disables pager) |
| `--json` | NDJSON to stdout |
| `--raw` | Include original line in JSON |
| `--utc` | Timestamps in UTC |
| `--no-pager` | Disable interactive pager |
| `--backend` | `auto` (default), `native`, or `builtin` |
| `--quiet` | Suppress notices on stderr |
| `-V, --version` | Print version |

On a TTY, long output opens a less-like pager (`space`/`b`/`j`/`k`/`g`/`G`/`/`/`q`). Non-TTY and `--no-pager` write plain lines. No matches exits 0 with empty stdout (and a stderr notice unless `--quiet`).

## Fallback matrix

Every filter/search/pager feature runs in Go over a normalized record stream. Missing native tools only reduce which stores can be opened.

| Platform | Source | Preferred | Builtin fallback |
|----------|--------|-----------|------------------|
| Linux | `system` | `journalctl -o json` | `/var/log/syslog`, `/var/log/messages`, `/dev/kmsg` |
| Linux | `kernel` | `journalctl -k` | `/dev/kmsg`, `/var/log/kern.log` |
| Linux | `auth` | `journalctl` | `/var/log/auth.log`, `/var/log/secure` |
| macOS | `system` | `log show --style ndjson` | `/var/log/system.log` |
| macOS | `install` | — | `/var/log/install.log` |
| Windows | `system` / `application` / `security` | Win32 `EvtQuery` | `wevtutil`, then Stage B `.evtx` (deferred) |
| Windows | `setup` | — | `C:\Windows\Logs\…` |
| Any | `-S path` | — | Builtin file reader (rotations + `.gz`) |

`aitlog sources` shows the chain and marks the selected layer with `*`.

## Examples

```bash
aitlog -n 50
aitlog error --since 1h
aitlog -e 'fail(ed|ure)' -n 100
aitlog --backend builtin -S /var/log/system.log -n 20
aitlog --json --no-pager -n 5
aitlog search version
aitlog sources --json
```
