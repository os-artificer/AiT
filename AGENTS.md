# AiT — Agent orientation

**AiT** is a Go **toolbox** (module `github.com/os-artificer/AiT`). Shipped tools: **`ait`** (toolbox REPL with plugin discovery), **`aissh`** (SSH client with sessions, SFTP/ZMODEM transfer; **no AI**), **`ai`** (Ops AI Agent plugin; experimental), **`aiimg`** (image EXIF metadata + Nominatim reverse geocode), **`ailog`** (system log viewer/search), and **`aihost`** (live host CPU/mem/disk/net/process monitor). Layout follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

## Build and test

```bash
make build       # → bin/ait, bin/aissh, bin/ai, bin/aiimg, bin/ailog, bin/aihost (+ bin/plugins/*)
make test        # go test ./...
make vet         # go vet ./...
make fmt         # go fmt ./...
make test-cross  # five-platform go build ./... (linux/darwin/windows)
./bin/ait           # toolbox REPL
./bin/aissh        # run SSH tool
./bin/ai           # Ops AI Agent (experimental)
./bin/aiimg        # run image metadata tool
./bin/ailog        # run log viewer tool
./bin/aihost       # run host monitor
```

Requires **Go 1.25+** (see `go.mod`). Windows cross-compile: `GOOS=windows GOARCH=amd64 go build -o bin/aissh.exe ./cmd/aissh`.

## What AiT does

- **SSH sessions**: default OpenSSH-style connect (`aissh [flags] destination`), `shell`, `session` (`--list`, `--load`, `--suspend`, `--switch`, `--delete`)
- **Persistence**: SQLite (`internal/pkg/store`); scrollback/cwd on suspend; auth TTL via `session.auth_ttl` (default 7d; keyring for secrets); default `AppRoot` is `~/.local/share/ait` (Linux), `~/Library/Application Support/ait` (macOS), `%LOCALAPPDATA%\ait` (Windows); override with `AIT_HOME`
- **Optional tmux**: `--tmux` on connect/`load`/`switch` for remote shell keep-alive (`internal/pkg/tmux`); `--tmux` implies `--save` and stores `login_destination` for ssh config Host aliases
- **Ops AI Agent (`ai`, experimental, off by default)**: Enable with `ai.experimental.enabled`, `AIT_AI_EXPERIMENTAL=1`, or `ai --ai-experimental` (process entry only). Binary/plugin `cmd/ai` → `internal/opsai`. Agent Loop kernel: `internal/pkg/agentcore` (ops toolbox `agentops`, UI `agentui`); chat engine `internal/pkg/ai`. Multi-provider routing; `ai.providers.<id>.enabled` gates routing; `ai ask` / `ai agent` (one-shot); `ai provider enable|disable`; `ai ollama` (setup/status/start/stop/disable; `models_dir`/`OLLAMA_MODELS`); `ai model list`; interactive Loop via **`ait agent`** (bare `ai` prints help + Try); remote via `ai connect user@host` (not `!!` / `!ait ai` in `aissh`); cloud keys `ai.<provider>.api_key`. Config `ai.*` may still be set via `aissh config`. **`aissh` has zero AI** (no `aissh ai`, no in-session AI meta).
- **Transfer**: `aissh get` / `aissh put` — SFTP single-file transfer (`internal/pkg/transfer`); `!ait get` / `!ait put` reuse live SSH (unified router: lrzsz/ZMODEM first, SFTP fallback on direct hop, chain auto-install when missing lrzsz)
- **Proxy chain support**: `!ait chain [on|off]` — PTY-pipe ZMODEM file transfer and tmux when SSH'd beyond the directly connected host (`internal/pkg/ssh/client/inject.go`, `transfer_zmodem.go`; `internal/pkg/tmux/remote_pty.go`)
- **OpenSSH config**: default connect and `load` resolve `~/.ssh/config` via `ssh -G` (`internal/pkg/ssh/args`); data plane is Go `x/crypto/ssh`
- **Image metadata (`aiimg`)**: EXIF extract via `internal/pkg/imgmeta`; reverse geocode via `internal/pkg/geocode` (Nominatim)
- **System logs (`ailog`)**: view/search via `internal/pkg/logsrc` (native tools + builtin fallbacks) and less-like paging via `internal/pkg/pager`
- **Host monitor (`aihost`)**: live multi-panel dashboard via Bubble Tea (header: USER/OS/CPU/MEM/KERNEL + datetime; wraps on narrow terminals; panels: centered title, faint percents, Braille curves (DISK/NET dual-rate), `HH:MM:SS` X-axis; TCP/UDP/ARP/ICMP protocol row; PROC samples all processes by default and fills the window height with PEER listen/remote and LOG (latest system-log line) columns; `/` fuzzy search + Enter detail overlay with matching logs and a full peer list; `x` kill with SIGTERM confirm / optional SIGKILL); metrics via `internal/pkg/hoststats` + panels in `internal/pkg/hostdash`. Runs on Linux, macOS, and Windows (`CGO_ENABLED=0`); protocol/log backends are OS-specific (see [docs/tools/aihost.md](docs/tools/aihost.md)).

