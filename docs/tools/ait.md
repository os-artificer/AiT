# ait — toolbox REPL

Interactive shell for the AiT toolbox. Discovers plugins under `<install-dir>/plugins/<binary>/` and runs them by plugin name (`aissh`, `ailog`, `aiimg`, `aihost`), matching each installed binary.

## Usage

```bash
ait                 # open REPL (ait>)
ait -h              # root help
ait -V              # version
```

One-shot `ait aissh …` is **not** supported — open the REPL, then run plugin commands.

## REPL

```text
ait> help
ait> help aissh
ait> aissh user@host.example.com
ait> ailog search error
ait> aiimg photo.jpg
ait> aihost snapshot
ait> history
ait> clear
ait> exit
```

Reserved builtins: `help`, `exit`/`quit`, `version`, `history`, `clear`.

On a TTY, Tab completes the **first** command word against discovered plugin names and reserved builtins (longest common prefix, then a candidate list when still ambiguous). For example, with the four shipped tools installed, `ai` + Tab lists them; `ais` + Tab expands to `aissh`. Tab does not complete plugin subcommands or flags. Non-TTY input is unchanged.

Ctrl+C while a plugin is running stops that plugin only and returns to `ait>`. Leave the REPL with `exit` or `quit` (empty-line Ctrl+D also ends the session on EOF). At the prompt, Ctrl+C clears the current line and does not exit.

## Plugin layout

After `make build` or `make install`:

```text
bin/ait                    # or $(go env GOPATH)/bin/ait
bin/aissh
bin/plugins/aissh/plugin.json
bin/plugins/aissh/aissh → ../../aissh
…
```

Discovery prefers `plugin.json`; if missing, runs `<exe> ait-plugin-manifest` and requires `binary` to match the plugin directory name. When no `plugins/` tree exists (e.g. plain `go install` without `make install`), the four builtins are filled via `PATH` + `ait-plugin-manifest` (name/binary must match the expected builtin). On Windows, `make build`/`make install` fall back to copying binaries into `plugins/` when `ln` is unavailable.

Manifest SSOT: `internal/pkg/aitplugin/manifests/<binary>/plugin.json`.

## History

REPL input history is stored as `ait_repl_history` under the AiT data directory (`AIT_HOME` / platform AppRoot), not `repl_history`.
