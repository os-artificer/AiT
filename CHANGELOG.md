# Changelog

This file records **product and CLI** functional changes only; documentation-only updates are not included. When a release is tagged, move items from `[Unreleased]` into a dated version section.

## [Unreleased]

- 【Added】AI is experimental and off by default; enable via `ai.experimental.enabled`, `AIT_AI_EXPERIMENTAL=1`, or `--ai-experimental`.
- 【Fixed】Default remote prompt is AiT-managed again; set `AIT_REMOTE_SYSTEM_PROMPT=1` for the native shell prompt.
- 【Fixed】After nested `ssh`, active host and chain mode update once the shell is quiet (including idle MOTD).
- 【Fixed】`-J` / ProxyJump: jump clients stay open; Host aliases resolve via `ssh -G`.
- 【Fixed】Chain-mode `!ait put` no longer hangs at `finishing…` (including zero-byte uploads).

## [v1.0.0]

- 【Optimized】Restructure the project codebase into independent modules by package name; rename the SSH client to `aitssh` to lay the groundwork for extending the toolbox with more tools.
- 【Optimized】Refactor the AI module on the Eino framework; improve fallback logic when interacting with LLMs so non-standard response formats are handled instead of breaking Agent mode.

## [v1.0.0-beta.2]

- 【Added】Shell mode supports full command-line interaction; command lookup order: builtins first, then extension commands and the PATH environment variable.
- 【Optimized】Session connection uniqueness key extended to username+host+port+alias; multiple `--alias` values on the same target host save separate sessions; login without flags updates only the empty-alias matching row.

## v1.0.0-beta.1

- 【Added】AI-enhanced SSH client: activate AI in an SSH session via `!! 'question'`, `!ait ai ask`, `!ait ai agent`, or `!ait ai`.
- 【Added】Built-in file transfer: download and upload files in an SSH session via `!ait get` / `!ait put`.
- 【Added】Switch between local and remote modes at any time within an SSH session.