Project structure: [`docs/architecture.md`](docs/architecture.md). ait: [`docs/tools/ait.md`](docs/tools/ait.md), [`docs/tools/ait-internals.md`](docs/tools/ait-internals.md). aissh: [`docs/tools/aissh.md`](docs/tools/aissh.md), [`docs/tools/aissh-internals.md`](docs/tools/aissh-internals.md). ai: [`docs/tools/ai.md`](docs/tools/ai.md), [`docs/tools/ai-internals.md`](docs/tools/ai-internals.md). aiimg: [`docs/tools/aiimg.md`](docs/tools/aiimg.md), [`docs/tools/aiimg-internals.md`](docs/tools/aiimg-internals.md). ailog: [`docs/tools/ailog.md`](docs/tools/ailog.md), [`docs/tools/ailog-internals.md`](docs/tools/ailog-internals.md). aihost: [`docs/tools/aihost.md`](docs/tools/aihost.md), [`docs/tools/aihost-internals.md`](docs/tools/aihost-internals.md). Example config: [`configs/ait.example.yaml`](configs/ait.example.yaml). Archive Cursor plans to `docs/<kebab-case>.md` per [`docs-exec-plans.mdc`](.cursor/rules/docs-exec-plans.mdc).

## Where code lives

| Path | Role |
|------|------|
| `cmd/ait/main.go` | Thin `main` — delegates to `ait.Run` |
| `cmd/aissh/main.go` | Thin `main` — delegates to `aissh.Run` |
| `cmd/ai/main.go` | Thin `main` — delegates to `opsai.Run` |
| `cmd/aiimg/main.go` | Thin `main` — delegates to `aiimg.Run` |
| `cmd/ailog/main.go` | Thin `main` — delegates to `ailog.Run` |
| `cmd/aihost/main.go` | Thin `main` — delegates to `aihost.Run` |
| `internal/ait/` | `ait` toolbox REPL (package **`ait`**: `Run`, `Execute`, plugin Discover); nested **`ait/agent`** interactive Agent Loop |
| `internal/opsai/` | Ops AI Agent CLI (package **`opsai`**: `Run`, `Execute`, ask/agent/provider/ollama/connect; experimental gate) |
| `internal/pkg/aitplugin/` | Plugin Manifest v1, Discover (`plugins/<binary>/`), embedded SSOT manifests |
| `internal/pkg/aitrelease/` | GitHub Release fetch/extract/apply for REPL `install` / `update` / `upgrade` |
| `internal/pkg/pluginbin/` | `ExecPath` + optional Runner (TTY Restore) for discovered plugins |
| `internal/img/` | `aiimg` CLI (package **`aiimg`**: `Run`, `Execute`, help, version ldflags) |
| `internal/log/` | `ailog` CLI (package **`ailog`**: `Run`, `Execute`, view/search/sources/version) |
| `internal/host/` | `aihost` CLI (package **`aihost`**: `Run`, `Execute`, top/cpu/mem/disk/net/proto/proc/snapshot) |
| `internal/pkg/logsrc/` | System log sources, filters, native/builtin backends, chain resolution |
| `internal/pkg/hoststats/` | Host CPU/mem/disk/net/process sampling and Section formatters |
| `internal/pkg/hostdash/` | aihost dashboard Panel primitives (sparklines) |
| `internal/pkg/pager/` | Less-like interactive pager |
| `internal/pkg/buildinfo/` | Shared version string formatting (aissh/ai/aiimg/ailog/aihost) |
| `internal/pkg/imgmeta/` | Image EXIF extract + human/JSON formatters |
| `internal/pkg/geocode/` | Nominatim reverse geocode HTTP client |
| `internal/ssh/` | CLI entry (Go package **`aissh`**: `Run`, `ExitCode`, `Execute`, `NewManager`, Cobra root); subpackages below |
| `internal/ssh/registry/` | Top-level command registry (OCP); root help via `registry/root_help.go` + `internal/pkg/helpdoc` |
| `internal/pkg/helpdoc/` | Shared `--help` layout (`Body`, `Bind`, OpenSSH blocks, `WriteTransferHelp`, TTY keys) |
| `internal/ssh/connect/` | SSH connect parse/run (default top-level route; not a registered subcommand) |
| `internal/ssh/get/` / `internal/ssh/put/` | `aissh get` / `aissh put` SFTP |
| `internal/ssh/shell/` | `aissh shell` REPL (`AiT>`) and top-level `aissh shell` Cobra |
| `internal/ssh/session/` | `aissh session` (flags: `--list`, `--load`, …); use `pkgsession` for `internal/pkg/session` |
| `internal/pkg/agentcore/` | Agent Loop kernel for `ait agent` / `ai connect` (`risk`, `approve`, `runner`, `RunLoop`/`RunOnce`) |
| `internal/pkg/agentops/` | Ops toolbox `DispatchToolbox` + first-turn context snapshot |
| `internal/pkg/agentui/` | Plain / TUI emitters for the Agent Loop |
| `internal/pkg/aichat/` | One-shot `ai ask` (`RunOnce`) |
| `internal/ssh/config/` | `aissh config` parent + `config/subregistry`; nested `set`, `get`, `init`, `view`, `edit`, `validate` (includes `ai.*` keys) |
| `internal/ssh/version/` | `aissh version` |
| `internal/ssh/meta/` | In-session `!ait <subcmd>` delegation |
| `internal/pkg/ssh/args/` | OpenSSH argv parse, `ssh -G` resolve, `Resolved` → `client.Options` |
| `internal/pkg/session/` | Login, load, suspend, switch, delete, list (`manager.go`, `manager_login.go`, `manager_load.go`, `manager_suspend_switch.go`, `manager_list_pick.go`, `manager_helpers.go`) |
| `internal/pkg/store/` | SQLite schema, migrations, CRUD (`store.go`, `store_sessions.go`, `store_state.go`, `store_meta.go`, `store_scan.go`, `session_helpers.go`, `open_helpers.go`) |
| `internal/pkg/ssh/client/` | `Client`, `Options`, `Dial`, `DialChain`, `RunInteractive` facade, silent PTY injection (see file layout below) |
| `internal/pkg/ssh/client/prompt/` | AiT display prompt strings + remote PS1 setup fragments |
| `internal/pkg/ssh/client/scrollback/` | Remote PTY scrollback `Delta` + capture timing constants |
| `internal/pkg/ssh/client/localmode/` | `!ait local` PTY `Shell`, env, bash/zsh init scripts |
| `internal/pkg/ssh/client/meta/` | `!ait` byte-level scanner + meta line constants (`!!` is not an AI entry) |
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
| `internal/pkg/agentcore/risk/` | `Classify`, `Interactive`, `ClassifyAit` — command risk classification |
| `internal/pkg/agentcore/approve/` | `Confirm`, `Policy`, `PatternKey`, `SessionTrust` — approval prompts and trust |
| `internal/pkg/agentcore/runner/` | `Action`, `ParseAction`, `Run`, capture runner — Agent step execution |
| `internal/pkg/ai/provider/` | `ProviderID` constants and sentinel errors |
| `internal/pkg/ai/router/` | Routing helpers (Strategy + Chain of Responsibility — failover order, classify, API-key hint) |
| `internal/pkg/ai/agent/` | Function-call `Tools()` + tool-call → AiT Agent JSON adapter |
| `internal/pkg/ai/models/` | `ai model list` row formatting + role / key labels |
| `internal/pkg/ollama/` | Ollama setup, `FetchTags`, `EnsureLocal`, `ManualSetupInstructions` |
| `internal/pkg/transfer/` | SFTP get/put (single file) |
| `internal/pkg/platform/` | `AppRoot` / `DataDir` / `ConfigDir` (`AIT_HOME`), `ExpandUserAbsPath`, paths, OpenSSH |
| `internal/pkg/config/` | YAML config (`Load`/`Save`/`EnsureFile`/`Validate`; `aissh config init/view/edit/validate`) |
| `internal/pkg/creds/` | OS keyring (API key, session unlock) |
| `internal/pkg/pathsafe/` | Session ID and path validation (scrollback sidecars) |
| `pkg/` | Public importable libs; [`pkg/third-party/`](pkg/third-party/) vendored upstream Go sources (e.g. xiwh ZMODEM) |
| `configs/` | `ait.example.yaml` |
| `docs/` | [`architecture.md`](docs/architecture.md), [`tools/aissh.md`](docs/tools/aissh.md), [`tools/ai.md`](docs/tools/ai.md), … |
| `.cursor/rules/` | Cursor rules; Go structure: [`go-structure.mdc`](.cursor/rules/go-structure.mdc) |
| `scripts/` | Build and CI helpers |

