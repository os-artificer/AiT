# ait — toolbox REPL

Interactive shell for the AiT toolbox. Discovers plugins under `<install-dir>/plugins/<binary>/` and runs them by plugin name (`aissh`, `ai`, `ailog`, `aiimg`, `aihost`), matching each installed binary.

## Usage

```bash
ait                 # open REPL (ait>)
ait -h              # root help
ait -V              # version
ait agent [goal...] # Agent task loop (experimental; TTY default TUI)
```

One-shot `ait aissh …` is **not** supported — open the REPL, then run plugin commands.

## REPL

```text
ait> help
ait> help aissh
ait> aissh user@host.example.com
ait> ai
ait> agent check disk with aihost
ait> ailog search error
ait> aiimg photo.jpg
ait> aihost snapshot
ait> history
ait> clear
ait> exit
```

Reserved builtins: `help`, `exit`/`quit`, `version`, `history`, `clear`, `agent`, `install`, `update`, `upgrade`.

## install / update / upgrade

From the REPL, download official binaries from the latest [GitHub Release](https://github.com/os-artificer/AiT/releases) into the directory that contains the running `ait` executable (same layout as `make install`). `AIT_HOME` does not change this install root.

| Command | Default (no args) | With plugin names |
|---------|-------------------|-------------------|
| `install` | all official plugins (`aissh`, `ai`, `ailog`, `aiimg`, `aihost`) | only the named official plugins (overwrite if present) |
| `update` | all **installed** official plugins | intersection: installed ∩ requested; skip others with a message |
| `upgrade` | `ait` + all installed official plugins | **no** extra arguments |

Each command downloads one platform archive per invocation (the full release bundle for your OS/CPU); `install aissh` still fetches the complete bundle but applies only the selected tools. Non-TTY use is supported. There is no top-level `ait install` — run these inside the REPL.

```text
ait> install
ait> install aissh ailog
ait> update
ait> update aissh
ait> upgrade
ait> help install
```

If no GitHub release exists yet, or the install directory is not writable, multi-line hints suggest `make install` or `go install github.com/os-artificer/AiT/cmd/ait@latest`.

On a TTY, Tab completes the **first** command word against discovered plugin names and reserved builtins (longest common prefix, then a candidate list when still ambiguous). For example, with the shipped tools installed, `ai` + Tab uniquely completes to `ai` (trailing space); `aih` + Tab expands to `aihost`. Tab does not complete plugin subcommands or flags. Non-TTY input is unchanged.

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

Discovery prefers `plugin.json`; if missing, runs `<exe> ait-plugin-manifest` and requires `binary` to match the plugin directory name. When no `plugins/` tree exists (e.g. plain `go install` without `make install`), official builtins are filled via `PATH` + `ait-plugin-manifest` (name/binary must match the expected builtin). On Windows, `make build`/`make install` fall back to copying binaries into `plugins/` when `ln` is unavailable.

Manifest SSOT: `internal/pkg/aitplugin/manifests/<binary>/plugin.json`.

## History

REPL input history is stored as `ait_repl_history` under the AiT data directory (`AIT_HOME` / platform AppRoot), not `repl_history`.
