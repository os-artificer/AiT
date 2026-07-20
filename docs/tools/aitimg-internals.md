# aitimg internals

## Package layering

```text
cmd/aitimg/main.go          → thin main; delegates to aitimg.Run
internal/aitimg             → package aitimg: Run, Execute, Cobra root, help
internal/aitimg/version     → VERSION + ldflags (not a Cobra subcommand)
internal/pkg/imgmeta        → Extract + human/JSON formatters (no network)
internal/pkg/geocode        → Nominatim HTTP reverse geocode
```

Import direction: `cmd` → `internal/aitimg` → `internal/pkg/*`.  
`internal/pkg` must not import `internal/aitimg`.

## Data flow

1. `Extract(path)` opens the file and uses `evanoberholster/imagemeta` (best-effort).
2. If GPS is present and `--no-geocode` is unset, CLI calls `geocode.Client.Reverse`.
3. Formatters print `Info` (including optional `Address` filled by the CLI).

## Acceptance (IM*)

See change-acceptance notes in the delivery checklist: IM1–IM13 cover extract, format, CLI, Nominatim mock, and multi-file partial failure.