Prefer extending `internal/pkg/*` and `internal/ssh` over new top-level directories.

### `internal/ssh` layout (root Go package: `aissh`)

| Package | Files (examples) |
|---------|------------------|
| `aissh` (root) | `root.go` (`Run`, `ExitCode`); `cli_root.go` (`Execute`, `newRootCmd`); `manager.go` (`NewManager`, `runWithManager`); `context.go`, `runner.go`; `init_hooks.go` (meta/shell inject; **no AI**); integration tests |
| `registry` | `registry.go`, `top_level.go`, `deps.go`, `root_help.go`, `all/imports.go` |
| `get`, `put`, `shell`, `session`, `version` | Top-level commands: `register.go`, `cmd.go`, `help.go`, `run.go` / flags; SSH connect via default route (`internal/ssh/connect`) |
| `config` | `cmd.go`, `help.go`, `subregistry/`, `shared/`; nested `set`, `get`, `init`, `view`, `edit`, `validate` |
| `meta` | `exec.go` — `Execute`, `SetExecutorWriters` (wired from `init_hooks.go`; CRLF stdout in SSH) |

Ops AI lives under **`internal/opsai/`** (not `internal/ssh`). Agent Loop kernel: **`internal/pkg/agentcore`**; interactive entry: **`ait agent`**.

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
| `client_interactive_helpers.go` | meta dispatch helpers, refresh remote prompt, transfer entry points, `handleChainMeta` |
| `client_init_helpers.go`, `client_setup_filter_helpers.go` | PTY setup output filter (suppresses PS1 leakage) |
| `client_transfer.go`, `client_transfer_helpers.go` | `Client.handleTransfer*` (SFTP or PTY based on chainMode), transfer hint wrappers |
| `inject.go`, `inject_helpers.go` | `Client.Inject` / `Client.InjectCapture` — silent PTY command injection with history/echo suppression and capture delimiters |
| `transfer_route.go`, `internal/pkg/lrzsz/`, `internal/pkg/transfer/route/` | Unified in-session get/put routing (lrzsz probe, SFTP fallback, chain install) |
| `transfer_zmodem.go`, `zmodem_session.go`, `internal/pkg/transfer/zmodem/` | `ZModemTransfer` — lrzsz-compatible ZMODEM over PTY |
| `scrollback_capture.go` | `Client.RunPTYCapture` driving `scrollback.Delta` (uses `Inject` for history suppression) |
| `resize.go` | `sessionResize` adapter wired into `pty.Watch` |
| `prompt_alias.go`, `meta_alias.go`, `pty_alias.go` | Facade aliases re-exporting subpackage symbols under their legacy names |

