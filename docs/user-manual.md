# AiT User Manual

**Languages:** English (this page) · [简体中文](user-manual.zh-CN.md)

This document describes post-install configuration, commands, prompts, data locations, and security conventions. For architecture, see [design.md](design.md).

---

## Quick start

```bash
# AI API keys (OS keyring, or encrypted files under ~/.local/share/ait/keyring on headless Linux)
# On servers without D-Bus, set a master password: export AIT_KEYRING_FILE_PASSWORD='your-secret'
./bin/ait config set ai.deepseek.api_key <your-key>
./bin/ait config set ai.minimax.api_key <your-key>   # optional second provider
# Aliases still work: deepseek.api_key, minimax.api_key

# Login: see help for full flags
./bin/ait login --help

# Interactive login and persist session
./bin/ait login deploy@example.com --alias prod
./bin/ait login -p 22 -i ~/.ssh/id_ed25519 deploy@10.0.0.5
./bin/ait login -J bastion app.internal

# One-shot remote command (like ssh host cmd; not persisted to SQLite by default)
./bin/ait login deploy@example.com uptime

# List, restore, switch sessions
./bin/ait list
./bin/ait load prod
./bin/ait session switch staging

# File transfer (OpenSSH flags; paths and flags in any order)
./bin/ait get -p22 -i~/.ssh/id ./app.log user@deploy:/var/log/app.log
./bin/ait get deploy@10.0.0.5 /var/log/app.log ./app.log
./bin/ait get ./app.log /var/log/app.log --session prod
./bin/ait put -l root ./bin/app user@host:/usr/local/bin/app
```

Optional config template:

```bash
mkdir -p ~/.config/ait
cp configs/ait.example.yaml ~/.config/ait/config.yaml
```

---

## Command reference

| Command | Description |
|---------|-------------|
| `ait shell` | Interactive mode (`AiT> `): run `login`, `list`, `help`, etc.; `exit` to quit |
| `ait login` | Connect to host; `ait login --help` for flags; `host cmd` for one-shot execution |
| `ait list` | List saved sessions |
| `ait load <id\|alias>` | Restore session; use `--password` when auth expired |
| `ait suspend` | Use `!ait suspend` inside SSH session |
| `ait session switch [id\|alias]` | Switch session (suspend current, then load target) |
| `ait session delete <id\|alias>` | Delete one session |
| `ait session delete all` | Delete all sessions (confirmation; `--force` to skip) |
| `ait get [flags] <paths...>` | Download one file; OpenSSH flags like `login` (`-p`, `-i`, `-J`, …); `ait get --help` |
| `ait put [flags] <paths...>` | Upload one file; same flag/path rules as `get` |
| `ait config set/get <key>` | Configuration; `ai.<provider>.api_key` uses keyring |
| `ait ai [--model deepseek\|minimax]` | Local AI REPL; optional cloud provider override for this session |
| `ait ai model list` | List AI providers (ENABLED, reachability, Ollama tags) |
| `ait ai chat [message...]` | One-shot AI question; reads stdin when no args; in SSH: `!ait ai chat` / `!! chat` |
| `ait ai provider enable <provider>` | Enable provider; for `ollama`, optional install checks ([y/N]) |
| `ait ai provider disable <provider>` | Disable provider (no AI requests sent to it) |
| `ait ai ollama status` | Ollama tiered/enabled flags, API reachability, pulled models |
| `ait ai ollama setup` | Enable tiered routing and set up Ollama (install prompt [y/N]) |
| `ait ai ollama start` | Start `ollama serve` (`--wait` until API up) |
| `ait ai ollama stop` | Stop local `ollama serve` |
| `ait ai ollama disable` | Disable Ollama provider in config (`--stop` to stop serve) |

Common flags: `--tmux` (remote keep-alive), `--port`, `--identity`, `--password`, `--alias`. `session switch --force` switches while current session is still active (does not save scrollback/cwd).

---

## Prompts

| Context | Prompt |
|---------|--------|
| `ait shell` | `AiT> ` |
| SSH remote | `AiT <user>@<host> ` |
| `!ait local` | `AiT Local> ` |

---

## Security and SSH

By default, `~/.ssh/known_hosts` is enforced. For new hosts, run `ssh-keyscan -H <host> >> ~/.ssh/known_hosts` first. Set `AIT_SSH_INSECURE=1` only in development (not for production). `--password` appears in the process list; prefer OS keyring or interactive credentials in production.

---

## AI mode (cloud + optional local Ollama)

