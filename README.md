# AiT

**AiT** is a cross-platform (Linux / macOS / Windows) AI-enhanced command-line **toolbox** for operations engineers—terminal and SSH workflows today, more tools over time.

**Tools:**

- [`ait`](cmd/ait) — toolbox interactive REPL (discovers plugins under `bin/plugins/`). Guide: [ait](docs/tools/ait.md).
- [`aissh`](cmd/aissh) — AI-enhanced SSH client (sessions, in-terminal AI, file transfer). Guide: [aissh](docs/tools/aissh.md).
- [`aiimg`](cmd/aiimg) — image metadata extractor (EXIF, GPS, reverse-geocoded address). Guide: [aiimg](docs/tools/aiimg.md).
- [`ailog`](cmd/ailog) — system log viewer and search (Linux/macOS/Windows). Guide: [ailog](docs/tools/ailog.md).
- [`aihost`](cmd/aihost) — live host CPU/mem/disk/net/process monitor. Guide: [aihost](docs/tools/aihost.md).

## What you get

- OpenSSH-compatible connect (`-J`, config Host aliases, `ssh -G` resolution)
- Session bookmarks: `--alias`, `session --load`, `session --switch`
- In-session `!ait` meta commands (suspend, chain hop, transfer)
- Default AiT-managed remote prompt (`AiT user@host:…>`); set `AIT_REMOTE_SYSTEM_PROMPT=1` for the native shell prompt
- Experimental AI (`!!` / `!ait ai` / `aissh ai`) — **off by default**; enable with `ai.experimental.enabled`, `AIT_AI_EXPERIMENTAL=1`, or `--ai-experimental`
- `!ait get` / `!ait put` on the live SSH session, standalone `aissh get` / `put` via SFTP
- Optional `--tmux` to keep remote shells alive after disconnect
- Local `aissh shell` REPL (and optional AI REPL) with history and UTF-8 line editing

## Examples

```bash
# 1) Bookmark a host and resume later (OpenSSH-style flags still work)
aissh deploy@prod --alias prod
aissh session --list
aissh session --load prod

# 2) Jump host in one command (same as ssh -J; Host aliases resolve via ssh -G)
aissh -J bastion app.internal

# 3) Enable experimental AI once, then use it locally or inside SSH
aissh config init
aissh config set ai.experimental.enabled true
# Inside an SSH session:
!!
# Ask in natural language; read-only commands run automatically; writes need confirmation

# 4) Transfer over the current session (single-arg defaults to ./basename)
!ait get ./build/app-linux-amd64
!ait put ./patch.tar.gz

# 5) tmux keep-alive: remote work survives disconnect
aissh devhost --tmux
# Later: aissh session --load devhost --tmux

# 6) One-shot local AI (non-interactive; requires experimental AI enabled)
aissh ai ask "summarize this log pattern: connection reset"
```

## Installation

**Prerequisites:** Go **1.25+**; **OpenSSH client** (`ssh`) on PATH for default connect/`load` to resolve `~/.ssh/config`.

**Build from source** (local build and verification):

```bash
git clone https://github.com/os-artificer/AiT.git
cd AiT
make build
./bin/ait --help
./bin/aissh --help
./bin/aiimg --help
./bin/ailog --help
./bin/aihost --help
```

**Install all tools** via `go install` (requires Go **1.25+**; ZMODEM support is vendored under `pkg/third-party/` — no external `github.com/xiwh/zmodem` fetch):

```bash
go install github.com/os-artificer/AiT/cmd/ait@latest github.com/os-artificer/AiT/cmd/aissh@latest github.com/os-artificer/AiT/cmd/aiimg@latest github.com/os-artificer/AiT/cmd/ailog@latest github.com/os-artificer/AiT/cmd/aihost@latest
ait --help
aissh --help
aiimg --help
ailog --help
aihost --help
```

Ensure `$(go env GOPATH)/bin` is on your PATH. When adding a new `cmd/<tool>`, append `github.com/os-artificer/AiT/cmd/<tool>@latest` to the `go install` line above.

## Documentation

| Document | Path |
|----------|------|
| Project structure | [docs/architecture.md](docs/architecture.md) |
| aissh internals | [docs/tools/aissh-internals.md](docs/tools/aissh-internals.md) |
| aissh guide | [docs/tools/aissh.md](docs/tools/aissh.md) |
| aiimg internals | [docs/tools/aiimg-internals.md](docs/tools/aiimg-internals.md) |
| aiimg guide | [docs/tools/aiimg.md](docs/tools/aiimg.md) |
| ailog internals | [docs/tools/ailog-internals.md](docs/tools/ailog-internals.md) |
| ailog guide | [docs/tools/ailog.md](docs/tools/ailog.md) |
| aihost internals | [docs/tools/aihost-internals.md](docs/tools/aihost-internals.md) |
| aihost guide | [docs/tools/aihost.md](docs/tools/aihost.md) |
| Python E2E tests | [tests/README.md](tests/README.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Code layout | [AGENTS.md](AGENTS.md) |
| Code of conduct | [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
