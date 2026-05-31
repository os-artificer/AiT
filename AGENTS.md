# AiT — Agent orientation

**Languages:** English (this page) · [简体中文](AGENTS.zh-CN.md)

**AiT** (AI Terminal) is a Go CLI: an AI-enhanced SSH client with SQLite session storage, optional remote tmux persistence, multi-provider AI (Deepseek, MiniMax, optional tiered Ollama), and SFTP file get/put. Module: `github.com/os-artificer/AiT`. Layout follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

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

- **SSH sessions**: `shell`, `login`, `session` (`--list`, `--load`, `--suspend`, `--switch`, `--delete`)
- **Persistence**: SQLite (`internal/pkg/store`); scrollback/cwd on suspend; auth TTL via `session.auth_ttl` (default 7d; keyring for secrets)
- **Optional tmux**: `--tmux` on `login`/`load`/`switch` for remote shell keep-alive (`internal/pkg/tmux`)
- **AI**: Multi-provider (`internal/pkg/ai` router + `tiered.go`); `ai.providers.<id>.enabled` gates routing; `ait ai chat` (one-shot); `ait ai provider enable|disable`; `ait ai ollama` (setup/status/start/stop/disable; `models_dir`/`OLLAMA_MODELS`; install `[y/N]`); REPL `model` / `ait ai model list`; SSH `!ait ai chat` / `!! chat`; cloud keys `ai.<provider>.api_key`
- **Transfer**: `ait get` / `ait put` — SFTP single-file transfer (`internal/pkg/transfer`); `!ait get` / `!ait put` reuse live SSH (unified router: lrzsz/ZMODEM first, SFTP fallback on direct hop, chain auto-install when missing lrzsz)
- **Proxy chain support**: `!ait chain [on|off]` — PTY-pipe ZMODEM file transfer and tmux when SSH'd beyond the directly connected host (`internal/pkg/ssh/client/inject.go`, `transfer_zmodem.go`; `internal/pkg/tmux/remote_pty.go`)
- **OpenSSH config**: `login`/`load` resolve `~/.ssh/config` via `ssh -G` (`internal/pkg/ssh/args`); data plane is Go `x/crypto/ssh`

Full design: [`docs/design.md`](docs/design.md). Example config: [`configs/ait.example.yaml`](configs/ait.example.yaml). Archive Cursor plans to `docs/<kebab-case>.md` per [`docs-exec-plans.mdc`](.cursor/rules/docs-exec-plans.mdc).

## Where code lives

