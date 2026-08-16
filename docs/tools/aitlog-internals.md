# aitlog internals

## Package layering

```text
cmd/aitlog/main.go
  → internal/aitlog (package aitlog: Run, Execute, WriteRootHelp)
       → registry (wraps internal/pkg/subregistry)  ← OCP registration
       → view (default route, not registered)
       → search / sources / version (registered)
       → queryflags (shared flag bind + Query build)
  → internal/pkg/logsrc   (Cursor/Source/Resolver, native + builtin backends)
  → internal/pkg/pager    (less-like viewer)
  → internal/pkg/buildinfo (shared version formatting)
```

Import direction: `cmd` → `internal/aitlog` → `internal/pkg/*`. `logsrc` and `pager` must not import `internal/aitlog` or `cobra`.

### Registry divergence from aitimg

`aitlog` wraps generic [`internal/pkg/subregistry`](../../internal/pkg/subregistry/subregistry.go) instead of a hand-written registry like `aitimg`. That is intentional (reuse rule): aitlog has no injected `Deps`. New top-level commands: add `internal/aitlog/<cmd>/register.go` and a blank import in [`registry/all/imports.go`](../../internal/aitlog/registry/all/imports.go).

### Routing

| argv | Route |
|------|-------|
| empty / `-h` / `--help` / `help` | `WriteRootHelp` |
| registered name (`search`, `sources`, `version`) | Cobra via `registry.AttachRoot` |
| else | `view.Run` (default extract-style route) |

## logsrc

- **Cursor** `{ Next, Err, Close }` — not a channel.
- **Filter** — keyword AND, regex, `--expr` (AND/OR/NOT), level, time; single Go implementation for all backends.
- **Limit** — last N *matches* (not N scanned). Reverse-capable backends stop after N newest matches. `-f` ignores `-n`/`--head`.
- **Resolver** — injectable `LookPath` / `Notice` / `Now` for tests (`--backend builtin` → zero LookPath).
- **Chains** — `chains_linux.go` / `chains_darwin.go` / `chains_windows.go` / stub `chains_other.go`.
- **macOS `log show`** — `WindowedSource` / `WindowedSourceFor`: unfiltered last-N widens `5m→1h→6h→24h→7d`; keyword/regex/`--expr` push `--predicate` and widen only through `1h` (use `--since` for more). `journalctl` gets `-g` for keywords/regex only (not `--expr`).
- **`--expr`** — `ParseExpr` / `Expr.Eval` in `expr.go` + `expr_helpers.go`; mutual exclusion with keywords/`-e` in `queryflags`.
- Stage B binary readers (`.journal` / `.asl` / `.evtx`) are listed as `Available: false, Reason: "not implemented"`.

## pager

Interactive raw-TTY viewer with progressive vs bounded modes. Non-TTY / `--no-pager` passthrough. Raw writes go through `terminal/crlf`.

## Acceptance IDs

See change-acceptance.mdc: **LG1–LG22**, **PG1–PG9**, **FB1–FB12**, **BI1–BI4**.
