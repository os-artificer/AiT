# AiT — Agent orientation

**Languages:** English (this page) · [简体中文](AGENTS.zh-CN.md)

**AiT** (AI Terminal) is a Go CLI: an AI-enhanced SSH client with SQLite session storage, optional remote tmux persistence, Deepseek assistant, and SFTP file get/put. Module: `github.com/os-artificer/AiT`. Layout follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

## Build and test

```bash
make build       # → bin/ait
make test        # go test ./...
make vet         # go vet ./...
make fmt         # go fmt ./...
make test-cross  # windows + darwin go build
./bin/ait        # run binary
```

Requires **Go 1.25+** (see `go.mod`). Windows cross-compile: `GOOS=windows GOARCH=amd64 go build -o bin/ait.exe ./cmd/ait`.

## What AiT does

- **SSH sessions**: `shell`, `login`, `list`, `load`, `suspend`, `session switch`, `session delete`
- **Persistence**: SQLite (`internal/pkg/store`); scrollback/cwd on suspend; auth TTL via `session.auth_ttl` (default 7d; keyring for secrets)
- **Optional tmux**: `--tmux` on `login`/`load`/`switch` for remote shell keep-alive (`internal/pkg/tmux`)
- **AI**: Deepseek via OpenAI-compatible API (`internal/pkg/ai`); in-session `!ait ai` / `!!`; `!ait local`/`!ait remote`; install hints when `rsync`/`scp`/`tmux` missing
- **Transfer**: `ait get` / `ait put` — SFTP single-file transfer (`internal/pkg/transfer`); `!ait get` / `!ait put` reuse live SSH
- **OpenSSH config**: `login`/`load` resolve `~/.ssh/config` via `ssh -G` (`internal/pkg/ssh/args`); data plane is Go `x/crypto/ssh`

Full design: [`docs/design.md`](docs/design.md). Example config: [`configs/ait.example.yaml`](configs/ait.example.yaml). Archive Cursor plans to `docs/<kebab-case>.md` per [`docs-exec-plans.mdc`](.cursor/rules/docs-exec-plans.mdc).

## Where code lives

| Path | Role |
|------|------|
| `cmd/ait/main.go` | Thin `main` — delegates to `internal/ait.Run` |
| `internal/ait/` | CLI entry (`Run`, `ExitCode`); subpackages below |
| `internal/ait/cli/` | Cobra command tree, `Execute` / `ExecuteWriters` |
| `internal/ait/app/` | `NewManager`, `CommandContext` (store/config/ai bootstrap) |
| `internal/ait/login/` | OpenSSH login parse/resolve, help, exit codes |
| `internal/ait/shell/` | `ait shell` REPL (`AiT>`), TTY I/O, line execution |
| `internal/ait/meta/` | In-session `!ait <subcmd>` delegation |
| `internal/pkg/ssh/args/` | OpenSSH argv parse, `ssh -G` resolve, `Resolved` → `client.Options` |
| `internal/pkg/session/` | Login, load, suspend, switch, delete, list (`manager.go`, `manager_helpers.go`) |
| `internal/pkg/store/` | SQLite schema, migrations, CRUD (`store.go`, `store_scan.go`, `session_helpers.go`, `open_helpers.go`) |
| `internal/pkg/ssh/client/` | Dial, PTY, `!ait` meta, local mode, host keys (see file layout below) |
| `internal/pkg/tmux/` | Remote tmux ensure/attach/detach |
| `internal/pkg/terminal/` | Scrollback buffer, replay, path resolution (`scrollback.go`, `scrollback_io.go`, `path.go`, helpers) |
| `internal/pkg/ai/` | Deepseek client (`client.go`, `client_helpers.go`) |
| `internal/pkg/transfer/` | SFTP get/put (single file) |
| `internal/pkg/platform/` | Data/config dirs, `OpenSSHPath`, `LookPath`, OS paths |
| `internal/pkg/config/` | YAML config (`config.go` types; `config_load.go` Load/Save) |
| `internal/pkg/creds/` | OS keyring (API key, session unlock) |
| `internal/pkg/pathsafe/` | Session ID and path validation (scrollback sidecars) |
| `pkg/` | Reserved for public importable libs (empty today) |
| `configs/` | `ait.example.yaml` |
| `docs/` | [`design.md`](docs/design.md), [`user-manual.md`](docs/user-manual.md) |
| `.cursor/rules/` | Cursor rules; Go structure: [`go-structure.mdc`](.cursor/rules/go-structure.mdc) |
| `scripts/`, `build/` | Build and CI helpers |

Prefer extending `internal/pkg/*` and `internal/ait` over new top-level directories.

### `internal/ait` layout

