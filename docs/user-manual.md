# AiT User Manual

**Languages:** English (this page) · [简体中文](user-manual.zh-CN.md)

This document describes post-install configuration, commands, prompts, data locations, and security conventions. For architecture, see [design.md](design.md).

---

## Quick start

```bash
# AI API keys (OS keyring, or encrypted files under <AppRoot>/keyring on headless Linux)
# First-time config: ./bin/ai-ssh config init
# On servers without D-Bus, set a master password: export AIT_KEYRING_FILE_PASSWORD='your-secret'
./bin/ai-ssh config set ai.deepseek.api_key <your-key>
./bin/ai-ssh config set ai.minimax.api_key <your-key>   # optional second provider
# Aliases still work: deepseek.api_key, minimax.api_key

# Login: see help for full flags
./bin/ai-ssh login --help

# Interactive login and persist session
./bin/ai-ssh login deploy@example.com --alias prod
./bin/ai-ssh login -p 22 -i ~/.ssh/id_ed25519 deploy@10.0.0.5
./bin/ai-ssh login -J bastion app.internal

# One-shot remote command (like ssh host cmd; not persisted to SQLite by default)
./bin/ai-ssh login deploy@example.com uptime

# List, restore, switch sessions
./bin/ai-ssh session --list
./bin/ai-ssh session --list -v    # include FULL_ID for scripts
./bin/ai-ssh session --load       # interactive pick by # from list
./bin/ai-ssh session --load prod
./bin/ai-ssh session --switch staging

# File transfer (OpenSSH flags; paths and flags in any order)
./bin/ai-ssh get -p22 -i~/.ssh/id ./app.log user@deploy:/var/log/app.log
./bin/ai-ssh get deploy@10.0.0.5 /var/log/app.log ./app.log
./bin/ai-ssh get ./app.log /var/log/app.log --session prod
./bin/ai-ssh put -l root ./bin/app user@host:/usr/local/bin/app
```

Optional config template:

```bash
./bin/ai-ssh config init
# Or manually: mkdir -p ~/.local/share/AiT && cp configs/ait.example.yaml ~/.local/share/AiT/config.yaml
```

---

## Command reference

| Command | Description |
|---------|-------------|
| `ai-ssh shell` | Interactive mode (`AiT> `): builtins (`ll` = local `ls -la` listing, shell only), then `<AppRoot>/bin`, ai-ssh CLI, PATH; `exit` to quit; TTY: UTF-8, **Ctrl+L** clear screen |
| `ai-ssh login` | Connect to host; `--save` or `--alias` to bookmark; multiple `--alias` names on the same `user@host:port` are separate sessions; plain re-login updates only an **empty-alias** match; `host cmd` for one-shot |
| `ai-ssh session --list` | List saved sessions (`#` column; use `-v` for full UUID) |
| `ai-ssh session --load [<id\|alias>]` | Restore session (omit id to pick); use alias or `FULL_ID` from `-v`, not the short ID column |
| `ai-ssh session --suspend` | Use `!ai-ssh suspend` inside SSH session |
| `ai-ssh session --switch [id\|alias]` | Switch session (blocked if current session still `active`; run `!ai-ssh suspend` or use `--force` first) |
| `ai-ssh session --delete <id\|alias>` | Delete one session |
| `ai-ssh session --delete all` | Delete all sessions (confirmation; `--force` to skip) |
| `ai-ssh get [flags] <paths...>` | Download one file; OpenSSH flags like `login` (`-p`, `-i`, `-J`, …); `ai-ssh get --help` |
| `ai-ssh put [flags] <paths...>` | Upload one file; same flag/path rules as `get` |
| `ai-ssh config init [--dry-run]` | Create or merge `config.yaml` (missing keys only); `--dry-run` prints planned changes |
| `ai-ssh config view` | Print `config.yaml` to stdout (creates file if missing) |
| `ai-ssh config edit` | Edit in `$VISUAL`/`$EDITOR` (default: vi on Unix, Notepad on Windows); validates after exit |
| `ai-ssh config validate` | Validate YAML and values; prints each error with field path; exit 1 on failure |
| `ai-ssh config set/get <key>` | Configuration; `ai.<provider>.api_key` uses keyring |
| `ai-ssh ai [--model deepseek\|minimax]` | Local AI REPL (UTF-8 TTY input); optional cloud provider override for this session |
| `ai-ssh ai model list` | List AI providers (ENABLED, reachability, Ollama tags) |
| `ai-ssh ai ask [message...]` | One-shot AI question; reads stdin when no args; in SSH: `!ai-ssh ai ask` / `!! ask` |
| `ai-ssh ai agent [message...]` | One-shot agent when a message is given; in SSH, bare `!ai-ssh ai agent` / `!! agent` enters Agent REPL (same as `!ai-ssh ai` / `!!`) |
| `ai-ssh ai provider enable <provider>` | Enable provider; for `ollama`: optional install checks ([y/N]); flags `--yes`, `--no-install`, `--no-pull` |
| `ai-ssh ai provider disable <provider>` | Disable provider (no AI requests sent to it) |
| `ai-ssh ai ollama status` | Ollama tiered/enabled flags, API reachability, pulled models |
| `ai-ssh ai ollama setup [--dry-run]` | Enable tiered routing and set up Ollama (install prompt [y/N]); `--dry-run` prints planned steps |
| `ai-ssh ai ollama start` | Start `ollama serve` (`--wait` until API up) |
| `ai-ssh ai ollama stop` | Stop local `ollama serve` |
| `ai-ssh ai ollama disable` | Disable Ollama provider in config (`--stop` to stop serve) |

