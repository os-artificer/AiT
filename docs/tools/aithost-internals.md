# aithost internals

## Package layering

```text
cmd/aithost/main.go
  → internal/aithost (package aithost: Run, Execute, WriteRootHelp)
       → registry (wraps internal/pkg/subregistry)  ← OCP registration
       → top (default empty-argv route + registered)
       → cpu / mem / disk / net / proc / snapshot / sources / version
       → shared (flags facade; not registered)
  → internal/pkg/hoststats   (Collector, Section, process sort/limit)
  → internal/pkg/hostdash    (Panel + sparkline; Bubble Tea-free primitives)
  → internal/pkg/buildinfo
```

Import direction: `cmd` → `internal/aithost` → `internal/pkg/*`.  
`hoststats` and `hostdash` must not import `internal/aithost` or `cobra`.

### Registry (OCP)

Same pattern as [aitlog-internals](aitlog-internals.md): wrap `pkg/subregistry`, no Deps.

New top-level command:

1. Add `internal/aithost/<cmd>/{register,cmd,help,run?}.go` with `registry.Register` in `init`
2. Blank-import from [`registry/all/imports.go`](../../internal/aithost/registry/all/imports.go)
3. Do **not** hand-edit `AddCommand` lists or add per-name branches in `cli_root.go`
4. Command packages must **not** import sibling command packages (AH15)

### Routing

| argv | Route |
|------|-------|
| `-h` / `--help` / `help` | `WriteRootHelp` |
| `-V` / `--version` | print `version.Version()` |
| empty | `top.Run` (live dashboard) |
| registered name | Cobra via `registry.AttachRoot` |
| else | errhint + Try |

### hoststats

- `Collector.Sample` fills system CPU/Mem/Disk/Net and sorted `Processes`
- Default process sort: CPU% → Mem% → NetRate → DiskIORate; `--sort` sets primary key
- Default `--limit` 10 (cap 500)
- `Section` strategy: CPU/Mem/Disk/Net/Proc Human + JSON

### hostdash

- `Panel` interface: Title / Update / View
- `DefaultPanels()`: CPU, MEM, DISK, NET, PROC
- Sparklines are Unicode block characters (no termui)

### Acceptance IDs

See change-acceptance.mdc: **AH1–AH20**.
