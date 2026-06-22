<p align="center">
  <img src="assets/logo.png" alt="AiT logo" width="420">
</p>

**Languages:** English (this page) · [简体中文](README.zh-CN.md)

# AiT

**AiT** — A cross-platform (Linux / macOS / Windows) AI-enhanced command-line **toolbox** that includes, but is not limited to, terminal and SSH workflows—built to be a practical everyday tool for operations engineers.

**First tool:** [`ai-ssh`](cmd/ai-ssh) — AI-enhanced SSH client (session management, in-terminal AI, file transfer). Binary: `bin/ai-ssh`; install with `go install github.com/os-artificer/AiT/cmd/ai-ssh@latest`.

## Features

While remaining compatible with OpenSSH configuration and connection flags, AiT adds session management, recovery, and in-terminal collaboration beyond a typical interactive `ssh` session. Interactive connect and session load show the remote AiT prompt immediately; in-session `!ai-ssh get`/`put` accept optional destination defaults and report chain ZMODEM/SFTP progress through `done:` (details in the [user manual](docs/user-manual.md)).

### Session management

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| Session persistence | SQLite stores connection parameters, aliases, and authentication TTL (`session.auth_ttl`) | No session store; rely on shell history or external notes |
| Fast resume | `ai-ssh session --load` / `--switch`; credential-free reconnect within auth TTL (TTL refreshes when you re-enter the password) | Must specify `user@host` and flags on each connect |
| Suspend and context | `!ai-ssh suspend` saves scrollback and remote working directory | Terminal context is lost after disconnect |

### Remote keep-alive

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| Process keep-alive | `--tmux` auto attach/detach `ait-<session_id>` | Manual tmux or screen session management |

### In-session collaboration

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| AI mode | Local `ai-ssh ai` (`AiT AI(Agent)>`) and in-SSH `!ai-ssh ai` / `!!` (remote `run` after confirm); Deepseek + MiniMax + optional tiered Ollama | No built-in assistant; separate tools required |
| Local mode | `!ai-ssh local` / `!ai-ssh remote` switch input target while SSH stays connected | Local and remote work in separate terminals |
| File transfer | `ai-ssh get` / `ai-ssh put` (SFTP, single file); `!ai-ssh get` / `!ai-ssh put` reuse the live SSH session | Must assemble scp/rsync SSH options manually |

### Local entry point

| Capability | AiT | Typical SSH client |
|------------|-----|-------------------|
| Unified CLI | `ai-ssh shell` REPL for ssh-style connect / `session --list` / `session --load`; input `history` and Ctrl+R (see user manual) | No equivalent integrated REPL |

## Installation

**Prerequisites:** Go **1.25+**; **OpenSSH client** (`ssh`) on PATH for default connect/`load` to resolve `~/.ssh/config`.

**Build from source** (local build and verification):

```bash
git clone https://github.com/os-artificer/AiT.git
cd AiT
make build
./bin/ai-ssh --help
```

**`go install`** (requires Go **1.25+**; ZMODEM support is vendored under `pkg/third-party/` — no external `github.com/xiwh/zmodem` fetch):

```bash
go install github.com/os-artificer/AiT/cmd/ai-ssh@latest
ai-ssh --help
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
