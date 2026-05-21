# AiT User Manual

**Languages:** English (this page) · [简体中文](user-manual.zh-CN.md)

This document describes post-install configuration, commands, prompts, data locations, and security conventions. For architecture, see [design.md](design.md).

---

## Quick start

```bash
# Deepseek API key (stored in OS keyring, not in config files)
./bin/ait config set deepseek.api_key <your-key>

# Login: see help for full flags
./bin/ait login --help

# Interactive login and persist session
./bin/ait login deploy@example.com --alias prod
./bin/ait login -p 2222 -i ~/.ssh/id_ed25519 deploy@10.0.0.5
./bin/ait login -J bastion app.internal

# One-shot remote command (like ssh host cmd; not persisted to SQLite by default)
./bin/ait login deploy@example.com uptime

# List, restore, switch sessions
./bin/ait list
./bin/ait load prod
./bin/ait session switch staging

# File sync (reuse saved session)
./bin/ait sync ./local prod:/remote/path --session prod
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
| `ait sync <src> <dst>` | Cross-host sync (`--session` reuses connection) |
| `ait config set/get <key>` | Configuration; `deepseek.api_key` uses keyring |

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

## In-session SSH commands

| Input | Action |
|-------|--------|
| `!ait help` / `!ait` | Print in-session commands and local CLI help |
| `!ait <subcmd> …` | Run `ait` subcommand locally; exit SSH before `login`/`load` |
| `!ait ai <question>` | Invoke Deepseek (preferred) |
| `!! <question>` | Same (backward-compatible alias; space required after `!!`) |
| `!ait local` | Local mode: run commands on local machine; SSH stays connected |
| `!ait remote` | Exit local mode; resume remote input |
| `!ait suspend` | Suspend: save scrollback and remote cwd; disconnect |

---

## Data and config paths

| Platform | Data (SQLite, `scrollback/`) | Config |
|----------|------------------------------|--------|
| Linux | `~/.local/share/ait/` | `~/.config/ait/config.yaml` |
| macOS | `~/Library/Application Support/ait/` | same |
| Windows | `%LOCALAPPDATA%\ait\` | `%LOCALAPPDATA%\ait\config.yaml` |

API keys and SSH password cache use the **OS keyring**, not plaintext SQLite.

---

## Limitations

- Without `--tmux`, remote processes are not guaranteed to survive after suspend.
- `ait sync` with password auth uses built-in SFTP (not non-interactive system `rsync`/`scp`).
- Before `ait session switch` from the local shell while still in SSH, run `!ait suspend` or use `--force` (does not save scrollback/cwd).

---

## Further reading

| Document | Description |
|----------|-------------|
| [README.md](../README.md) | Overview and installation |
| [design.md](design.md) | Architecture and behavior |
| [AGENTS.md](../AGENTS.md) | Code layout |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contributing guide |
