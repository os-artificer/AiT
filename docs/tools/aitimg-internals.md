# aitimg internals

## Package layering

```text
cmd/aitimg/main.go                → thin main; delegates to aitimg.Run
internal/aitimg                   → Run, ExecuteWriters (route), WriteRootHelp
internal/aitimg/registry          → TopLevel Register / AttachRoot / IsRegistered
internal/aitimg/registry/all      → blank-imports command registrars
internal/aitimg/extract           → default route (not registry.Register); package name ≠ CLI verb
internal/aitimg/set               → registered `aitimg set` (JPEG EXIF write)
internal/aitimg/clear             → registered `aitimg clear` (JPEG EXIF clear)
internal/aitimg/version           → VERSION + ldflags + registered `aitimg version`
internal/pkg/toolname             → AIImg (and AISsh) SSOT
internal/pkg/imgmeta              → Extract + Patch/Clear (JPEG write) + formatters
internal/pkg/geocode              → Nominatim HTTP reverse geocode
```

Import direction: `cmd` → `internal/aitimg` → `extract` / `set` / `clear` / `registry` / `version` → `internal/pkg/*`.  
`internal/pkg` must not import `internal/aitimg`.  
`registry` must not import `extract`.

### Routing

1. Empty argv or single `-h` / `--help` / `help` → `WriteRootHelp`
2. `args[0]` registered (`version`, `set`, `clear`, …) → Cobra via `AttachRoot`
3. Else → `extract.Run` (flags + files), including `-V` / `--version`

`help <cmd>` is not supported; use `aitimg set --help`, `aitimg clear --help`, `aitimg version --help`.

### OCP: add a top-level command

1. Create `internal/aitimg/<cmd>/{register,cmd,help,run}.go` with `registry.Register` in `init`
2. Blank-import the package from `registry/all/imports.go`
3. Do not hand-edit `AddCommand` lists in `cli_root.go`

Command names that collide with common filenames (e.g. `version`) take precedence over the default extract path.

### EXIF write path

- `imgmeta.Patch` / `imgmeta.Clear` use `dsoprea/go-exif/v3` + `dsoprea/go-jpeg-image-structure/v2`
- JPEG only (SOI check); other formats → `ErrUnsupported`
- In-place via temp file + `os.Rename`
- Writable fields: Make, Model, Software, LensMake, LensModel, SerialNumber, DateTime, GPS, Altitude

## File layout

### `internal/aitimg`

| File / dir | Role |
|------------|------|
| `root.go` | `Run` (errhint; `errors.Is(imgmeta.ErrPartialFailure)`) |
| `cli_root.go` | `Execute` / `ExecuteWriters` / `newRootCmd` (AttachRoot only) |
| `help.go` | `WriteRootHelp` (extract sections + Commands) |
| `init_hooks.go` | `_ "…/registry/all"` |
| `registry/` | Register, AttachRoot, Entries, IsRegistered, RootHelpEntries, Deps |
| `extract/` | Default-route Cobra, `Run`, `WriteHelp`, geocode test hooks |
| `set/` / `clear/` | Registered write/clear subcommands |
| `version/` | `Version()` / `Short()` + registered subcommand |

### `internal/pkg/imgmeta`

| File | Role |
|------|------|
| `types.go` | `Info`, package comment |
| `errors.go` | `ErrEmptyPath`, `ErrUnsupported`, `ErrPartialFailure`, `ErrNothingToChange` |
| `extract.go` / `extract_helpers.go` | `Extract` + dims/format helpers |
| `patch.go` / `patch_jpeg.go` | `PatchFields`, `ClearMask`, `Patch`, `Clear` |
| `format_human.go` / `format_helpers.go` | locked summary fields |
| `format_json.go` | JSON array (+ optional nested `"exif"`) |
| `format_full.go` / helpers | `--all` Field/Value tables |

### `internal/pkg/geocode`

| File | Role |
|------|------|
| `nominatim.go` | `Client.Reverse` (Nominatim) |

## Data flow

1. `Extract(path)` opens the file and uses `evanoberholster/imagemeta` (best-effort).
2. Successful/partial decode retains `Info.Exif` (`*exif.Exif`) for `--all`.
3. If GPS is present and `--no-geocode` is unset, CLI calls `geocode.Client.Reverse`.
4. Formatters print locked `Info` fields; `--all` adds `FormatFullHuman` / nested JSON `"exif"` from the library marshal of `Info.Exif`.
5. `set`/`clear` call `imgmeta.Patch`/`Clear` then write JPEG segments back.

## Acceptance

**IM\*** / **IM-FULL\*** / **IM-TABLE\***: extract and `--all` formatting (unchanged).

**SC1–SC10**: set/clear JPEG write (`TestPatch_*`, `TestClear_*`, `TestSetCmd_NoFlags`, `TestClearCmd_NoFlags`, `TestCLI_SetAndClearMake`, `TestCLI_SetPartialFailure`, root help Commands).