| Path | Role |
|------|------|
| `cmd/ait/main.go` | Thin `main` — delegates to `internal/ait.Run` |
| `internal/ait/` | CLI entry (`Run`, `ExitCode`); subpackages below |
| `internal/ait/cli/` | Cobra command tree facade: `Execute` / `ExecuteWriters`, `newRootCmd`, manager glue (subcommands wired from subpackages below) |
| `internal/ait/cli/ai/` | `ait ai` command (`chat`, `provider enable|disable`, local AI REPL runner) |
| `internal/ait/cli/config/` | `ait config` command (`set`, `get`, `init`, `view`, `edit`, `validate`) |
| `internal/ait/cli/session/` | `ait session` command (`--list`, `--load`, `--suspend`, `--switch`, `--delete`) |
| `internal/ait/cli/transfer/` | `ait get` / `ait put` SFTP commands |
| `internal/ait/cli/ollama/` | `ait ai ollama` subcommands (`status`, `setup`, `start`, `stop`, `disable`) |
| `internal/ait/cli/login/` | `ait login` OpenSSH-compatible command |
| `internal/ait/helpdoc/` | Login-style `--help` text (`Write*Help`, OpenSSH shared blocks); `Bind` / `BindCustomArgv` Cobra helpers |
| `internal/ait/app/` | `NewManager`, `CommandContext` (store/config/ai bootstrap) |
| `internal/ait/login/` | OpenSSH login parse/resolve, help, exit codes |
| `internal/ait/shell/` | `ait shell` REPL (`AiT>`), TTY I/O, line execution |
| `internal/ait/meta/` | In-session `!ait <subcmd>` delegation |
| `internal/pkg/ssh/args/` | OpenSSH argv parse, `ssh -G` resolve, `Resolved` → `client.Options` |
| `internal/pkg/session/` | Login, load, suspend, switch, delete, list (`manager.go`, `manager_login.go`, `manager_load.go`, `manager_suspend_switch.go`, `manager_list_pick.go`, `manager_helpers.go`) |
| `internal/pkg/store/` | SQLite schema, migrations, CRUD (`store.go`, `store_sessions.go`, `store_state.go`, `store_meta.go`, `store_scan.go`, `session_helpers.go`, `open_helpers.go`) |
| `internal/pkg/ssh/client/` | `Client`, `Options`, `Dial`, `DialChain`, `RunInteractive` facade, silent PTY injection (see file layout below) |
| `internal/pkg/ssh/client/prompt/` | AiT display prompt strings + remote PS1 setup fragments |
| `internal/pkg/ssh/client/scrollback/` | Remote PTY scrollback `Delta` + capture timing constants |
| `internal/pkg/ssh/client/localmode/` | `!ait local` PTY `Shell`, env, bash/zsh init scripts |
| `internal/pkg/ssh/client/meta/` | `!ait` / `!!` byte-level scanner + meta line constants |
| `internal/pkg/ssh/client/dial/` | ProxyCommand stdio conn, jump parser, OpenSSH key signer |
| `internal/pkg/ssh/client/pty/` | TTY stdin reader, REPL byte channel, cross-platform resize watcher |
| `internal/pkg/ssh/client/transfer/` | `!ait get|put` meta argv parsing + usage hint |
| `internal/pkg/tmux/` | Remote tmux (SFTP-based + PTY-injection-based for chain mode) |
| `internal/pkg/terminal/` | Terminal facade (`path.go`) with type aliases to subpackages |
| `internal/pkg/terminal/scrollback/` | Ring-buffer scrollback `Buffer` (write/replay, file IO) |
| `internal/pkg/terminal/screen/` | ANSI screen primitives (clear) |
| `internal/pkg/terminal/crlf/` | CRLF-normalizing `io.Writer` |
| `internal/pkg/terminal/lineedit/` | `ReadLine` (Ctrl+A/E, history nav, Ctrl+R search) |
| `internal/pkg/terminal/history/` | `LineHistory` (in-memory + file sync) and `history` command |
| `internal/pkg/ai/` | `Client`, `Chat`, multi-provider routing facade |
| `internal/ait/aimode/risk/` | `Classify`, `Interactive`, `ClassifyAit` — command risk classification |
| `internal/ait/aimode/approve/` | `Confirm`, `Policy`, `PatternKey`, `SessionTrust` — approval prompts and trust |
| `internal/ait/aimode/agent/` | `Action`, `ParseAction`, `Run`, `CapturingRunner` — Agent mode execution |
| `internal/ait/aimode/chat/` | `RunOnce`, `ParseCommand` — AI chat one-shot |
| `internal/pkg/ai/provider/` | `ProviderID` constants and sentinel errors |
| `internal/pkg/ai/router/` | Routing helpers (Strategy + Chain of Responsibility — failover order, classify, API-key hint) |
| `internal/pkg/ai/agent/` | Function-call `Tools()` + tool-call → AiT Agent JSON adapter |
| `internal/pkg/ai/models/` | `ait ai model list` row formatting + role / key labels |
| `internal/pkg/ollama/` | Ollama setup, `FetchTags`, `EnsureLocal`, `ManualSetupInstructions` |
| `internal/pkg/transfer/` | SFTP get/put (single file) |
| `internal/pkg/platform/` | `AppRoot` / `DataDir` / `ConfigDir` (`AIT_HOME`), `ExpandUserAbsPath`, paths, OpenSSH |
| `internal/pkg/config/` | YAML config (`Load`/`Save`/`EnsureFile`/`Validate`; `ait config init/view/edit/validate`) |
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
| `ait/app` | `manager.go`, `context.go`, `run.go` (`Runner`) |
| `ait/cli` | `root.go`, `manager.go`, integration tests; subcommands live in `cli/{ai,config,session,transfer,ollama,login}` |
| `ait/login` | `run.go`, `helpers.go`, `help.go` |
| `ait/shell` | `repl.go`, `exec.go`, `parse.go`, `io.go`, `*_helpers.go` |
| `ait/aimode` | `scope.go`, `mode.go`, `repl.go`, `exec.go`, `run_local.go`, `agent.go` — AI REPL (subpackages `risk`, `approve`, `agent`, `chat`) |
| `ait/meta` | `exec.go` — `Execute`, `SetExecutorWriters` (wired from `cli` init; CRLF stdout in SSH) |

