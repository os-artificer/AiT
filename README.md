# AiT

**AiT** is a cross-platform (Linux / macOS / Windows) AI-enhanced command-line **toolbox** for operations engineers—terminal and SSH workflows today, more tools over time.

**Tools:**

- [`ait`](cmd/ait) — toolbox interactive REPL (discovers plugins under `bin/plugins/`). Guide: [ait](docs/tools/ait.md).
- [`aissh`](cmd/aissh) — SSH client (sessions, file transfer, shell/meta). Guide: [aissh](docs/tools/aissh.md).
- [`ai`](cmd/ai) — Ops AI Agent (experimental; local REPL, toolbox tools, optional `ai connect`). Guide: [ai](docs/tools/ai.md).
- [`aiimg`](cmd/aiimg) — image metadata extractor (EXIF, GPS, reverse-geocoded address). Guide: [aiimg](docs/tools/aiimg.md).
- [`ailog`](cmd/ailog) — system log viewer and search (Linux/macOS/Windows). Guide: [ailog](docs/tools/ailog.md).
- [`aihost`](cmd/aihost) — live host CPU/mem/disk/net/process monitor. Guide: [aihost](docs/tools/aihost.md).

## What you get

- OpenSSH-compatible connect (`-J`, config Host aliases, `ssh -G` resolution)
- Session bookmarks: `--alias`, `session --load`, `session --switch`
- In-session `!ait` meta commands (suspend, chain hop, transfer)
- Default AiT-managed remote prompt (`AiT user@host:…>`); set `AIT_REMOTE_SYSTEM_PROMPT=1` for the native shell prompt
- Experimental Ops AI Agent (`ai`) — **off by default**; enable with `ai.experimental.enabled`, `AIT_AI_EXPERIMENTAL=1`, or `ai --ai-experimental` (see [ai](docs/tools/ai.md)); `aissh` has no AI subtree
- `!ait get` / `!ait put` on the live SSH session, standalone `aissh get` / `put` via SFTP
- Optional `--tmux` to keep remote shells alive after disconnect
- Local `aissh shell` REPL with history and UTF-8 line editing

## Examples

```bash
# 1) Bookmark a host and resume later (OpenSSH-style flags still work)
aissh deploy@prod --alias prod
aissh session --list
aissh session --load prod

# 2) Jump host in one command (same as ssh -J; Host aliases resolve via ssh -G)
aissh -J bastion app.internal

# 3) Enable experimental Ops AI Agent, then use local or remote scope via `ai`
aissh config init
aissh config set ai.experimental.enabled true
ai                          # local Agent REPL
ai connect user@host        # remote Agent on an SSH host (not !! / !ait ai)

# 4) Transfer over the current session (single-arg defaults to ./basename)
!ait get ./build/app-linux-amd64
!ait put ./patch.tar.gz

# 5) tmux keep-alive: remote work survives disconnect
aissh devhost --tmux
# Later: aissh session --load devhost --tmux

# 6) One-shot local AI (non-interactive; requires experimental AI enabled)
ai ask "summarize this log pattern: connection reset"
```

## Installation

**Prerequisites:** Go **1.25+** if building from source; **OpenSSH client** (`ssh`) on PATH for `aissh` connect/`load` to resolve `~/.ssh/config`.

### Install `ait`

```bash
go install github.com/os-artificer/AiT/cmd/ait@latest
ait --help
```

Ensure `$(go env GOPATH)/bin` is on your PATH.

### Install official plugins (`aissh`, `ai`, `aiimg`, `ailog`, `aihost`)

After a [GitHub Release](https://github.com/os-artificer/AiT/releases) is available, use the **ait** REPL (no separate `go install` for plugins):

```bash
ait
```

```text
ait> install              # all official plugins
ait> install aissh ai     # one or more by name
ait> update               # update installed plugins
ait> upgrade              # update ait + installed plugins
ait> help install
```

Binaries are written next to the running `ait` executable (same layout as `make install`). Details: [docs/tools/ait.md](docs/tools/ait.md#install--update--upgrade).

### Build from source (development)

```bash
git clone https://github.com/os-artificer/AiT.git
cd AiT
make build
./bin/ait --help
```

For local development, `make build` also produces `bin/aissh`, `bin/ai`, `bin/aiimg`, `bin/ailog`, and `bin/aihost`. To install plugins into `bin/` without a Release, run `make install` or use `ait> install` once a Release exists. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Documentation

| Document | Path |
|----------|------|
| Project structure | [docs/architecture.md](docs/architecture.md) |
| aissh internals | [docs/tools/aissh-internals.md](docs/tools/aissh-internals.md) |
| aissh guide | [docs/tools/aissh.md](docs/tools/aissh.md) |
| ai internals | [docs/tools/ai-internals.md](docs/tools/ai-internals.md) |
| ai guide | [docs/tools/ai.md](docs/tools/ai.md) |
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