Configure at least one **cloud** provider API key (see top of this page), unless you use tiered Ollama only for simple questions. For local AI, run **`ait ai ollama setup`** — choose a models directory (`OLLAMA_MODELS`; default AiT data dir, or `~/.ollama/models`, or `--models-dir`); you will be asked `[y/N]` before installing Ollama via brew/winget/install.sh; declining still enables tiered config and prints manual steps. Use **`ait ai ollama status`** to check service state; **`ait ai ollama start`** / **`stop`** for the daemon; **`ait ai ollama disable`** (or `ait ai provider disable ollama`) to stop routing without removing tiered config. Ollama does not need a keyring entry. AI never runs shell commands without your approval.

| Where | How to enter | What runs |
|-------|----------------|-----------|
| Local (system shell or `ait shell`) | `ait ai` or `ait ai --model minimax`, or `ai` in `ait shell` | `AiT AI>` — `run <cmd>` runs on **this machine**; `ait <subcmd>` runs AiT CLI after confirm |
| Inside SSH | `!ait ai` or `!!` (optional first question) | `AiT AI user@host>` — `run <cmd>` runs on **remote host** via the live session |

In AI mode: `help`, `exit`/`quit`, `clear`, `model` (show routing including tiered), `model list` (same as `ait ai model list`), `model <deepseek|minimax>` (cloud override for this session; `model ollama` is rejected), `policy per-cmd` (confirm each `run`), `policy once` (confirm once per AI session), then type questions or `run …`. Do not use `!ait ai` as a meta subcommand (`!ait ai` line form only).

With tiered routing on, short first-turn questions use local Ollama; longer prompts, code fences, complex keywords, or multi-turn history use cloud providers. `--model` / `model <cloud>` skip Ollama for that session.

If the cloud provider fails (network, 5xx, rate limit), AiT tries other cloud providers with API keys (custom `ai.fallback` order, or automatic `deepseek` then `minimax`). If Ollama fails while tiered is on, AiT fails up to cloud the same way.

---

## In-session SSH commands

| Input | Action |
|-------|--------|
| `!ait help` / `!ait` | Print in-session commands and local CLI help |
| `!ait <subcmd> …` | Run `ait` subcommand locally; exit SSH before `login`/`load`/`ai` |
| `!ait ai [question]` | Remote AI mode (see above) |
| `!! [question]` | Same as `!ait ai` |
| `!ait local` | Local mode: run commands on local machine; SSH stays connected |
| `!ait remote` | Exit local mode; resume remote input |
| `!ait suspend` | Suspend: save scrollback and remote cwd; disconnect |
| `!ait get <remote> <local>` | Download one file over current SSH (no re-auth) |
| `!ait put <local> <remote>` | Upload one file over current SSH (no re-auth) |

---

## Data and config paths

| Platform | Data (SQLite, `scrollback/`) | Config |
|----------|------------------------------|--------|
| Linux | `~/.local/share/ait/` | `~/.config/ait/config.yaml` |
| macOS | `~/Library/Application Support/ait/` | same |
| Windows | `%LOCALAPPDATA%\ait\` | `%LOCALAPPDATA%\ait\config.yaml` |

API keys and SSH password cache use the **OS keyring**, not plaintext SQLite. When no desktop keyring is available, secrets are stored as encrypted files under `<DataDir>/keyring` (Linux default: `~/.local/share/ait/keyring`). Set `AIT_KEYRING_FILE_PASSWORD` in production headless environments.

---

## Limitations

- Without `--tmux`, remote processes are not guaranteed to survive after suspend.
- `ait get` / `ait put` support single files only (not directories).
- `ait put` creates missing **parent directories** on the remote host before uploading the file.
- Transfer errors are prefixed with `local file` or `remote file` so you can tell which side failed. A message like `file does not exist` or `no such file or directory` on **put** often means the remote path or parent directory is wrong (not that the local file is missing).
- Standalone `get`/`put`: flags and paths may appear in any order; glued port/key forms work (`-p22`, `-i~/.ssh/id`).
- In SSH, `!ait get` / `!ait put` reuse the live connection (two paths only, no flags).
- Before `ait session switch` from the local shell while still in SSH, run `!ait suspend` or use `--force` (does not save scrollback/cwd).

---

## Further reading

| Document | Description |
|----------|-------------|
| [README.md](../README.md) | Overview and installation |
| [design.md](design.md) | Architecture and behavior |
| [AGENTS.md](../AGENTS.md) | Code layout |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contributing guide |
