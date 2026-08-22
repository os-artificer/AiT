# ailog

View and search system logs from the command line on Linux, macOS, and Windows.

## Build

```bash
make build   # → bin/ailog
./bin/ailog --help
```

## Usage

```text
ailog [flags] [keyword...]
ailog search [flags] <keyword...>
ailog sources [--json] [--backend auto|native|builtin]
ailog version
```

Default route views or searches logs. Keywords are case-insensitive **AND** matches against the message. Use `ailog search …` when a keyword would collide with a subcommand name (for example `version`).

Use `--expr` for boolean matching (`&`/`|`/`!` or `AND`/`OR`/`NOT`, parentheses, `"phrases"`). Shell-quote the value so `|` and `!` are not interpreted by the shell. `--expr` cannot be combined with positional keywords or `-e/--regex`.

Watch-style streaming: `ailog -f --expr '…'` follows new entries and prints only matches (silent until a hit). With `-f`, `-n`/`--head` are ignored; stop with Ctrl+C.

On macOS, keyword/regex/`--expr` queries use `log show --predicate` and expand at most through the last hour (stderr notes the window). Pass `--since` to search a wider explicit range.

### Common flags

| Flag | Meaning |
|------|---------|
| `-n, --lines` | Last N **matching** records (default 1000; ignored with `-f`) |
| `--head` | First N matches instead of last N (ignored with `-f`) |
| `-e, --regex` | Regular expression match |
| `--expr` | Boolean keyword/phrase expression (`&\|!` or `AND/OR/NOT`) |
| `--case-sensitive` | Case-sensitive keywords/regex/expr |
| `--since` / `--until` | Time window (`15m`, `2h`, `today`, RFC3339) |
| `-p, --priority` | Minimum priority (`emerg`…`debug`) |
| `-u, --unit` | Filter by unit/service |
| `-S, --source` | Logical source name or file path/glob |
| `-f, --follow` | Follow new entries (disables pager; watch until Ctrl+C) |
| `--json` | NDJSON to stdout |
| `--raw` | Include original line in JSON |
| `--utc` | Timestamps in UTC |
| `--no-pager` | Disable interactive pager |
| `--backend` | `auto` (default), `native`, or `builtin` |
| `--quiet` | Suppress notices on stderr |
| `-V, --version` | Print version |

On a TTY, long output opens a less-like pager (`space`/`b`/`j`/`k`/`g`/`G`/`/`/`q`). Non-TTY and `--no-pager` write plain lines (one line per matching record; embedded newlines in the message are flattened to spaces). No matches exits 0 with empty stdout (and a stderr notice unless `--quiet`). Follow mode stays quiet until a match.

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

`ailog sources` shows the chain and marks the selected layer with `*`. Windows Event Log does not support `-f`; use `-f --expr -S file` or historical `--expr` without follow.

## Examples

```bash
ailog -n 50
ailog error --since 1h
ailog -e 'fail(ed|ure)' -n 100
ailog -f --expr '(error|fail) & !debug'
ailog --expr '(error OR fail) AND NOT timeout' -n 50 --since 1h
ailog search --expr 'error|fail' --since 1h
ailog --backend builtin -S /var/log/system.log -n 20
ailog --json --no-pager -n 5
ailog search version
ailog sources --json
```