Package-level helpers live in `*_helpers.go` (not mixed with struct methods in the same file).

## Conventions

- **Human contributors**: [CONTRIBUTING.md](CONTRIBUTING.md) — development workflow and Go coding standards.
- **Go coding standards**: when generating or refactoring Go, follow [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) — import groups (stdlib / module / third-party), one method-bearing struct per file, package-level functions in separate `*_helpers.go`, line width 150 (`.editorconfig`), nesting ≤2, blank line between responsibility blocks inside functions (§11), full godoc on exports.
- **Package layering**: `cmd/aissh` → `internal/ssh` (package `aissh`) / `<subcommand>` → shared code in `internal/pkg` (must not import `internal/ssh`); root `pkg/` for external, non-business libs. See [`docs/tools/aissh-internals.md#package-layering`](docs/tools/aissh-internals.md#package-layering) and [`.cursor/rules/package-layering.mdc`](.cursor/rules/package-layering.mdc).
- **Reuse**: follow [`.cursor/rules/code-reuse.mdc`](.cursor/rules/code-reuse.mdc) — search existing helpers/shared packages before duplicating logic.
- **CLI help**: shared layout in [`internal/pkg/helpdoc`](internal/pkg/helpdoc) (`pkghelpdoc.Bind`); per-command body in `internal/ssh/<cmd>/help.go` (nested: `<parent>/<sub>/help.go`); see [`.cursor/rules/cli-help.mdc`](.cursor/rules/cli-help.mdc).
- **Formatting**: tabs for Go; run `make fmt` before finishing Go changes.
- **Tests**: `*_test.go` beside implementation; `make test` must pass. Store tests use `:memory:` SQLite.
- **`internal/`**: module-private — do not import from outside this module.
- **Dependencies**: add only when needed; see `go.mod` (cobra, modernc sqlite, x/crypto/ssh, x/term, keyring, uuid, sftp, cloudwego/eino, yaml).
- **Secrets**: API keys and SSH passwords in OS keyring (`creds`); never commit or log them. Do not store passwords in SQLite.
- **Commits**: only when the user explicitly asks; never change git config or commit secrets (`.env`).

## CLI quick reference

```text
ait                                      # ait> REPL (no one-shot plugin exec)
ait> install [plugin...]                 # GitHub Release → install dir (all official plugins if no args)
ait> update [plugin...]                  # update installed official plugins only
ait> upgrade                             # update ait + installed official plugins (no extra args)
ait agent [--plain] [goal...]            # interactive Agent Loop (TTY default TUI)
ait> agent [goal...]                     # nested Loop from toolbox REPL
aissh shell                              # AiT> REPL: ssh-style connect, session --list/--load, history, help, exit
aissh [ssh-options] destination [cmd ...] [--save] [--alias] [--tmux] [--password]
# ssh-options: -p -l -i -F -J -o ... ; legacy --host/--user still supported
# Requires OpenSSH client (ssh) on PATH for ssh -G (all platforms)
# aissh has no AI subtree (use the `ai` plugin / ait agent)
ai [--ai-experimental]                   # help + Try ait agent (no REPL); experimental gate
ai model list
ai ask [message...]                      # one-shot; echo "q" | ai ask
ai agent [message...]                    # one-shot agent with tools (Loop: ait agent)
ai connect [ssh-options] destination     # remote Agent Loop on SSH host
ai provider enable|disable <provider>
ai ollama status|setup|start|stop|disable # local Ollama (setup: models_dir + [y/N] before install)
aissh session --list [-v]
aissh session --load [<id|alias>] [--tmux] [--password]
aissh session --switch [id|alias] [--tmux] [--password] [--force]
aissh session --delete <id|alias|all> [--force]
aissh get [ssh-flags] <paths...>   # ParseTransfer + Resolve + DialChain; flags/paths any order
aissh put [ssh-flags] <paths...>
aissh config init|view|edit|validate|set|get <key> [<value>]  # includes ai.* keys
```

Prompts: `AiT>` (`aissh shell` / `ait`), Agent Loop via `ait agent` / `ai connect` (plain or TUI; confirm `y`/`c`/`s`/`n`; read-only auto-run). Native shell PS1 for SSH remote interactive; `AiT Local>` (`!ait local`). In SSH (`aissh`): `!ait help`; `!ait <subcmd>` for local CLI; **no** `!ait ai` / `!!` AI entry — use `ai connect` or `ait agent`. Also `!ait local` / `remote`; `!ait suspend`; `!ait chain [on|off]` for PTY-pipe injection over multiple SSH hops. Use `aissh shell` for nested connect/`load` (returns to `AiT>` after remote `exit`). Merge gate: [change-acceptance.mdc](.cursor/rules/change-acceptance.mdc) (**OA\***, **AG\***, **AC-V4\***, **R1–R14**, **D\***, **S\***, **B1–B7**, **H0a–H4d**).

## Pitfalls

- Keep `cmd/aissh` thin — all logic in `internal/ssh` and `internal/pkg`.
- Default SSH connect (and `get`/`put`) resolve config via `ssh -G` (`ssh/args`); data plane is Go `x/crypto/ssh`. Set `AIT_SSH_INSECURE=1` only for dev.
- **Default mode** does not keep remote processes alive; only `--tmux` does.
- **`session --switch` from local shell** cannot suspend an active SSH; user must `!ait suspend` inside the session first, or use `--force` (does not save current scrollback/cwd).
- `pkg/third-party/` is for vendored tools, not primary app code.

- Top-level commands: add `internal/ssh/<cmd>/` with `register.go` (`registry.Register`) and a blank import in [`registry/all/imports.go`](internal/ssh/registry/all/imports.go). Nested commands: parent `subregistry` + `<parent>/registry/all/imports.go` (see [`docs/tools/aissh-internals.md#package-layering`](docs/tools/aissh-internals.md#package-layering)).
- Do not add markdown files unless the user asks (except updating `README.md` / `AGENTS.md` / `docs/` when documenting the product).
