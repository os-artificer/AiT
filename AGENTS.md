# AiT — Agent orientation

**AiT** (AI Terminal) is a Go CLI: an AI-enhanced SSH client with SQLite session storage, optional remote tmux persistence, Deepseek assistant, and cross-host file sync. Module: `github.com/os-artificer/AiT`. Layout follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

## Build and test

```bash
make build    # → bin/ait
make test     # go test ./...
make vet      # go vet ./...
make fmt      # go fmt ./...
./bin/ait     # run binary
```

Requires **Go 1.25+** (see `go.mod`). Windows cross-compile: `GOOS=windows GOARCH=amd64 go build -o bin/ait.exe ./cmd/ait`.

## What AiT does

- **SSH sessions**: `login`, `list`, `load`, `suspend`, `session switch`, `session delete`
- **Persistence**: SQLite (`internal/pkg/store`); scrollback/cwd on suspend; 7-day auth TTL (keyring for secrets)
- **Optional tmux**: `--tmux` on `login`/`load`/`switch` for remote shell keep-alive (`internal/pkg/tmux`)
- **AI**: Deepseek via OpenAI-compatible API (`internal/pkg/ai`); in-session `!ait ai` / `!!` prompts; `!ait local`/`!ait remote` local mode; install hints when `rsync`/`scp`/`tmux` missing
- **Sync**: `ait sync` — rsync → scp → SFTP (`internal/pkg/sync`)

Full design: [`docs/design.md`](docs/design.md). Example config: [`configs/ait.example.yaml`](configs/ait.example.yaml).

## Where code lives

| Path | Role |
|------|------|
| `cmd/ait/main.go` | Thin `main` — delegates to `internal/ait.Run` |
| `internal/ait/` | Cobra root and subcommands (`cmd_*.go`, `root.go`, `context.go`) |
| `internal/pkg/store/` | SQLite schema, migrations, session CRUD |
| `internal/pkg/session/` | Login, load, suspend, switch, delete, list |
| `internal/pkg/sshclient/` | Dial, PTY loop, `!ait` meta (`ai`/`local`/`remote`/`suspend`), `!!` alias |
| `internal/pkg/tmux/` | Remote tmux ensure/attach/detach |
| `internal/pkg/terminal/` | Scrollback buffer and replay |
| `internal/pkg/ai/` | Deepseek client |
| `internal/pkg/sync/` | rsync/scp/SFTP sync |
| `internal/pkg/platform/` | Data/config dirs, `LookPath`, OS description |
| `internal/pkg/config/` | YAML config load/save |
| `internal/pkg/creds/` | OS keyring (API key, session unlock) |
| `pkg/` | Reserved for public importable libs (empty today) |
| `configs/` | `ait.example.yaml` |
| `docs/` | `design.md` architecture and behavior |
| `scripts/`, `build/` | Build and CI helpers |

Prefer extending `internal/pkg/*` and `internal/ait` over new top-level directories.

## Conventions

- **Formatting**: tabs for Go (`.editorconfig`); run `make fmt` before finishing Go changes.
- **Tests**: `*_test.go` beside implementation; `make test` must pass. Store tests use `:memory:` SQLite.
- **`internal/`**: module-private — do not import from outside this module.
- **Dependencies**: add only when needed; key deps: cobra, modernc.org/sqlite, golang.org/x/crypto/ssh, go-openai (Deepseek base URL), keyring, pkg/sftp.
- **Secrets**: API keys and SSH passwords go in OS keyring (`creds`); never commit or log them. Do not store passwords in SQLite.
- **Commits**: only when the user explicitly asks; never change git config or commit secrets (`.env`).

## CLI quick reference

```text
ait login --host H [--alias] [--user] [--port] [--identity] [--password] [--tmux]
ait list | load <id|alias> [--tmux] [--password]
ait session switch [id|alias] [--tmux]
ait session delete <id|alias> | ait session delete all [--force]
ait sync <src> <dst> [--session <id|alias>]
ait config set|get <key> [<value>]
```

Prompts: `AiT>` (`ait shell`), `AiT user@host>` (SSH remote), `AiT Local>` (`!ait local`). In SSH: `!ait help`; `!ait <subcmd>` for local CLI; `!ait ai` / `!!`; `!ait local` / `remote`; `!ait suspend`. Use `ait shell` for nested `login`/`load` (returns to `AiT>` after remote `exit`).

## Pitfalls

- Keep `cmd/ait` thin — all logic in `internal/ait` and `internal/pkg`.
- `sshclient` uses `InsecureIgnoreHostKey` today — do not weaken further; prefer adding `known_hosts` support when touching auth.
- **Default mode** does not keep remote processes alive; only `--tmux` does.
- **`session switch` from local shell** cannot suspend an active SSH; user must `\ait suspend` inside the session first.
- `third_party/` is for vendored tools, not primary app code.
- `init/` and `deployments/` are ops configs — avoid unrelated refactors there.
- Do not add markdown files unless the user asks (except updating `README.md` / `AGENTS.md` / `docs/` when documenting the product).