Common flags: `--tmux` (remote keep-alive), `--port`, `--identity`, `--password`, `--alias`. `session --switch --force` switches while current session is still active (does not save scrollback/cwd).

---

## Local REPL input (UTF-8)

On **macOS, Linux, and Windows**, interactive `ai-ssh shell` and `ai-ssh ai` accept **Unicode** input (including Chinese) when the terminal uses UTF-8:

- **Linux / macOS:** `export LANG=zh_CN.UTF-8` (or `en_US.UTF-8`); Terminal.app, iTerm2, GNOME Terminal, etc.
- **Windows:** Use [Windows Terminal](https://aka.ms/terminal); enable UTF-8 in settings or run `chcp 65001` in legacy consoles.
- **IME:** Type with your system input method; after you confirm composition, characters appear in the line. AiT does not show an in-app preedit box.
- **Tab:** Inserts a tab character into the current line (for indentation).
- **TTY keys (local REPL):** Ctrl+A/E (start/end of line), Left/Right (move cursor), Up/Down (recall loaded input history, including prior sessions via `<AppRoot>/repl_history`), **Ctrl+R** (reverse incremental search), Ctrl+U/K/W (kill line/to end/word), Backspace, Ctrl+C (cancel line), Ctrl+D (exit when empty), Ctrl+L (clear screen, keep buffer). Type `help` in `ai-ssh shell` or `ai-ssh ai` for the full Keys list.
- **History:** Only commands you submit at the main `AiT>` / `AiT AI(Agent)>` prompt are stored—not `[y/N]` confirm answers, session picker input, or AiT internal env setup. Use the **`history`** command (bash subset: list, `history n`, `-c`, `-d`, `-s`, `-a`, `-r`, `-w`) to inspect or manage input history. Lines persist across sessions in **`<AppRoot>/repl_history`** (see **Data directory** below). In `ai-ssh ai`, **`clear`** clears chat history only; **`history`** is input history.
- **Non-TTY:** `echo 你好 | ai-ssh ai ask` works without special locale setup.
- **Clear screen:** In `ai-ssh shell`, type `clear` or press **Ctrl+L** (TTY) to clear the display; the REPL keeps running. In `ai-ssh ai`, **Ctrl+L** clears the display only; `clear` clears conversation history.

SSH sessions (`ai-ssh login`) and `!ai-ssh local` use the remote/local shell’s line editing; local `AiT>` / `AiT AI(Agent)>` editing is what this section describes.

## Prompts

| Context | Prompt |
|---------|--------|
| `ai-ssh shell` | `AiT> ` |
| `ai-ssh ai` (local) | `AiT AI(Agent)> ` by default; `mode ask` for `AiT AI(Ask)> ` |
| SSH remote | `AiT user@host:dir[:git-branch]> ` (cwd + optional Git branch; `session.remote_prompt.path` short\|absolute) |
| SSH nested | same on inner host after `ssh` from an AiT session | Or use `ai-ssh login -J jump,final` to skip manual nested `ssh` |
| SSH AI mode (`!ai-ssh ai` / `!!`) | `AiT AI(Ask) user@host> ` or `AiT AI(Agent) user@host> ` |
| `!ai-ssh local` | `AiT Local> ` |

After **`ai-ssh login`** or **`ai-ssh session --load`**, the remote **`AiT user@host:…>`** prompt should appear immediately—no extra Enter key. **`session --load` / `--switch`** does not redraw prior command output on screen; saved scrollback is kept in memory for suspend/AI context only.

---

## Security and SSH

By default, `~/.ssh/known_hosts` is enforced. For new hosts, run `ssh-keyscan -H <host> >> ~/.ssh/known_hosts` first. Set `AIT_SSH_INSECURE=1` only in development (not for production). `--password` appears in the process list; prefer OS keyring or interactive credentials in production.

---

## AI mode (cloud + optional local Ollama)

Configure at least one **cloud** provider API key (see top of this page), unless you use tiered Ollama only for simple questions. For local AI, run **`ai-ssh ai ollama setup`** — choose a models directory (`OLLAMA_MODELS`; default AiT data dir, or `~/.ollama/models`, or `--models-dir`); you will be asked `[y/N]` before installing Ollama via brew/winget/install.sh; declining still enables tiered config and prints manual steps. Use **`ai-ssh ai ollama status`** to check service state; **`ai-ssh ai ollama start`** / **`stop`** for the daemon; **`ai-ssh ai ollama disable`** (or `ai-ssh ai provider disable ollama`) to stop routing without removing tiered config. Ollama does not need a keyring entry. In **Agent** mode (default), read-only commands such as `ls` or `git status` run automatically; write commands and high-risk commands still require confirmation.

| Where | How to enter | What runs |
|-------|----------------|-----------|
| Local (system shell or `ai-ssh shell`) | `ai-ssh ai` or `ai-ssh ai --model minimax`, or `ai` in `ai-ssh shell` | `AiT AI(Agent)>` by default (`mode ask` → `AiT AI(Ask)>`) — `run <cmd>` runs on **this machine**; `ai-ssh <subcmd>` runs AiT CLI after confirm |
| Inside SSH | `!ai-ssh ai` or `!!` (optional first question) | `AiT AI(Agent) user@host>` by default — `run <cmd>` runs on **remote host** via the live session |

Before each write or high-risk `run` / Agent tool execution, AiT shows a multi-line confirm menu: **`y`** run once, **`c`** trust this command’s first word for the rest of the AI session, **`s`** trust all safe commands this session, **`n`** skip. High-risk commands offer only **`y`** / **`n`**. Read-only commands in Agent mode (e.g. `ls`, `git status`) run without a prompt.

In AI mode: `help`, `exit`/`quit`, `clear`, `mode ask|agent` (default **agent**; **`mode ask`** for Q&A only), `approve each|session` (default each; `trust` = session), `model` / `model list` / `model <deepseek|minimax>`, then questions (ask) or agent planning, or `run …` / local `ai-ssh …`. `policy` is deprecated (use `approve`). Do not use `!ai-ssh ai` as a meta subcommand (`!ai-ssh ai` line form only).

With tiered routing on, short first-turn questions use local Ollama; longer prompts, code fences, complex keywords, or multi-turn history use cloud providers. `--model` / `model <cloud>` skip Ollama for that session.

If the cloud provider fails (network, 5xx, rate limit), AiT tries other cloud providers with API keys (custom `ai.fallback` order, or automatic `deepseek` then `minimax`). If Ollama fails while tiered is on, AiT fails up to cloud the same way.

---

## In-session SSH commands

| Input | Action |
|-------|--------|
| `!ai-ssh help` / `!ai-ssh` | Print in-session commands and local CLI help |
| `!ai-ssh <subcmd> …` | Run `ai-ssh` subcommand locally; `login`/`load`/`session --switch` chain from SSH (see below) |
| `!ai-ssh login [flags] dest` | Auto-save if needed, suspend current session, log in to another host |
| `!ai-ssh load [id\|alias]` | Auto-save if needed, suspend, load saved session |
| `!ai-ssh session --switch [id\|alias]` | Auto-save if needed, suspend, switch to another session |
| `!ai-ssh ai [question]` | Remote AI mode (see above) |
| `!! [question]` | Same as `!ai-ssh ai` |
| `!ai-ssh local` | Local mode: run commands on local machine; SSH stays connected |
| `!ai-ssh remote` | Exit local mode; resume remote input |
| `!ai-ssh suspend` | Auto-save if needed; suspend: save scrollback and remote cwd; disconnect |
| `!ai-ssh get <remote> [<local>]` | Download one file over current SSH (no re-auth); optional local defaults to `./basename` |
| `!ai-ssh put <local> [<remote>]` | Upload one file over current SSH (no re-auth); optional remote defaults to `./basename` |
| `!ai-ssh chain [on\|off]` | Toggle chain mode: PTY-pipe transfer for SSH hops; `on` = ZMODEM over PTY, `off` = prefer lrzsz then SFTP (default) |

---

## Data directory

| Platform | Default `AppRoot` (config, DB, scrollback, history, …) |
|----------|--------------------------------------------------------|
| Linux | `~/.local/share/AiT/` |
| macOS | `~/Library/Application Support/AiT/` |
| Windows | `%LOCALAPPDATA%\AiT\` |

All AiT runtime files live under one **`AppRoot`** directory. Override with **`AIT_HOME`** or **`AIT_DATA_DIR`**. On Linux upgrade from older builds, legacy config at `$XDG_CONFIG_HOME/ait/config.yaml` (typically `~/.config/ait/config.yaml`) is migrated automatically. Run **`ai-ssh config init`** after install to create or fill in missing keys in `config.yaml`.

API keys and SSH password cache use the **OS keyring**, not plaintext SQLite. When no desktop keyring is available, secrets are stored as encrypted files under `<AppRoot>/keyring`. Set `AIT_KEYRING_FILE_PASSWORD` in production headless environments.

---

## Error hints

Many CLI errors include a **`Try:`** block with suggested next commands (see `internal/pkg/errhint`):

| Situation | Typical hint |
|-----------|--------------|
| `--load` with list **ID column** (8-char prefix) | `The list ID column is for display only` → use alias, `FULL_ID` from `-v`, or interactive `--load` |
| `session --switch` while current session still **active** (local CLI) | Suspend inside SSH (`!ai-ssh suspend`) or `session --switch --force`; use `!ai-ssh session --switch` in SSH to auto-suspend |
| Empty `ai-ssh ai ask` (no message / stdin) | `ask requires a message` → pass a message or pipe stdin |
| Unknown config key | `ai-ssh config get <key>` or see example YAML |
| `ai-ssh config validate` failure | Fix listed field paths or run `ai-ssh config edit` |
| Missing config file for `validate` | Run `ai-ssh config init` first |
| Editor not available (`edit`) | Set `EDITOR` (Windows: `set EDITOR=code --wait`) |

---

## Limitations

- Without `--tmux`, remote processes are not guaranteed to survive after suspend.
- `ai-ssh get` / `ai-ssh put` support single files only (not directories).
- `ai-ssh put` creates missing **parent directories** on the remote host before uploading the file.
- Transfer errors are prefixed with `local file` or `remote file` so you can tell which side failed. A message like `file does not exist` or `no such file or directory` on **put** often means the remote path or parent directory is wrong (not that the local file is missing).
- Standalone `get`/`put`: flags and paths may appear in any order; glued port/key forms work (`-p22`, `-i~/.ssh/id`).
- In SSH, `!ai-ssh get` / `!ai-ssh put` reuse the live connection (one or two paths, no flags); omitted destination defaults to the source basename.
- **`!ai-ssh put` local path** is on the machine running AiT (your laptop/desktop), not the remote shell’s current directory. **`!ai-ssh get` local path** is also on the AiT client. Remote paths are relative to the **active shell** (which may be a nested hop when chain mode is on).
- In SSH, **`!ai-ssh get` / `!ai-ssh put` print phase progress and a `done:` line on stderr** (e.g. `get … → …`, `handshake…`, `copying… 50%`, then **`closing…`** for SFTP or **`finishing…`** for chain ZMODEM, then `done: get …`). On a TTY, **`copying… N%` updates in place** on one line (get and put use the same format, including SFTP and chain ZMODEM). Phase and done lines use CRLF-safe output on raw TTY (Linux, macOS, Windows clients).
- Before `ai-ssh session --switch` from the local shell while still in SSH, run `!ai-ssh suspend` or use `--force` (does not save scrollback/cwd).

---

## Proxy chain mode

**ProxyJump / `-J` at login** reaches the final target through `DialChain`; in-session `!ai-ssh get`/`put` on that login **probe lrzsz first**, then **fall back to SFTP** when lrzsz is missing—the same strategy as standalone `ai-ssh get`/`put` (SFTP only). **Chain mode** applies when the **interactive shell** is deeper than the dial connection (usually after you manually `ssh` inside the session).

When you login to a bastion and then manually `ssh` to an internal host, AiT's direct connection only reaches the bastion. SFTP-based file transfer can only reach the dial target by default. Features that use the PTY pipe (`!!`, `!ai-ssh ai`, scrollback) work at any depth.

When chain mode is off and you are on the **same host as the dial target**, **`!ai-ssh get` / `!ai-ssh put` use SFTP** on the direct SSH connection (even when lrzsz is installed on the remote). **ZMODEM over PTY is used only in chain mode** when `sz`/`rz` are available. When lrzsz is missing on a direct hop, AiT **falls back to SFTP** with a notice. When you have manually SSH'd to a **deeper host**, **`!ai-ssh get` / `!ai-ssh put` automatically detect** the hostname mismatch (compared to both the resolved dial host and your original login destination alias) and enable chain mode with a one-line notice.

**`!ai-ssh chain on`** enables chain mode manually, which switches to PTY-pipe injection:

- **File transfer**: **lrzsz first** — ZMODEM over the PTY when `sz`/`rz` are available. In chain mode without lrzsz, AiT shows install instructions and, on an interactive terminal, asks **`Install lrzsz on <host>? [y/N]`** before running an OS-specific install command on the active shell (Linux package managers, Homebrew, choco/winget; may need remote `sudo`). Progress shows `handshake…`, **`copying… NN%`** (same label as SFTP; in-place on TTY), then **`finishing…`** before `done:` (get and put symmetric). AiT client (Linux/macOS/Windows Terminal) needs no local lrzsz. Standalone `ai-ssh get/put` still uses SFTP (100% → **`closing…`** → `done:`).
- **Tmux**: Tmux operations are injected through the PTY pipe instead of opening a new SSH session.

**`!ai-ssh chain off`** returns to the default SFTP/direct-session mode.

**`!ai-ssh chain`** (no argument) shows the current chain mode status.

All AiT-injected commands are silent — they do not appear in shell history or arrow-key navigation.

```text
AiT user@bastion> ssh internal-server
…logged into internal-server via bastion…
AiT user@bastion> !ai-ssh get /var/log/app.log ./app.log
  chain mode auto-enabled: active shell is on internal-server (connected to bastion).
  File transfer will use PTY-pipe injection.
  Use !ai-ssh chain off to disable.
AiT user@bastion> !ai-ssh chain   # shows on
AiT user@bastion> !ai-ssh chain off
  chain mode disabled
  File transfer and tmux will use SFTP/direct sessions.
```

**Chain mode requirements on the active shell host**: **lrzsz** (`sz` and `rz`) for ZMODEM transfer. If missing in chain mode, install manually or approve the `[y/N]` auto-install prompt. Windows OpenSSH servers have limited auto-install support in v1.

---

## Further reading

| Document | Description |
|----------|-------------|
| [README.md](../README.md) | Overview and installation |
| [design.md](design.md) | Architecture and behavior |
| [AGENTS.md](../AGENTS.md) | Code layout |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Contributing guide |
