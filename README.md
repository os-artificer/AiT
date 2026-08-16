# AiT

**AiT** is a cross-platform (Linux / macOS / Windows) AI-enhanced command-line **toolbox** for operations engineers—terminal and SSH workflows today, more tools over time.

**Tools:**

- [`aitssh`](cmd/aitssh) — AI-enhanced SSH client (sessions, in-terminal AI, file transfer). Guide: [aitssh](docs/tools/aitssh.md).
- [`aitimg`](cmd/aitimg) — image metadata extractor (EXIF, GPS, reverse-geocoded address). Guide: [aitimg](docs/tools/aitimg.md).
- [`aitlog`](cmd/aitlog) — system log viewer and search (Linux/macOS/Windows). Guide: [aitlog](docs/tools/aitlog.md).
- [`aithost`](cmd/aithost) — live host CPU/mem/disk/net/process monitor. Guide: [aithost](docs/tools/aithost.md).

## What you get

- OpenSSH-compatible connect (`-J`, config Host aliases, `ssh -G` resolution)
- Session bookmarks: `--alias`, `session --load`, `session --switch`
- In-session `!ait` meta commands (suspend, chain hop, transfer)
- Default AiT-managed remote prompt (`AiT user@host:…>`); set `AIT_REMOTE_SYSTEM_PROMPT=1` for the native shell prompt
- Experimental AI (`!!` / `!ait ai` / `aitssh ai`) — **off by default**; enable with `ai.experimental.enabled`, `AIT_AI_EXPERIMENTAL=1`, or `--ai-experimental`
- `!ait get` / `!ait put` on the live SSH session, standalone `aitssh get` / `put` via SFTP
- Optional `--tmux` to keep remote shells alive after disconnect
- Local `aitssh shell` REPL (and optional AI REPL) with history and UTF-8 line editing

## Examples

```bash
# 1) Bookmark a host and resume later (OpenSSH-style flags still work)
aitssh deploy@prod --alias prod
aitssh session --list
aitssh session --load prod

# 2) Jump host in one command (same as ssh -J; Host aliases resolve via ssh -G)
aitssh -J bastion app.internal

# 3) Enable experimental AI once, then use it locally or inside SSH
aitssh config init
aitssh config set ai.experimental.enabled true
# Inside an SSH session:
!!
# Ask in natural language; read-only commands run automatically; writes need confirmation

# 4) Transfer over the current session (single-arg defaults to ./basename)
!ait get ./build/app-linux-amd64
!ait put ./patch.tar.gz

# 5) tmux keep-alive: remote work survives disconnect
aitssh devhost --tmux
# Later: aitssh session --load devhost --tmux

# 6) One-shot local AI (non-interactive; requires experimental AI enabled)
aitssh ai ask "summarize this log pattern: connection reset"
```

## Installation

**Prerequisites:** Go **1.25+**; **OpenSSH client** (`ssh`) on PATH for default connect/`load` to resolve `~/.ssh/config`.

**Build from source** (local build and verification):

```bash
git clone https://github.com/os-artificer/AiT.git
cd AiT
make build
./bin/aitssh --help
./bin/aitimg --help
./bin/aitlog --help
./bin/aithost --help
```

**Install all tools** via `go install` (requires Go **1.25+**; ZMODEM support is vendored under `pkg/third-party/` — no external `github.com/xiwh/zmodem` fetch):

```bash
go install github.com/os-artificer/AiT/cmd/aitssh@latest github.com/os-artificer/AiT/cmd/aitimg@latest github.com/os-artificer/AiT/cmd/aitlog@latest github.com/os-artificer/AiT/cmd/aithost@latest
aitssh --help
aitimg --help
aitlog --help
aithost --help
```

Ensure `$(go env GOPATH)/bin` is on your PATH. When adding a new `cmd/<tool>`, append `github.com/os-artificer/AiT/cmd/<tool>@latest` to the `go install` line above.

## Documentation

| Document | Path |
|----------|------|
| Project structure | [docs/architecture.md](docs/architecture.md) |
| aitssh internals | [docs/tools/aitssh-internals.md](docs/tools/aitssh-internals.md) |
| aitssh guide | [docs/tools/aitssh.md](docs/tools/aitssh.md) |
| aitimg internals | [docs/tools/aitimg-internals.md](docs/tools/aitimg-internals.md) |
| aitimg guide | [docs/tools/aitimg.md](docs/tools/aitimg.md) |
| aitlog internals | [docs/tools/aitlog-internals.md](docs/tools/aitlog-internals.md) |
| aitlog guide | [docs/tools/aitlog.md](docs/tools/aitlog.md) |
| aithost internals | [docs/tools/aithost-internals.md](docs/tools/aithost-internals.md) |
| aithost guide | [docs/tools/aithost.md](docs/tools/aithost.md) |
| Python E2E tests | [tests/README.md](tests/README.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Code layout | [AGENTS.md](AGENTS.md) |
| Code of conduct | [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
