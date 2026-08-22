# Changelog

This file records **product and CLI** functional changes only; documentation-only updates are not included. When a release is tagged, move items from `[Unreleased]` into a dated version section.

## [Unreleased]

- 【Added】`ait` TTY REPL: Tab completes the first command word against discovered plugins and reserved builtins (LCP then candidate list).
- 【Changed】**BREAKING** Renamed shipped CLI binaries and `ait` REPL plugin names from `aitssh`/`aitimg`/`aitlog`/`aithost` to `aissh`/`aiimg`/`ailog`/`aihost` (same names for binary and REPL). Package dirs are `internal/{ssh,img,log,host}`; `go install …/cmd/aissh@latest` etc. Default `aiimg` sibling outputs use `*_aiimg.jpg`. Re-run `make install` so `plugins/` matches the new binary names.

## [v1.1.0-beta.01]

- 【Added】`aihost`: live host monitor for CPU, memory, disk, network, protocols, and processes (`top` TUI, `snapshot`, `cpu`/`mem`/`disk`/`net`/`proto`/`proc`, `sources`, `version`).
- 【Added】`aihost top`: USER/OS/CPU/MEM/KERNEL header; Braille spark curves with series-matched colors (CPU/MEM percents; DISK `r`/`w` and NET/proto `↓`/`↑` dual Y-axes); TCP/UDP/ARP/ICMP panels; PROC with PEER column, `/` fuzzy search, detail overlay, and `x` kill (SIGTERM confirm; optional SIGKILL on failure).
- 【Added】`aihost top` / `proc`: PROC `LOG` column shows the latest system-log line for each process; the detail overlay lists matching records from the same `ailog` backends.
- 【Added】`aihost top` PROC detail: peer records are listed one per line in full; extra rows collapse to `… +N more` when the panel is short.
- 【Fixed】`aihost` protocol/log collectors: Windows Event XML records `Execution ProcessID` (and process image basename) for PROC LOG; Windows TCP/UDP stats include IPv6; macOS parses singular `netstat -s` lines and takes TCP `estab` from `netstat -an`; Linux/macOS/Windows proto backends remain best-effort (ARP lookup rates Linux-only).
- 【Added】`aiimg`: extract image EXIF metadata (device, time, GPS) with Nominatim reverse-geocoded Location; `--json` / `--no-geocode`.
- 【Added】`aiimg --all`: print full structured EXIF (IFD0/ExifIFD/GPS/…) after the locked summary; `--json --all` embeds an `exif` object.
- 【Added】`aiimg set` / `aiimg clear`: write or clear locked EXIF fields on JPEG; `aiimg version` subcommand; OCP command registry.
- 【Changed】**BREAKING** `aiimg set`/`clear`: default write target is a new sibling file (`*_aiimg.jpg`); TTY prompts to overwrite; use `--inplace`/`-y` or `-o` for explicit paths.



## [v1.0.1]

- 【Added】AI is experimental and off by default; enable via `ai.experimental.enabled`, `AIT_AI_EXPERIMENTAL=1`, or `--ai-experimental`.
- 【Fixed】Default remote prompt is AiT-managed again; set `AIT_REMOTE_SYSTEM_PROMPT=1` for the native shell prompt.
- 【Fixed】After nested `ssh`, active host and chain mode update once the shell is quiet (including idle MOTD).
- 【Fixed】`-J` / ProxyJump: jump clients stay open; Host aliases resolve via `ssh -G`.
- 【Fixed】Chain-mode `!ait put` no longer hangs at `finishing…` (including zero-byte uploads).



## [v1.0.0]

- 【Optimized】Restructure the project codebase into independent modules by package name; rename the SSH client to `aissh` to lay the groundwork for extending the toolbox with more tools.
- 【Optimized】Refactor the AI module on the Eino framework; improve fallback logic when interacting with LLMs so non-standard response formats are handled instead of breaking Agent mode.



## [v1.0.0-beta.2]

- 【Added】Shell mode supports full command-line interaction; command lookup order: builtins first, then extension commands and the PATH environment variable.
- 【Optimized】Session connection uniqueness key extended to username+host+port+alias; multiple `--alias` values on the same target host save separate sessions; login without flags updates only the empty-alias matching row.



## v1.0.0-beta.1

- 【Added】AI-enhanced SSH client: activate AI in an SSH session via `!! 'question'`, `!ait ai ask`, `!ait ai agent`, or `!ait ai`.
- 【Added】Built-in file transfer: download and upload files in an SSH session via `!ait get` / `!ait put`.
- 【Added】Switch between local and remote modes at any time within an SSH session.