### `internal/pkg/ssh/client` file layout

The package keeps only `Client`-bearing methods and the user-facing facade
(`Dial`, `DialChain`, `RunInteractive`, `RunPTYCapture`, alias files for each
subpackage's exported symbols). Receiver-free helpers live in the subpackages
listed above.

| Files | Role |
|-------|------|
| `client.go` | `Client`, `Options`, `RunInteractive`, meta dispatch, local-mode hooks, `chainMode` toggle |
| `client_dial.go`, `client_dial_config.go` | `Dial`, `DetectAuthMethod`, `buildAuth`, `buildSSHClientConfig` |
| `dialchain.go`, `dialchain_helpers.go` | `DialChain`, proxy jump, `RunRemoteCommand` |
| `hostkeys.go` | known_hosts callbacks with TTY confirm / accept-new |
| `client_interactive_helpers.go` | meta dispatch helpers, refresh remote prompt, AI/transfer entry points, `handleChainMeta` |
| `client_init_helpers.go`, `client_setup_filter_helpers.go` | PTY setup output filter (suppresses PS1 leakage) |
| `client_transfer.go`, `client_transfer_helpers.go` | `Client.handleTransfer*` (SFTP or PTY based on chainMode), transfer hint wrappers |
| `inject.go`, `inject_helpers.go` | `Client.Inject` / `Client.InjectCapture` — silent PTY command injection with history/echo suppression and capture delimiters |
| `transfer_route.go`, `internal/pkg/lrzsz/`, `internal/pkg/transfer/route/` | Unified in-session get/put routing (lrzsz probe, SFTP fallback, chain install) |
| `transfer_zmodem.go`, `zmodem_session.go`, `internal/pkg/transfer/zmodem/` | `ZModemTransfer` — lrzsz-compatible ZMODEM over PTY |
| `scrollback_capture.go` | `Client.RunPTYCapture` driving `scrollback.Delta` (uses `Inject` for history suppression) |
| `resize.go` | `sessionResize` adapter wired into `pty.Watch` |
| `ai_mode.go` | `RemoteAIParams` / `RemoteAIRunner` types for remote AI REPL |
| `prompt_alias.go`, `meta_alias.go`, `pty_alias.go` | Facade aliases re-exporting subpackage symbols under their legacy names |

Package-level helpers live in `*_helpers.go` (not mixed with struct methods in the same file).

## Conventions

- **Human contributors**: [CONTRIBUTING.md](CONTRIBUTING.md) — development workflow and Go coding standards.
- **Go coding standards**: when generating or refactoring Go, follow [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) — import groups (stdlib / module / third-party), one method-bearing struct per file, package-level functions in separate `*_helpers.go`, line width 150 (`.editorconfig`), nesting ≤2, blank line between responsibility blocks inside functions (§11), full godoc on exports.
- **Reuse**: follow [`.cursor/rules/code-reuse.mdc`](.cursor/rules/code-reuse.mdc) — search existing helpers/shared packages before duplicating logic.
- **CLI help**: all `ait <cmd> --help` uses login-style sections via [`internal/ait/helpdoc`](internal/ait/helpdoc) (`helpdoc.Bind` / `helpdoc.BindCustomArgv`); see [`.cursor/rules/cli-help.mdc`](.cursor/rules/cli-help.mdc).
- **Formatting**: tabs for Go; run `make fmt` before finishing Go changes.
- **Tests**: `*_test.go` beside implementation; `make test` must pass. Store tests use `:memory:` SQLite.
- **`internal/`**: module-private — do not import from outside this module.
- **Dependencies**: add only when needed; see `go.mod` (cobra, modernc sqlite, x/crypto/ssh, x/term, keyring, uuid, sftp, go-openai, yaml).
- **Secrets**: API keys and SSH passwords in OS keyring (`creds`); never commit or log them. Do not store passwords in SQLite.
- **Commits**: only when the user explicitly asks; never change git config or commit secrets (`.env`).

## CLI quick reference

```text
ait shell                              # AiT> REPL: login, session --list/--load, history, help, exit
ait login [ssh-options] destination [cmd ...] [--save] [--alias] [--tmux] [--password]
# ssh-options: -p -l -i -F -J -o ... ; legacy --host/--user still supported
# Requires OpenSSH client (ssh) on PATH for ssh -G (all platforms)
ait ai [--model deepseek|minimax]            # local AI REPL (AiT AI>); tiered may use Ollama first
ait ai model list
ait ai chat [message...]              # one-shot; echo "q" | ait ai chat
ait ai provider enable|disable <provider>
ait ai ollama status|setup|start|stop|disable # local Ollama (setup: models_dir + [y/N] before install)
ait session --list [-v]
ait session --load [<id|alias>] [--tmux] [--password]
ait session --switch [id|alias] [--tmux] [--password] [--force]
ait session --delete <id|alias|all> [--force]
ait get [ssh-flags] <paths...>   # ParseTransfer + Resolve + DialChain; flags/paths any order
ait put [ssh-flags] <paths...>
ait config init|view|edit|validate|set|get <key> [<value>]
```

Prompts: `AiT>` (`ait shell`), `AiT AI(Ask)>` / `AiT AI(Agent)>` (`ait ai`; `mode agent` switches), `AiT AI(Ask) user@host>` (SSH AI), `AiT user@host>` (SSH remote), `AiT Local>` (`!ait local`). Run confirm: `y` once, `c` trust first-token pattern, `s` trust session low-risk, `n` skip. In SSH: `!ait help`; `!ait <subcmd>` for local CLI (not `ai`); `!ait ai` / `!!` for remote AI REPL; `!ait local` / `remote`; `!ait suspend`; `!ait chain [on|off]` for PTY-pipe injection over multiple SSH hops. Use `ait shell` for nested `login`/`load` (returns to `AiT>` after remote `exit`). Merge gate: [change-acceptance.mdc](.cursor/rules/change-acceptance.mdc) (**R1–R18**, **D\***, **S\***, **B1–B7**, **H0a–H4d**).

## Pitfalls

- Keep `cmd/ait` thin — all logic in `internal/ait` and `internal/pkg`.
- `login` resolves config via `ssh -G` (`ssh/args`); data plane is Go `x/crypto/ssh`. Set `AIT_SSH_INSECURE=1` only for dev.
- **Default mode** does not keep remote processes alive; only `--tmux` does.
- **`session --switch` from local shell** cannot suspend an active SSH; user must `!ait suspend` inside the session first, or use `--force` (does not save current scrollback/cwd).
- `third_party/` is for vendored tools, not primary app code.
- `init/` and `deployments/` are ops configs — avoid unrelated refactors there.
- `cli` subpackages register through their `New(...)` factory in [`cli/root.go`](internal/ait/cli/root.go); add new commands in a subpackage and wire them via `root.AddCommand(...)`.
- Do not add markdown files unless the user asks (except updating `README.md` / `AGENTS.md` / `docs/` when documenting the product).
