# `/internal`

Private application and library code. The Go compiler enforces that code outside this module cannot import `internal/` packages.

This project follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

- **`/internal/ait`** — AiT CLI and `ait shell` REPL (Cobra commands, line parsing, meta delegation).
- **`/internal/pkg`** — Shared libraries used only by `ait`.

## `/internal/ait`

| Package | Role |
|---------|------|
| `ait` (root) | `Run`, `ExitCode` — entry from `cmd/ait` |
| `ait/app` | Store/config/AI bootstrap, `NewManager`, signal context |
| `ait/cli` | Cobra commands (`login`, `session`, `shell`, `config`, `ai`, …) |
| `ait/helpdoc` | Login-style `--help` text (`Write*Help`, shared OpenSSH blocks) |
| `ait/login` | Login parse/resolve, help text, remote exit codes |
| `ait/shell` | `ait shell` REPL, TTY I/O, line execution |
| `ait/aimode` | `ait ai` / `!ait ai` AI REPL (local/remote runners injected) |
| `ait/meta` | In-session `!ait` delegation (`!ait list`, etc.) |

Helpers live in `*_helpers.go` beside each package’s main files.

## `/internal/pkg`

| Package | Role |
|---------|------|
| `ssh/args` | Parse `ait login` argv; `ssh -G` resolve to connection options |
| `session` | Login, load, suspend, switch, delete, list |
| `store` | SQLite sessions and `session_state` |
| `ssh/client` | SSH dial, PTY, `!ait` / `!!`, local mode, host keys |
| `terminal` | Scrollback buffer, replay, `ReadLine`, REPL input history, Ctrl+R reverse search |
| `tmux` | Remote tmux attach/detach |
| `ai` | Multi-provider router, tiered Ollama, model list |
| `ollama` | Ollama setup, `FetchTags`, `EnsureLocal` |
| `transfer` | SFTP get/put (single file) |
| `platform` | `AppRoot` / `AIT_HOME` / `AIT_DATA_DIR`, paths, OpenSSH path, `LookPath` |
| `config` | YAML configuration (`Load`/`Save`/`EnsureFile`; `ait config init`) |
| `creds` | OS keyring for secrets |
| `pathsafe` | Safe session IDs and scrollback paths |

See [AGENTS.md](../AGENTS.md) for file-level layout (especially `ssh/client`) and [docs/design.md](../docs/design.md) for behavior and data model.
