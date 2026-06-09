# Package layering (AiT CLI)

**Languages:** English (this page) · [简体中文](package-layering.zh-CN.md)

Authoritative copy for Cursor agents: [`.cursor/rules/package-layering.mdc`](../.cursor/rules/package-layering.mdc) (keep in sync with this document when conventions change).

This document defines **where code lives** in the AiT module. It complements [go-structure](../.cursor/rules/go-structure.mdc) (file layout inside a package) and [code-reuse](../.cursor/rules/code-reuse.mdc) (when to extract shared helpers).

---

## Call chain

```text
cmd/ait/main.go          → minimal main only
internal/ait             → Run, ExitCode, Execute, NewManager, Cobra root (thin)
internal/ait/<cmd>       → one package per top-level subcommand (closed loop)
internal/pkg/*           → shared libraries (no product subcommand ownership)
pkg/*                    → importable by external modules; generic, non–AiT-business only
```

---

## Subcommand package (design closure)

Each **top-level** `ait <subcommand>` (e.g. `login`, `get`, `put`, `shell`, `ai`, `session`, `config`) lives under **`internal/ait/<subcommand>/`** and owns its vertical slice in **one package**:

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
- Leave the same subcommand’s argv parsing in `internal/pkg/ssh/parse` while business logic lives in `internal/ait/login` (long term). Shared OpenSSH mechanics belong in `internal/pkg`; command-specific flags belong in `internal/ait/<cmd>`.

---

## `internal/ait` (root package)

- **Entry and root wiring:** `Run`, `ExitCode`, `newRootCmd`, `Execute` / `ExecuteWriters`, `NewManager`, `runWithManager`, init hooks (`init_hooks.go` wires `meta`, `shell`).
- Attach subcommands via **`registry.AttachRoot`**; each **`internal/ait/<cmd>`** implements `registry.TopLevel` in `register.go`.
- Avoid thin wrapper packages that only wrap Cobra while parse/flags live elsewhere.

---

## `internal/pkg/*`

Use when **two or more** subcommands need the same **non–subcommand-specific** behavior, for example:

- OpenSSH-compatible flags (`-p`, `-l`, `-i`, `-F`, `-J`, `-o`, noop/unsupported tables)
- Destination parsing (`user@host`, `ssh://`, bracket hosts)
- `store`, `session`, `ssh/client`, SFTP `transfer`, `config`, `creds`, `ai` router, …

**Import rule:** `internal/pkg` **must not** import `internal/ait`. Dependencies flow **down** only.

When `internal/pkg` must recognize a subcommand name (e.g. meta `!ait get`):

- Wire handlers from `internal/ait` `init`, or
- Keep protocol constants in `internal/pkg` and **test** they match `internal/ait/<cmd>.Name`.

---

## `pkg/*` (repository root)

- Libraries meant for **external** import, not tied to AiT product semantics.
- Do not place `ait login` business logic here.
- Prefer `internal/pkg` for module-private shared code.

---

## `helpdoc`

- **`internal/pkg/helpdoc`** holds shared builders (`Body`, OpenSSH blocks, `Bind`, `WriteTransferHelp`, TTY keys).
- Per-command flag lists and long `--help` belong in the **command package** (`internal/ait/<cmd>/help.go`).
- Root `ait --help` uses `registry.RootHelpEntries()`; nested parents use their **`subregistry.Entries()`**.

---

## Package name = command name

- Directory and Go package name match the Cobra command (`login`, `shell`, `ai`, `config`, `ai/ask`, `config/set`, …).
- Do not suffix packages with `cmd` (e.g. ~~`shellcmd`~~). The file `cmd.go` is fine for Cobra wiring.
- REPL for `ait ai` lives in **`internal/ait/ai/mode`** (package `mode`), distinct from one-shot **`ai/ask`** and **`ai/agent`**.
- Top-level `session` CLI is **`internal/ait/session`**; import the session manager as `pkgsession "…/internal/pkg/session"`.
- **`ait shell` REPL builtins** (`exit`, `help`, `ll`, `login`, …) live in **`internal/ait/shell/builtin/`** (package `builtin`); parent `shell` wires REPL I/O and non-builtin dispatch (`exec_dispatch.go`).

---

## Nested subcommands (`ait ai`, `ait config`, …)

Subcommands of a parent command live **under the parent’s package tree**, with the same closure as top-level commands:

| Parent | Registry | Child packages (examples) |
|--------|----------|---------------------------|
| `ait` | `internal/ait/registry` | `login`, `get`, `ai`, `config`, `session`, … |
| `ait ai` | `internal/ait/ai/subregistry` | `ask`, `agent`, `provider`, `ollama` |
| `ait ai provider` | `internal/ait/ai/provider/subregistry` | `enable`, `disable` |
| `ait ai ollama` | `internal/ait/ai/ollama/subregistry` | `status`, `setup`, `start`, `stop`, `disable` |
| `ait config` | `internal/ait/config/subregistry` | `set`, `get`, `init`, `view`, `edit`, `validate` |

Each node: `register.go` (`init` → parent `subregistry.Register`), `cmd.go` (`New`), `help.go` (`WriteHelp`), run logic. Wire children via **`registry/all/imports.go`** blank imports—**do not** grow parent `cmd.go` `AddCommand` lists.

`ait session` has **no** Cobra children; use flag-driven actions in **`internal/ait/session`** (`list.go`, `load.go`, …).

---

## Shared extraction examples

| Package | Purpose |
|---------|---------|
| `internal/pkg/ssh/openssh` | Shared connection argv flags for `login`, `get`, `put` |
| `internal/pkg/ssh/transferparse` | Path classification for get/put only if both would duplicate large blocks |

Command-specific flags (`--alias`, `--save`, `--tmux` for login; `--session` for get/put) stay in `internal/ait/<cmd>/flags.go`.

---

## Extending top-level commands (Open-Closed)

Use **`internal/ait/registry`**: each top-level command implements `registry.TopLevel` and calls `registry.Register` from `init()` in `register.go`.

- **Do not** add names to `cli/root.go` `AddCommand` lists; root help comes from `registry.RootHelpEntries()` / [`registry/root_help.go`](../internal/ait/registry/root_help.go).
- **Do** add package `internal/ait/<cmd>/` with `register.go` + command closure files.
- **Wire imports**: add `_ "github.com/os-artificer/AiT/internal/ait/<cmd>"` to `registry/all/imports.go`, or run `go generate` if registrygen is enabled.

Adding a command must not require editing other subcommand packages.

Refactors (registry, parse moves) must preserve existing behavior: full `make test`, **R1–R19**, **A20–A22** per [change-acceptance.mdc](../.cursor/rules/change-acceptance.mdc)—no weakened tests or changed user-visible errors/exit codes on untouched paths.

## Checklist (new or refactored subcommand)

- [ ] `internal/ait/<cmd>/` has Name, flags, Parse, Cobra `New`, Help, Run, **`register.go`**
- [ ] Registered with `registry.Register` in `init()`
- [ ] No edits to other subcommand packages or `cli/root.go` command list
- [ ] No “names only” package under `internal/pkg`
- [ ] `internal/pkg` does not import `internal/ait`
- [ ] `cmd/ait` stays thin `main`
- [ ] Cross-command reuse is in `internal/pkg`, not copy-pasted across `ait` packages
