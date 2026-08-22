# ait internals

## Package layering

```text
cmd/ait → internal/ait (package ait)
         → internal/pkg/aitplugin  (manifest schema + Discover)
         → internal/pkg/pluginbin  (ExecPath + Runner)
```

- `internal/ait` must not import `internal/ssh` / `ailog` / `aiimg` / `aihost` for execution (exec only via discovered paths).
- `aitplugin` / `pluginbin` must not import `internal/ait`.
- `pluginbin` does not import `aitplugin`; Discover resolves `Path`, Run uses `ExecPath(path, …)`.

## Key types

| Piece | Role |
|-------|------|
| `aitplugin.Manifest` | v1 JSON schema (`apiVersion`, `name`, `binary`, `flags` for help only) |
| `aitplugin.Discover` | `sync.Once` scan of `InstallDir/plugins/<id>/` + PATH builtin fill (caller ctx ignored; internal manifest timeout) |
| `pluginbin.ExecPath` | run resolved path; optional Runner for TTY Restore→exec→MakeRaw |
| `ait.ExecuteLine` | one REPL line; swallows bare `*exec.ExitError` (AT14); Join/wrap not swallowed |
| `ait.runPlugin` | wraps `ExecPath` with Notify+drain of `os.Interrupt` so Ctrl+C stops the child only (AT29); do not use `signal.Ignore` (inherited as SIG_IGN) |
| `ait.completeREPLLine` | TTY Tab Completer (`Options.Completer`): first-token match against `Discover` names ∪ `ReservedCommandNames` |

## Manifest hook

Each shipped tool’s `ExecuteWriters` handles `ait-plugin-manifest` **before** default/empty/help/connect routing and writes embedded JSON via `aitplugin.WriteManifest`.
