# Changelog

**Languages:** English (this page) · [简体中文](CHANGELOG.zh-CN.md)

This file records **product and CLI** functional changes only; documentation-only updates are not included. When a release is tagged, move items from `[Unreleased]` into a dated version section.

## [Unreleased]

- **CLI rename:** Toolbox first tool is `aitssh` (`bin/aitssh`, in-session `!ait` meta); default data root `~/.local/share/ait` with automatic migration from mixed-case `AiT` when present; keyring service `ait` with lazy read from legacy `AiT` entries. If both `share/ait` and `share/AiT` exist with session data, merge manually (include `keyring/`).
- **Breaking:** Binary and CLI command renamed from `ai-ssh` to `aitssh` (`go install github.com/os-artificer/AiT/cmd/aitssh@latest`). Scripts and PATH aliases must use `aitssh`; optional transition: `ln -s aitssh ai-ssh`.
- **Breaking:** In-session meta prefix is `!ait`; `!ai-ssh` is no longer recognized. `!!` remains the AI shortcut.
- **AI (Eino refactor):** Replace `sashabaranov/go-openai` with CloudWeGo Eino ChatModel; shared DTOs in `internal/pkg/ai/apitypes`; tiered Chat routes to Ollama via Eino; tiered Agent on Ollama and Anthropic-style providers use `specialized_agent_completion`; `aitssh ai agent` one-shot now wires Capture/CLI/MaxSteps like REPL.

## [v1.0.0-beta.2]

- 【Added】Shell mode supports full command-line interaction; command lookup order: builtins first, then extension commands and the PATH environment variable.
- 【Optimized】Session connection uniqueness key extended to username+host+port+alias; multiple `--alias` values on the same target host save separate sessions; login without flags updates only the empty-alias matching row.

## v1.0.0-beta.1

- 【Added】AI-enhanced SSH client: activate AI in an SSH session via `!! 'question'`, `!ait ai ask`, `!ait ai agent`, or `!ait ai`.
- 【Added】Built-in file transfer: download and upload files in an SSH session via `!ait get` / `!ait put`.
- 【Added】Switch between local and remote modes at any time within an SSH session.
