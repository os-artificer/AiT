# Changelog

**Languages:** English (this page) · [简体中文](CHANGELOG.zh-CN.md)

This file records **product and CLI** functional changes only; documentation-only updates are not included. When a release is tagged, move items from `[Unreleased]` into a dated version section.

## [Unreleased]

- **CLI rename:** Toolbox first tool is `ai-ssh` (`bin/ai-ssh`, in-session `!ai-ssh` meta); default data root `~/.local/share/AiT` with automatic migration from legacy `share/ait`; keyring service `AiT` with lazy read from legacy `ait`.
- **AI (Eino refactor):** Replace `sashabaranov/go-openai` with CloudWeGo Eino ChatModel; shared DTOs in `internal/pkg/ai/apitypes`; tiered Chat routes to Ollama via Eino; tiered Agent on Ollama and Anthropic-style providers use `specialized_agent_completion`; `ai-ssh ai agent` one-shot now wires Capture/CLI/MaxSteps like REPL.

## [v1.0.0-beta.2]

- 【Added】Shell mode supports full command-line interaction; command lookup order: builtins first, then extension commands and the PATH environment variable.
- 【Optimized】Session connection uniqueness key extended to username+host+port+alias; multiple `--alias` values on the same target host save separate sessions; login without flags updates only the empty-alias matching row.

## v1.0.0-beta.1

- 【Added】AI-enhanced SSH client: activate AI in an SSH session via `!! 'question'`, `!ai-ssh ai ask`, `!ai-ssh ai agent`, or `!ai-ssh ai`.
- 【Added】Built-in file transfer: download and upload files in an SSH session via `!ai-ssh get` / `!ai-ssh put`.
- 【Added】Switch between local and remote modes at any time within an SSH session.
