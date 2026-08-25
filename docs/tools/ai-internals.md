# ai internals

> Package layout for the Ops AI Agent plugin (`cmd/ai`, `internal/opsai`).

## Call chain

```text
cmd/ai/main.go          → minimal main; delegates to opsai.Run
internal/opsai          → package opsai: Run, ExecuteWriters, registry, ask/agent/provider/ollama/connect
  cli_root.go           → Cobra root, experimental strip, runWithConfig
  cmd.go                → RunLocal (REPL), bare-destination → connect
  session_state.go      → REPL connect/disconnect + DynamicRemote wiring
  ops_domain.go         → withOpsDomain (Domain=ops, Toolbox=DispatchToolbox)
  connect/run.go        → ai connect + buildRemoteAgentOptions (ScopeRemote)
internal/pkg/aimode     → shared AI REPL (risk, approve, agent, chat, DynamicRemote)
internal/pkg/ai         → provider routing / chat facade (Ops prompt via AssistEnv.Domain)
internal/pkg/config     → ai.experimental.enabled SSOT (+ env / CLI flag in opsai)
```

## Layering

- `opsai` imports `internal/pkg/*` only (**must not** import `internal/ssh`).
- `aissh` must not import `aimode` / `opsai` (OA24).
- Shared UX stays in `aimode`; Ops-domain prompts, toolbox, and `connect` live in `opsai`.
- `internal/ssh/ai/**` is legacy/dead; the shipped plugin is `internal/opsai`.

## Dynamic remote (REPL `connect` / `disconnect`)

`sessionRuntime` owns an optional `*sshclient.Client` and `*aimode.DynamicRemote`. On connect:

1. Dial via `resolve.Resolve` + `sshclient.DialChain` (supports `-J` / ProxyJump — OA15).
2. `StartAgentPTY` → `AgentSessionHandles` (Capture, Scrollback, Stdin).
3. `DynamicRemote.SetRemote` updates scope for prompt, `run`, Ask env, confirm labels, and Agent Capture.

On `disconnect`, `closeLocked` closes PTY/client and `DynamicRemote.SetLocal`.

All scope-sensitive paths use `resolveDynamic(opt)` (mutex-safe `Snapshot()`).

## Remote connect command (`ai connect` / bare destination)

`connect.LooksLikeDestination` rejects REPL verbs (`mode`, `run`, `disconnect`, …) and subcommands (`ask`, `agent`, …) so they never bare-route to SSH dial (OA17).

`RunAgentSession` runs the remote REPL without a meta scanner (OA16). Unlike interactive `aissh` login, there is no `!!` / `!ait` handling in the Agent session PTY loop.

## Ops domain + toolbox

`withOpsDomain` sets `Options.Domain = "ops"` and `Options.Toolbox = DispatchToolbox`. Used by:

- Local REPL (`RunLocal`)
- `ai agent` one-shot (`deps.ApplyOpsDomain`)
- Remote connect REPL (`buildRemoteAgentOptions`)

`DispatchToolbox` runs `aihost` / `ailog` / `aissh` / `aiimg` from PATH. Agent tool routing uses `ClassifyToolbox` (read auto-run).

## Experimental gate

Only `opsai.ExecuteWriters` strips `--ai-experimental` and sets `config.SetAIExperimentalCLI(true)` (OA25 — aissh does not strip).

## Registry (OCP)

Top-level: `internal/opsai/<cmd>/register.go` + `registry/all/imports.go`. Nested: parent `subregistry` + child `register.go` (provider, ollama).

## Tests (OA*)

| ID | Package / test |
|----|----------------|
| OA1 | `opsai_test.TestOpsAI_DefaultEntersREPL` |
| OA5–OA6 | `connect/run_test.go` (ScopeRemote + Capture) |
| OA7 | `opsai/session_state_test.go` |
| OA14–OA17 | `opsai/oa_test.go`, `connect/run_test.go` |
| OA16 | `ssh/client/agent_session_test.go` |

## Related

| Document | Description |
|----------|-------------|
| [`ai.md`](ai.md) | Operator-facing guide |
| [`aissh-internals.md`](aissh-internals.md) | SSH tool (no AI subtree) |
| [`../architecture.md`](../architecture.md) | Toolbox structure |
