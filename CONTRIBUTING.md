# Contributing

**Languages:** English (this page) · [简体中文](CONTRIBUTING.zh-CN.md)

Thank you for contributing to **AiT** (AI Terminal). This document summarizes the **development workflow** and **Go coding standards** for human contributors.

- Module path: `github.com/os-artificer/AiT`
- Requires **Go 1.25+** (see [`go.mod`](go.mod))
- Before opening a PR, run: `make fmt`, `make vet`, `make test`

The canonical source for coding rules is [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc). **When you change coding conventions, update that file and both CONTRIBUTING documents** ([`CONTRIBUTING.md`](CONTRIBUTING.md), [`CONTRIBUTING.zh-CN.md`](CONTRIBUTING.zh-CN.md)).

---

## Development setup

```bash
make build       # build to bin/ait
make test        # go test ./...
make vet         # go vet ./...
make fmt         # go fmt ./...
make test-cross  # Windows / macOS cross-compile gate
./bin/ait        # run the binary
```

Windows cross-compile example:

```bash
GOOS=windows GOARCH=amd64 go build -o bin/ait.exe ./cmd/ait
```

---

## Repository layout

AiT is a Go CLI; directories follow [golang-standards/project-layout](https://github.com/golang-standards/project-layout).

| Path | Purpose |
|------|---------|
| [`cmd/ait`](cmd/ait) | Main entry; **only** a minimal `main` |
| [`internal/ait`](internal/ait) | Application entry (`Run`, Cobra CLI, `ait shell` REPL, etc.) |
| [`internal/pkg/*`](internal/pkg) | Libraries: `store`, `session`, `ssh`, `ai`, `sync`, `config`, `creds`, … |
| [`pkg/`](pkg/) | Libraries for **external** modules to import (empty today) |
| [`configs/`](configs/) | Config templates (e.g. `ait.example.yaml`) |

Conventions:

- Prefer extending `internal/ait` and `internal/pkg/*`; avoid adding top-level directories without reason.
- `internal/` is module-private — **must not** be imported from outside this module.
- See [`AGENTS.md`](AGENTS.md) for package layout; see [`docs/design.md`](docs/design.md) for product architecture.
- If subdirectories such as `api/` or `deployments/` have a `README.md`, follow them.

---

## Go coding standards

You **must** follow these conventions when generating or refactoring Go code.

### General principle: readability first

All rules prioritize **readability**. When rules conflict, choose the more readable option.

Optimize for performance only when there is a real bottleneck and readability would suffer; add a brief comment explaining why. Do not sacrifice clarity for cleverness.

### 1. Import order

Three groups, with one blank line between groups:

1. Standard library
2. This module (`github.com/os-artificer/AiT/...`)
3. Third-party packages

After `make fmt` / `goimports`, confirm blank lines between groups remain.

```go
import (
	"context"
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/store"

	"github.com/spf13/cobra"
)
```

### 2. Declaration order within a file

Top to bottom:

1. Package comment (if any)
2. File-level `const` / `var`
3. `struct` definitions that have methods
4. Methods for that struct (see §3 for visibility ordering)

### 3. Visibility ordering

- **Package-level functions**: exported (capitalized) first, unexported after.
- **Methods**: on the same receiver, exported first, unexported after.

### 4. Package-level helpers in separate files

Functions without a receiver live in dedicated files (e.g. `*_helpers.go`, `*_parse.go`), still following §3.

**Exception**: [`cmd/ait/main.go`](cmd/ait/main.go) keeps only a minimal `main`.

Do not mix package-level helpers with struct methods in the same file; see `internal/pkg/ssh/client` `*_helpers.go` layout.

### 5. No magic literals

Magic strings, numbers, and repeated literals must become named constants:

- **Shared across files**: a package-level file (e.g. `constants.go`) or the subpackage’s agreed constants file.
- **Single file only**: a `const` / `var` block at the top of that file (before or with file-level declarations per §2).

### 6. One method-bearing struct per file

Each `.go` file contains at most **one** struct that defines methods.

Pure data structs (no methods, or only JSON/SQL/config fields) are exempt; they may live in `types.go` or sibling files in the same package.

### 7. Godoc on exports

Every exported package-level function and method needs complete [godoc](https://go.dev/doc/comment):

- What it does
- Each parameter’s meaning and constraints
- Return values
- If it returns `error`, list **possible** error conditions (types or scenarios)

```go
// Login starts an SSH session for dest.
// host is the SSH destination; alias, if non-empty, must be unique in the store.
// Returns store.ErrNotFound if alias is unknown, or a dial/auth error from ssh/client.
func Login(ctx context.Context, host, alias string) (*Session, error) { ... }
```

### 8. Line width

- No line (including comments) longer than **150** characters.
- When too long: break lines, extract variables or helpers, shorten names; avoid pointless wraps.
- `go fmt` does not enforce line width; check manually or via IDE ruler ([`.editorconfig`](.editorconfig) `max_line_length = 150`).

### 9. Nesting depth

- Control flow (`if` / `for` / `switch` / `select`) nested **at most 2 levels**; refactor at 3+.
- Deeper nesting is allowed only when necessary (protocol parsing, state machines, etc.) with a one-line comment explaining why flattening is not feasible.
- Prefer early `return` / `continue` / `break`, guard clauses, and unexported helpers in separate files (§4).

```go
// Not recommended: 3 levels of nesting
for _, s := range sessions {
	if s.Active {
		if s.Tmux {
			attach(s)
		}
	}
}

// Recommended: guard clause, at most 2 levels
for _, s := range sessions {
	if !s.Active || !s.Tmux {
		continue
	}
	attach(s)
}
```

### 10. Readability details

- When rules conflict, prefer the more readable form (see general principle).
- Performance tweaks (inlining, fewer allocations, merged loops) only on profiled or known hot paths; comment the performance reason.
- Avoid: one-liner overload, cryptic abbreviations, deep nesting without comments.

### 11. Blank lines between responsibility blocks

Inside a function, group by **responsibility phase** with **exactly one** blank line between blocks.

**Typical blocks**:

- Input validation / defaults
- Resource acquisition (open, dial, NewSession)
- Core logic
- Cleanup (defer, close)

**No extra blank line needed**:

- Tightly related 2–3 lines (e.g. `err := f(); if err != nil { return err }`)
- Consecutive assignments in the same `switch` / `case`
- Very short bodies (about ≤5 lines) with a single responsibility

`go fmt` does not insert these blank lines; check manually after edits.

```go
func Load(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("empty id")
	}

	sess, err := m.store.GetSession(id)
	if err != nil {
		return err
	}

	client, err := m.dial(sess)
	if err != nil {
		return err
	}
	defer client.Close()

	return m.attach(ctx, client, sess)
}
```

---

## Formatting

The project uses [`.editorconfig`](.editorconfig):

| Scope | Convention |
|-------|------------|
| Go, `Makefile`, `go.mod` | `indent_style = tab`, `max_line_length = 150` |
| YAML, JSON | 2-space indent |
| General | UTF-8, LF, final newline, trim trailing whitespace (except Markdown) |

Run `make fmt` after Go changes. `make fmt` does **not** enforce line width ≤150 or responsibility blank lines; verify before submit.

---

## Testing

- Place `*_test.go` next to the code under test.
- `make test` must pass.
- Prefer `:memory:` SQLite for store tests.
- Add tests only when they add meaningful coverage; avoid tests that only assert the obvious.

---

## Dependencies and security

- **Dependencies**: add only when needed; see [`go.mod`](go.mod); avoid large unrelated deps.
- **Secrets**: API keys and SSH passwords go in the OS keyring ([`internal/pkg/creds`](internal/pkg/creds)).
- **Do not**: store passwords in SQLite; log secrets; commit `.env` or credential files.
- **Development**: use insecure options such as `AIT_SSH_INSECURE=1` only in dev (see [`AGENTS.md`](AGENTS.md) Pitfalls).

---

## Pre-submit checklist

- [ ] `make fmt`
- [ ] `make vet`
- [ ] `make test` (run `make test-cross` for cross-platform changes)
- [ ] New/changed Go code follows **Go coding standards** above (line width, nesting, godoc, responsibility blank lines)
- [ ] No unrelated drive-by refactors
- [ ] No secrets, `.env`, or local credentials committed

---

## Further reading

| Document | Contents |
|----------|----------|
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | This page: workflow and Go standards |
| [`CHANGELOG.md`](CHANGELOG.md) | Product and CLI changelog |
| [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) | Code of conduct |
| [`docs/user-manual.md`](docs/user-manual.md) | User manual |
| [`AGENTS.md`](AGENTS.md) | Code layout and agent orientation |
| [`docs/design.md`](docs/design.md) | Product architecture and behavior |
| [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) | Canonical Go rules (sync with Cursor / agents) |
| [`.cursor/rules/project.mdc`](.cursor/rules/project.mdc) | Project scope and build conventions |
