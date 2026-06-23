# Package layering (AiT toolbox — ai-ssh)

**Languages:** English (this page) · [简体中文](package-layering.zh-CN.md)

Authoritative copy for Cursor agents: [`.cursor/rules/package-layering.mdc`](../.cursor/rules/package-layering.mdc) (keep in sync with this document when conventions change).

This document defines **where code lives** in the AiT module. **AiT** is a toolbox; **`ai-ssh`** is the first tool (binary `bin/ai-ssh`, entry `cmd/ai-ssh`). It complements [go-structure](../.cursor/rules/go-structure.mdc) (file layout inside a package) and [code-reuse](../.cursor/rules/code-reuse.mdc) (when to extract shared helpers).

---

## Call chain

```text
cmd/ai-ssh/main.go          → minimal main only; delegates to aissh.Run
internal/aissh             → package aissh: Run, ExitCode, Execute, NewManager, Cobra root
internal/aissh/<cmd>       → one package per top-level subcommand (closed loop)
internal/pkg/*              → shared libraries (no product subcommand ownership)
pkg/*                       → importable by external modules; generic, non–AiT-business only
```

**Directory vs Go package:** the application root lives in **`internal/aissh/`** and declares **`package aissh`** (directory and package name match; the shipped binary remains `ai-ssh`). Subcommand directories still use package names that match the Cobra command (`login`, `shell`, `ai`, …).

---

## Subcommand package (design closure)

Each **registered top-level** `ai-ssh <subcommand>` (e.g. `get`, `put`, `shell`, `ai`, `session`, `config`) lives under **`internal/aissh/<subcommand>/`**. **SSH connect** uses OpenSSH-shaped default routing via [`internal/aissh/login/`](../internal/aissh/login/) (not registered as a Cobra subcommand) and [`internal/pkg/cliroute/`](../internal/pkg/cliroute/).

| Concern | Typical files | Notes |
|---------|---------------|--------|
| Command name | `cmd.go` or `name.go` | e.g. `const Name = "login"` — single source for Cobra `Use`, parse errors, meta alignment |
| Supported flags | `flags.go` | Flag tokens and command-specific `Apply*` |
| Argv parsing | `parse.go` | `Parse(argv)` for that command only |
| Cobra | `cmd.go` | `New(runWithMgr) *cobra.Command` — here, not a thin wrapper under another package |
| Help | `help.go` | `--help` body; flag lines use this package’s flag constants |
| Execution | `run.go` | `Run(ctx, mgr, argv)` or equivalent |

**Do not**

- Add a package whose only job is subcommand name constants (e.g. `internal/pkg/aitcmd`).
- Leave the same subcommand’s argv parsing in `internal/pkg/ssh/parse` while business logic lives in `internal/aissh/login` (long term). Shared OpenSSH mechanics belong in `internal/pkg`; command-specific flags belong in `internal/aissh/<cmd>`.

---

## `internal/aissh` (root package `aissh`)

- **Entry and root wiring:** `Run`, `ExitCode`, `newRootCmd`, `Execute` / `ExecuteWriters`, `NewManager`, `runWithManager`, init hooks (`init_hooks.go` wires `meta`, `shell`).
- **Go package name:** `aissh` (directory `internal/aissh`).
- Attach subcommands via **`registry.AttachRoot`**; each **`internal/aissh/<cmd>`** implements `registry.TopLevel` in `register.go`.
- Avoid thin wrapper packages that only wrap Cobra while parse/flags live elsewhere.

---

## `internal/pkg/*`

Use when **two or more** subcommands need the same **non–subcommand-specific** behavior, for example:

- OpenSSH-compatible flags (`-p`, `-l`, `-i`, `-F`, `-J`, `-o`, noop/unsupported tables)
- Destination parsing (`user@host`, `ssh://`, bracket hosts)
- `store`, `session`, `ssh/client`, SFTP `transfer`, `config`, `creds`, `ai` router, …

**Import rule:** `internal/pkg` **must not** import `internal/aissh`. Dependencies flow **down** only.

When `internal/pkg` must recognize a subcommand name (e.g. meta `!ait get`):

- Wire handlers from `internal/aissh` `init`, or
- Keep protocol constants in `internal/pkg` and **test** they match `internal/aissh/<cmd>.Name`.

---

## `pkg/*` (repository root)

- Libraries meant for **external** import, not tied to AiT product semantics.
- Do not place SSH connect business logic in `internal/pkg` root helpers; keep it in `internal/aissh/login` + `cliroute`.
- Prefer `internal/pkg` for module-private shared code.

---

## `helpdoc`

- **`internal/pkg/helpdoc`** holds shared builders (`Body`, OpenSSH blocks, `Bind`, `WriteTransferHelp`, TTY keys).
- Per-command flag lists and long `--help` belong in the **command package** (`internal/aissh/<cmd>/help.go`).
- Root `ai-ssh --help` uses `registry.RootHelpEntries()`; nested parents use their **`subregistry.Entries()`**.

---

## Package name = command name

- **Exception — toolbox root:** directory `internal/aissh/`, Go package **`aissh`** (matches the directory; binary name remains `ai-ssh`).
- Otherwise, directory and Go package name match the Cobra command (`login`, `shell`, `ai`, `config`, `ai/ask`, `config/set`, …).
- Do not suffix packages with `cmd` (e.g. ~~`shellcmd`~~). The file `cmd.go` is fine for Cobra wiring.
- REPL for `ai-ssh ai` lives in **`internal/aissh/ai/mode`** (package `mode`), distinct from one-shot **`ai/ask`** and **`ai/agent`**.
- Top-level `session` CLI is **`internal/aissh/session`**; import the session manager as `pkgsession "…/internal/pkg/session"`.
- **`ai-ssh shell` REPL builtins** (`exit`, `help`, `ll`, …) live in **`internal/aissh/shell/builtin/`** (package `builtin`); parent `shell` wires REPL I/O and non-builtin dispatch (`exec_dispatch.go`). SSH-style connect is **not** a builtin; `login` prefix is rejected.

