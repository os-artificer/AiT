<p align="center">
  <img src="assets/logo.png" alt="AiT logo" width="420">
</p>

**Languages:** English (this page) · [简体中文](README.zh-CN.md)

# AiT

**AI Terminal** — A cross-platform (Linux / macOS / Windows) AI-enhanced SSH client.

## Features

While remaining compatible with OpenSSH configuration and connection flags, AiT adds session management, recovery, and in-terminal collaboration beyond a typical interactive `ssh` session.

### Session management

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| Session persistence | SQLite stores connection parameters, aliases, and authentication TTL (`session.auth_ttl`) | No session store; rely on shell history or external notes |
| Fast resume | `ait load` / `session switch`; credential-free reconnect within auth TTL (TTL refreshes when you re-enter the password) | Must specify `user@host` and flags on each connect |
| Suspend and context | `!ait suspend` saves scrollback and remote working directory | Terminal context is lost after disconnect |

### Remote keep-alive

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| Process keep-alive | `--tmux` auto attach/detach `ait-<session_id>` | Manual tmux or screen session management |

### In-session collaboration

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| AI assistant | In-session `!ait ai` / `!!` (Deepseek); install suggestions when commands are missing | No built-in assistant; separate tools required |
| Local mode | `!ait local` / `!ait remote` switch input target while SSH stays connected | Local and remote work in separate terminals |
| File transfer | `ait get` / `ait put` (SFTP, single file); `!ait get` / `!ait put` reuse the live SSH session | Must assemble scp/rsync SSH options manually |

### Local entry point

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| Unified CLI | `ait shell` REPL for `login` / `list` / `load` | No equivalent integrated REPL |

## Installation

**Prerequisites:** Go **1.25+**; **OpenSSH client** (`ssh`) on PATH for `login`/`load` to resolve `~/.ssh/config`.

**Build from source** (local build and verification):

```bash
git clone https://github.com/os-artificer/AiT.git
cd AiT
make build
./bin/ait --help
```

**`go install`** (when the module resolves via `go install`):

```bash
go install github.com/os-artificer/AiT/cmd/ait@latest
ait --help
```

Ensure `$(go env GOPATH)/bin` is on your PATH.

> Configuration and usage: [User manual](docs/user-manual.md).

## Documentation

| Document | Path |
|----------|------|
| User manual | [docs/user-manual.md](docs/user-manual.md) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Code of conduct | [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| Architecture | [docs/design.md](docs/design.md) |
| Code layout | [AGENTS.md](AGENTS.md) |
