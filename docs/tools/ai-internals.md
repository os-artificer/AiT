# ai internals

> Package layout for the Ops AI Agent plugin (`cmd/ai`, `internal/opsai`) and the Agent Loop kernel.

## Call chain

```text
cmd/ai/main.go                 → minimal main; delegates to opsai.Run
internal/opsai                 → package opsai: Run, ExecuteWriters, registry
  cli_root.go                  → Cobra root, experimental strip, runWithConfig
  cmd.go                       → bare ai → help + Try ait agent; bare destination → connect
  ops_domain.go                → withOpsDomain (Domain=ops, Toolbox=agentops.DispatchToolbox)
  ask/                         → ai ask → aichat.RunOnce
  agent/                       → ai agent one-shot → agentcore.RunOnce
  connect/run.go               → ai connect → agentcore.RunLoop (remote Capture)
  model/, provider/, ollama/   → nested Cobra via subregistry

cmd/ait + internal/ait/agent   → ait agent interactive Loop → agentcore.RunLoop (+ agentui)

internal/pkg/agentcore         → Loop kernel (risk, approve, runner, DynamicRemote, RunLoop/RunOnce)
internal/pkg/agentops          → Ops toolbox DispatchToolbox + context snapshot
internal/pkg/agentui           → Plain / TUI emitters for the Loop
internal/pkg/aichat            → One-shot ask
internal/pkg/ai                → Provider routing / chat facade
internal/pkg/config            → ai.experimental.enabled SSOT (+ env / CLI flag in opsai)
```

## Layering

- `opsai` imports `internal/pkg/*` only (**must not** import `internal/ssh`).
- `ait/agent` imports `agentcore` / `agentops` / `agentui` / `ai` (not `opsai`).
- `agentcore` must not import product packages (`opsai`, `ait`, `internal/ssh`) — enforced by `TestAgentcore_NoProductImports` (AC-V4-1).
- `aissh` must not import `opsai` / `agentcore` (OA24).
- `internal/pkg/aimode` was **removed in v4** (`TestRepo_NoAimodeImport`, AC-V4-6).

## Interactive Loop vs one-shot

| Entry | Behavior |
|-------|----------|
| `ait agent` / `ait> agent` | Multi-turn Agent Loop (`agentcore.RunLoop`); TTY default TUI |
| Bare `ai` | Help + Try `ait agent` (no REPL) |
| `ai ask` | One-shot Ask (`aichat.RunOnce`) |
| `ai agent` | One-shot Agent (`agentcore.RunOnce`) |
| `ai connect` / bare destination | Remote Loop via `RunAgentSession` + `agentcore.RunLoop` |

Loop meta verbs are `exit` / `clear` / `history` / `model` only (no `mode ask|agent`, `approve`, or `run` REPL verbs).

## Remote connect (`ai connect` / bare destination)

`connect.LooksLikeDestination` rejects reserved tokens and subcommands so they never bare-route to SSH dial (OA17).

`RunAgentSession` runs the remote Loop without a meta scanner (OA16). Unlike interactive `aissh` login, there is no `!!` / `!ait` handling in the Agent session PTY loop.

Production connect sets static `Scope` / `Capture` on `agentcore.Config`. Optional `Config.Dynamic` (`DynamicRemote`) remains available for hot local/remote switching but is **not** wired by the shipped connect path.

## Ops domain + toolbox

`withOpsDomain` / agent entry points set `Domain = "ops"` and `Toolbox = agentops.DispatchToolbox`.

`agentops.DispatchToolbox` runs `aihost` / `ailog` / `aissh` / `aiimg` from PATH (single SSOT). Agent tool routing uses `ClassifyToolbox` (read auto-run).

## Experimental gate

Only `opsai.ExecuteWriters` (and `ait agent` paths that share the gate) strip `--ai-experimental` and set `config.SetAIExperimentalCLI(true)` (OA25 — aissh does not strip).

## Registry (OCP)

Top-level: `internal/opsai/<cmd>/register.go` + `registry/all/imports.go`. Nested: parent `subregistry` + child `register.go` (provider, ollama, model).

## Tests (OA* / AG* / AC-V4*)

| ID | Package / test |
|----|----------------|
| OA1 | `opsai_test.TestOpsAI_DefaultShowsHelp` |
| OA3 | `agentops.DispatchToolbox` (via opsai OA suite) |
| OA5–OA6 | `connect/run_test.go` (ScopeRemote + Capture) |
| OA7 | Connect teardown / DynamicRemote API (see change-acceptance) |
| OA14–OA17 | `opsai/oa_test.go`, `connect/run_test.go` |
| OA16 | `ssh/client/agent_session_test.go` |
| AG-2…6 | `agentcore/policy_test.go` |
| AC-V4-1 / AC-V4-6 | `agentcore/import_boundary_test.go` |

## Related

| Document | Description |
|----------|-------------|
| [`ai.md`](ai.md) | Operator-facing guide |
| [`ait.md`](ait.md) | Toolbox REPL and `ait agent` |
| [`aissh-internals.md`](aissh-internals.md) | SSH tool (no AI subtree) |
| [`../architecture.md`](../architecture.md) | Toolbox structure |
