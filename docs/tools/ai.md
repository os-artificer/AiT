# ai

> Ops AI Agent plugin — experimental one-shot ask/agent, remote connect, and toolbox dispatch. **Interactive Agent Loop lives in `ait agent`.**

## Overview

`ai` is the Ops AI Agent in the AiT toolbox (`cmd/ai` → `internal/opsai`). It owns one-shot **`ask`** / **`agent`**, **`model list`**, provider/ollama subcommands, and **`connect`** for remote sessions. The multi-turn Agent Loop kernel is **`internal/pkg/agentcore`**; the interactive loop entry is **`ait agent`** (`internal/ait/agent`).

**`aissh` has zero AI** — there is no `aissh ai`. In an SSH session, bare `!!` still forwards to the remote shell (not AI). `!ait ai` is unknown with a Try hint. Remote Agent uses **`ai connect user@host`** or bare **`ai user@host`**.

Toolbox structure: [`../architecture.md`](../architecture.md). Implementation: [`ai-internals.md`](ai-internals.md).

## Experimental gate (off by default)

Enable with any of:

1. `aissh config set ai.experimental.enabled true` (or edit YAML)
2. `AIT_AI_EXPERIMENTAL=1`
3. `ai --ai-experimental …` (current process only; does not write config)

When disabled, `ai ask`, `ai agent`, `ai connect`, and `ait agent` are blocked with a multi-line Try hint. `ai --help`, `ai version`, and `ai --ait-plugin-manifest` still work.

## CLI surface

```text
ai [--ai-experimental]                 # help + Try ait agent (no REPL)
ai model list                          # non-interactive model listing (OA14)
ai ask [message...]                    # one-shot Ask; stdin if no args
ai agent [message...]                  # one-shot Agent (use ait agent for Loop)
ai connect [ssh-options] destination   # remote Agent loop on SSH host
ai user@host                           # same as connect (bare destination)
ai provider enable|disable <provider>
ai ollama status|setup|start|stop|disable
```

Interactive multi-turn tasks:

```text
ait agent [--plain] [goal...]          # Agent Loop (TTY default TUI; --plain for line mode)
ait> agent [goal...]                   # nested Loop from toolbox REPL
```

## Remote connect

`ai connect` (or bare destination) opens **`RunAgentSession`** and runs **`agentcore.RunLoop`** with remote Capture/`run` injection (no aissh meta scanner). Toolbox dispatch uses `aihost`, `ailog`, `aissh`, `aiimg` from PATH.

## Quick examples

```bash
./bin/ai --help
./bin/ai ask "summarize disk usage"
./bin/ai agent "list large files in /var/log"
./bin/ai model list
./bin/ai connect user@host.example.com
./bin/ait agent "check host health with aihost"
```

## See also

- [`ait.md`](ait.md) — toolbox REPL and `ait agent`
- [`aissh.md`](aissh.md) — SSH client (no AI subtree)