| Package | Files (examples) |
|---------|------------------|
| `ait` (root) | `root.go` — `Run`, `ExitCode` |
| `ait/app` | `manager.go`, `context.go` |
| `ait/cli` | `root.go`, `login.go`, `shell.go`, `load.go`, `session.go`, … |
| `ait/login` | `run.go`, `helpers.go`, `help.go` |
| `ait/shell` | `repl.go`, `exec.go`, `parse.go`, `io.go`, `*_helpers.go` |
| `ait/meta` | `exec.go` — `Execute`, `SetExecutorWriters` (wired from `cli` init; CRLF stdout in SSH) |

### `internal/pkg/ssh/client` file layout

| Files | Role |
|-------|------|
| `client.go` | `Client`, `Options`, `RunInteractive`, meta handling |
| `client_dial.go` | `Dial`, `DetectAuthMethod`, `buildAuth`, `loadSigner` |
| `client_interactive_helpers.go` | stdin reader, term print, pwd capture, forward helpers |
| `dialchain.go`, `dialchain_helpers.go` | `DialChain`, proxy jump, `RunRemoteCommand` |
| `hostkeys.go` | known_hosts callbacks |
| `metascanner.go`, `metascanner_helpers.go` | `!ait` / `!!` line detection |
| `localmode.go`, `localmode_init.go`, `localmode_init_helpers.go` | `!ait local` PTY |
| `prompt.go` | Remote/local prompt strings |
| `resize.go`, `resize_unix.go`, `resize_windows.go`, `resize_helpers.go` | Window resize |
| `proxy_conn.go`, `proxy_unix.go`, `proxy_windows.go` | ProxyCommand stdio conn |

Package-level helpers live in `*_helpers.go` (not mixed with struct methods in the same file).

## Conventions

- **Human contributors**: [CONTRIBUTING.md](CONTRIBUTING.md) — development workflow and Go coding standards.
- **Go coding standards**: when generating or refactoring Go, follow [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) — import groups (stdlib / module / third-party), one method-bearing struct per file, package-level functions in separate `*_helpers.go`, line width 150 (`.editorconfig`), nesting ≤2, blank line between responsibility blocks inside functions (§11), full godoc on exports.
- **Formatting**: tabs for Go; run `make fmt` before finishing Go changes.
- **Tests**: `*_test.go` beside implementation; `make test` must pass. Store tests use `:memory:` SQLite.
- **`internal/`**: module-private — do not import from outside this module.
- **Dependencies**: add only when needed; see `go.mod` (cobra, modernc sqlite, x/crypto/ssh, x/term, keyring, uuid, sftp, go-openai, yaml).
- **Secrets**: API keys and SSH passwords in OS keyring (`creds`); never commit or log them. Do not store passwords in SQLite.
- **Commits**: only when the user explicitly asks; never change git config or commit secrets (`.env`).

## CLI quick reference

```text
ait shell                              # AiT> REPL: login, list, load, help, exit
ait login [ssh-options] destination [cmd ...] [--alias] [--tmux] [--password]
# ssh-options: -p -l -i -F -J -o ... ; legacy --host/--user still supported
# Requires OpenSSH client (ssh) on PATH for ssh -G (all platforms)
ait list | load <id|alias> [--tmux] [--password]
ait session switch [id|alias] [--tmux] [--force]
ait session delete <id|alias> | ait session delete all [--force]
ait get [ssh-flags] <paths...>   # ParseTransfer + Resolve + DialChain; flags/paths any order
ait put [ssh-flags] <paths...>
ait config set|get <key> [<value>]
```

Prompts: `AiT>` (`ait shell`), `AiT user@host>` (SSH remote), `AiT Local>` (`!ait local`). In SSH: `!ait help`; `!ait <subcmd>` for local CLI; `!ait ai` / `!!`; `!ait local` / `remote`; `!ait suspend`. Use `ait shell` for nested `login`/`load` (returns to `AiT>` after remote `exit`).

## Pitfalls

- Keep `cmd/ait` thin — all logic in `internal/ait` and `internal/pkg`.
- `login` resolves config via `ssh -G` (`ssh/args`); data plane is Go `x/crypto/ssh`. Set `AIT_SSH_INSECURE=1` only for dev.
- **Default mode** does not keep remote processes alive; only `--tmux` does.
- **`session switch` from local shell** cannot suspend an active SSH; user must `!ait suspend` inside the session first, or use `--force` (does not save current scrollback/cwd).
- `third_party/` is for vendored tools, not primary app code.
- `init/` and `deployments/` are ops configs — avoid unrelated refactors there.
- Do not add markdown files unless the user asks (except updating `README.md` / `AGENTS.md` / `docs/` when documenting the product).