---

## Nested subcommands (`ai-ssh ai`, `ai-ssh config`, …)

Subcommands of a parent command live **under the parent’s package tree**, with the same closure as top-level commands:

| Parent | Registry | Child packages (examples) |
|--------|----------|---------------------------|
| `ai-ssh` | `internal/aissh/registry` | `get`, `ai`, `config`, `session`, … (+ default SSH route) |
| `ai-ssh ai` | `internal/aissh/ai/subregistry` | `ask`, `agent`, `provider`, `ollama` |
| `ai-ssh ai provider` | `internal/aissh/ai/provider/subregistry` | `enable`, `disable` |
| `ai-ssh ai ollama` | `internal/aissh/ai/ollama/subregistry` | `status`, `setup`, `start`, `stop`, `disable` |
| `ai-ssh config` | `internal/aissh/config/subregistry` | `set`, `get`, `init`, `view`, `edit`, `validate` |

Each node: `register.go` (`init` → parent `subregistry.Register`), `cmd.go` (`New`), `help.go` (`WriteHelp`), run logic. Wire children via **`registry/all/imports.go`** blank imports—**do not** grow parent `cmd.go` `AddCommand` lists.

`ai-ssh session` has **no** Cobra children; use flag-driven actions in **`internal/aissh/session`** (`list.go`, `load.go`, …).

---

## Shared extraction examples

| Package | Purpose |
|---------|---------|
| `internal/pkg/ssh/openssh` | Shared connection argv flags for `login`, `get`, `put` |
| `internal/pkg/ssh/transferparse` | Path classification for get/put only if both would duplicate large blocks |

Command-specific flags (`--alias`, `--save`, `--tmux` for login; `--session` for get/put) stay in `internal/aissh/<cmd>/flags.go`.

---

## Command entry surfaces

AiT exposes three **independent** command registration paths. New user-facing verbs must be added at the correct layer; they are not interchangeable.

| Entry | Mechanism | Extension |
|-------|-----------|-----------|
| **Cobra subcommands** | `registry` / `subregistry` + per-package `register.go` | New package under `internal/aissh/<cmd>/` (or nested) + blank import in `registry/all/imports.go` |
| **`ai-ssh shell` builtins** | [`shell/builtin/registry.go`](../internal/aissh/shell/builtin/registry.go) map | Add name to `builtin/constants.go` + handler in registry map |
| **SSH meta / AI REPL** | [`meta/exec.go`](../internal/aissh/meta/exec.go), [`ai/mode/exec.go`](../internal/aissh/ai/mode/exec.go) | Meta: delegate to Cobra or block; AI REPL: meta verbs in `mode` |

**Special case — `ai-ssh ai model list`:** handled as a **parent positional** routed into the AI REPL meta path (R15), not as a Cobra child under `ai/model`. See [design.md](design.md) § AI CLI.

Shell builtins (`list`, `load`, …) are **not** Cobra children of `ai-ssh shell`; SSH `!ait <subcmd>` reuses top-level Cobra via [`meta.Execute`](../internal/aissh/meta/exec.go) except `ai` / `!!` which enter AI mode. Chain connect uses `!ait [flags] dest`, not `!ait login`.

**Try-block SSOT:** shared `ai-ssh session …` / `ai-ssh …` example lines for errhint live in [`internal/pkg/clihints/try_lines.go`](../internal/pkg/clihints/try_lines.go); `internal/pkg/session` re-exports them for convenience.

---

## Extending top-level commands (Open-Closed)

Use **`internal/aissh/registry`**: each top-level command implements `registry.TopLevel` and calls `registry.Register` from `init()` in `register.go`.

- **Do not** add names to `cli/root.go` `AddCommand` lists; root help comes from `registry.RootHelpEntries()` / [`registry/root_help.go`](../internal/aissh/registry/root_help.go).
- **Do** add package `internal/aissh/<cmd>/` with `register.go` + command closure files.
- **Wire imports**: add `_ "github.com/os-artificer/AiT/internal/aissh/<cmd>"` to [`registry/all/imports.go`](../internal/aissh/registry/all/imports.go) manually, or run `go generate ./internal/aissh/registry/...` to regenerate top-level blank imports (nested subregistry imports remain manual).

Adding a command must not require editing other subcommand packages.

Refactors (registry, parse moves) must preserve existing behavior: full `make test`, **R1–R19**, **A20–A22** per [change-acceptance.mdc](../.cursor/rules/change-acceptance.mdc)—no weakened tests or changed user-visible errors/exit codes on untouched paths.

## Checklist (new or refactored subcommand)

- [ ] `internal/aissh/<cmd>/` has Name, flags, Parse, Cobra `New`, Help, Run, **`register.go`**
- [ ] Registered with `registry.Register` in `init()`
- [ ] No edits to other subcommand packages or `cli/root.go` command list
- [ ] No “names only” package under `internal/pkg`
- [ ] `internal/pkg` does not import `internal/aissh`
- [ ] `cmd/ai-ssh` stays thin `main`
- [ ] Cross-command reuse is in `internal/pkg`, not copy-pasted across `ai-ssh` packages
