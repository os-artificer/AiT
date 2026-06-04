# AiT User Manual

**Languages:** English (this page) · [简体中文](user-manual.zh-CN.md)

This document describes post-install configuration, commands, prompts, data locations, and security conventions. For architecture, see [design.md](design.md).

---

## Quick start

```bash
# AI API keys (OS keyring, or encrypted files under <AppRoot>/keyring on headless Linux)
# First-time config: ./bin/ait config init
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
./bin/ait session --list
./bin/ait session --list -v    # include FULL_ID for scripts
./bin/ait session --load       # interactive pick by # from list
./bin/ait session --load prod
./bin/ait session --switch staging

# File transfer (OpenSSH flags; paths and flags in any order)
./bin/ait get -p22 -i~/.ssh/id ./app.log user@deploy:/var/log/app.log
./bin/ait get deploy@10.0.0.5 /var/log/app.log ./app.log
./bin/ait get ./app.log /var/log/app.log --session prod
./bin/ait put -l root ./bin/app user@host:/usr/local/bin/app
```

Optional config template:

```bash
./bin/ait config init
# Or manually: mkdir -p ~/.local/share/ait && cp configs/ait.example.yaml ~/.local/share/ait/config.yaml
```

---

## Command reference

| Command | Description |
|---------|-------------|
| `ait shell` | Interactive mode (`AiT> `): `login`, `list`, `help`, `clear`, etc.; `exit` to quit; TTY: UTF-8 input, **Ctrl+L** clear screen |
| `ait login` | Connect to host; `--save` or `--alias` to bookmark new hosts; re-login to a listed user@host:port updates that row; `host cmd` for one-shot |
| `ait session --list` | List saved sessions (`#` column; use `-v` for full UUID) |
| `ait session --load [<id\|alias>]` | Restore session (omit id to pick); use alias or `FULL_ID` from `-v`, not the short ID column |
| `ait session --suspend` | Use `!ait suspend` inside SSH session |
| `ait session --switch [id\|alias]` | Switch session (blocked if current session still `active`; run `!ait suspend` or use `--force` first) |
| `ait session --delete <id\|alias>` | Delete one session |
| `ait session --delete all` | Delete all sessions (confirmation; `--force` to skip) |
| `ait get [flags] <paths...>` | Download one file; OpenSSH flags like `login` (`-p`, `-i`, `-J`, …); `ait get --help` |
| `ait put [flags] <paths...>` | Upload one file; same flag/path rules as `get` |
| `ait config init [--dry-run]` | Create or merge `config.yaml` (missing keys only); `--dry-run` prints planned changes |
| `ait config view` | Print `config.yaml` to stdout (creates file if missing) |
| `ait config edit` | Edit in `$VISUAL`/`$EDITOR` (default: vi on Unix, Notepad on Windows); validates after exit |
| `ait config validate` | Validate YAML and values; prints each error with field path; exit 1 on failure |
| `ait config set/get <key>` | Configuration; `ai.<provider>.api_key` uses keyring |
| `ait ai [--model deepseek\|minimax]` | Local AI REPL (UTF-8 TTY input); optional cloud provider override for this session |
| `ait ai model list` | List AI providers (ENABLED, reachability, Ollama tags) |
| `ait ai ask [message...]` | One-shot AI question; reads stdin when no args; in SSH: `!ait ai ask` / `!! ask` |
| `ait ai agent [message...]` | One-shot agent mode with tool execution; SSH: `!ait ai agent` / `!! agent` |
| `ait ai provider enable <provider>` | Enable provider; for `ollama`: optional install checks ([y/N]); flags `--yes`, `--no-install`, `--no-pull` |
| `ait ai provider disable <provider>` | Disable provider (no AI requests sent to it) |
| `ait ai ollama status` | Ollama tiered/enabled flags, API reachability, pulled models |
| `ait ai ollama setup [--dry-run]` | Enable tiered routing and set up Ollama (install prompt [y/N]); `--dry-run` prints planned steps |
| `ait ai ollama start` | Start `ollama serve` (`--wait` until API up) |
| `ait ai ollama stop` | Stop local `ollama serve` |
| `ait ai ollama disable` | Disable Ollama provider in config (`--stop` to stop serve) |

Common flags: `--tmux` (remote keep-alive), `--port`, `--identity`, `--password`, `--alias`. `session --switch --force` switches while current session is still active (does not save scrollback/cwd).

---

## Local REPL input (UTF-8)

On **macOS, Linux, and Windows**, interactive `ait shell` and `ait ai` accept **Unicode** input (including Chinese) when the terminal uses UTF-8:

- **Linux / macOS:** `export LANG=zh_CN.UTF-8` (or `en_US.UTF-8`); Terminal.app, iTerm2, GNOME Terminal, etc.
- **Windows:** Use [Windows Terminal](https://aka.ms/terminal); enable UTF-8 in settings or run `chcp 65001` in legacy consoles.
- **IME:** Type with your system input method; after you confirm composition, characters appear in the line. AiT does not show an in-app preedit box.
- **Tab:** Inserts a tab character into the current line (for indentation).
- **TTY keys (local REPL):** Ctrl+A/E (start/end of line), Left/Right (move cursor), Up/Down (recall loaded input history, including prior sessions via `<AppRoot>/repl_history`), **Ctrl+R** (reverse incremental search), Ctrl+U/K/W (kill line/to end/word), Backspace, Ctrl+C (cancel line), Ctrl+D (exit when empty), Ctrl+L (clear screen, keep buffer). Type `help` in `ait shell` or `ait ai` for the full Keys list.
- **History:** Only commands you submit at the main `AiT>` / `AiT AI>` prompt are stored—not `[y/N]` confirm answers, session picker input, or AiT internal env setup. Use the **`history`** command (bash subset: list, `history n`, `-c`, `-d`, `-s`, `-a`, `-r`, `-w`) to inspect or manage input history. Lines persist across sessions in **`<AppRoot>/repl_history`** (see **Data directory** below). In `ait ai`, **`clear`** clears chat history only; **`history`** is input history.
- **Non-TTY:** `echo 你好 | ait ai ask` works without special locale setup.
- **Clear screen:** In `ait shell`, type `clear` or press **Ctrl+L** (TTY) to clear the display; the REPL keeps running. In `ait ai`, **Ctrl+L** clears the display only; `clear` clears conversation history.

SSH sessions (`ait login`) and `!ait local` use the remote/local shell’s line editing; local `AiT>` / `AiT AI>` editing is what this section describes.

## Prompts

| Context | Prompt |
|---------|--------|
| `ait shell` | `AiT> ` |
| `ait ai` (local) | `AiT AI(Ask)> ` or `AiT AI(Agent)> ` after `mode agent` |
| SSH remote | `AiT user@host:dir[:git-branch]> ` (cwd + optional Git branch; `session.remote_prompt.path` short\|absolute) |
| SSH nested | same on inner host after `ssh` from an AiT session | Or use `ait login -J jump,final` to skip manual nested `ssh` |
| SSH AI mode (`!ait ai` / `!!`) | `AiT AI(Ask) user@host> ` or `AiT AI(Agent) user@host> ` |
| `!ait local` | `AiT Local> ` |

---

## Security and SSH

By default, `~/.ssh/known_hosts` is enforced. For new hosts, run `ssh-keyscan -H <host> >> ~/.ssh/known_hosts` first. Set `AIT_SSH_INSECURE=1` only in development (not for production). `--password` appears in the process list; prefer OS keyring or interactive credentials in production.

---

## AI mode (cloud + optional local Ollama)

Configure at least one **cloud** provider API key (see top of this page), unless you use tiered Ollama only for simple questions. For local AI, run **`ait ai ollama setup`** — choose a models directory (`OLLAMA_MODELS`; default AiT data dir, or `~/.ollama/models`, or `--models-dir`); you will be asked `[y/N]` before installing Ollama via brew/winget/install.sh; declining still enables tiered config and prints manual steps. Use **`ait ai ollama status`** to check service state; **`ait ai ollama start`** / **`stop`** for the daemon; **`ait ai ollama disable`** (or `ait ai provider disable ollama`) to stop routing without removing tiered config. Ollama does not need a keyring entry. AI never runs shell commands without your approval.

| Where | How to enter | What runs |
|-------|----------------|-----------|
| Local (system shell or `ait shell`) | `ait ai` or `ait ai --model minimax`, or `ai` in `ait shell` | `AiT AI(Ask)>` / `AiT AI(Agent)>` — `run <cmd>` runs on **this machine**; `ait <subcmd>` runs AiT CLI after confirm |
| Inside SSH | `!ait ai` or `!!` (optional first question) | `AiT AI(Ask) user@host>` / `AiT AI(Agent) user@host>` — `run <cmd>` runs on **remote host** via the live session |

Before each `run` / Agent tool execution, AiT shows a multi-line confirm menu: **`y`** run once, **`c`** trust this command’s first word for the rest of the AI session, **`s`** trust all safe commands this session, **`n`** skip. High-risk commands offer only **`y`** / **`n`**.

In AI mode: `help`, `exit`/`quit`, `clear`, `mode ask|agent` (default ask), `approve each|session` (default each; `trust` = session), `model` / `model list` / `model <deepseek|minimax>`, then questions (ask) or agent planning, or `run …` / local `ait …`. `policy` is deprecated (use `approve`). Do not use `!ait ai` as a meta subcommand (`!ait ai` line form only).

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
| `!ait chain [on\|off]` | Toggle chain mode: PTY-pipe transfer for SSH hops; `on` = ZMODEM over PTY, `off` = prefer lrzsz then SFTP (default) |

---

## Data directory

| Platform | Default `AppRoot` (config, DB, scrollback, history, …) |
|----------|--------------------------------------------------------|
| Linux | `~/.local/share/ait/` |
| macOS | `~/Library/Application Support/ait/` |
| Windows | `%LOCALAPPDATA%\ait\` |

All AiT runtime files live under one **`AppRoot`** directory. Override with **`AIT_HOME`** or **`AIT_DATA_DIR`**. On Linux upgrade from older builds, legacy config at `$XDG_CONFIG_HOME/ait/config.yaml` (typically `~/.config/ait/config.yaml`) is migrated automatically. Run **`ait config init`** after install to create or fill in missing keys in `config.yaml`.

API keys and SSH password cache use the **OS keyring**, not plaintext SQLite. When no desktop keyring is available, secrets are stored as encrypted files under `<AppRoot>/keyring`. Set `AIT_KEYRING_FILE_PASSWORD` in production headless environments.

---

## Error hints

Many CLI errors include a **`Try:`** block with suggested next commands (see `internal/pkg/errhint`):

| Situation | Typical hint |
|-----------|--------------|
| `--load` with list **ID column** (8-char prefix) | `The list ID column is for display only` → use alias, `FULL_ID` from `-v`, or interactive `--load` |
| `session --switch` while current session still **active** | Suspend inside SSH (`!ait suspend`) or `session --switch --force` |
| Empty `ait ai ask` (no message / stdin) | `ask requires a message` → pass a message or pipe stdin |
| Unknown config key | `ait config get <key>` or see example YAML |
| `ait config validate` failure | Fix listed field paths or run `ait config edit` |
| Missing config file for `validate` | Run `ait config init` first |
| Editor not available (`edit`) | Set `EDITOR` (Windows: `set EDITOR=code --wait`) |

---

## Limitations

- Without `--tmux`, remote processes are not guaranteed to survive after suspend.
- `ait get` / `ait put` support single files only (not directories).
- `ait put` creates missing **parent directories** on the remote host before uploading the file.
- Transfer errors are prefixed with `local file` or `remote file` so you can tell which side failed. A message like `file does not exist` or `no such file or directory` on **put** often means the remote path or parent directory is wrong (not that the local file is missing).
- Standalone `get`/`put`: flags and paths may appear in any order; glued port/key forms work (`-p22`, `-i~/.ssh/id`).
- In SSH, `!ait get` / `!ait put` reuse the live connection (two paths only, no flags).
- **`!ait put` local path** is on the machine running AiT (your laptop/desktop), not the remote shell’s current directory. **`!ait get` local path** is also on the AiT client. Remote paths are relative to the **active shell** (which may be a nested hop when chain mode is on).
- In SSH, **`!ait get` / `!ait put` print phase progress and a `done:` line on stderr** (e.g. `get … → …`, `handshake…`, `copying… 50%`, `done: get …`). On a TTY, **`copying… N%` updates in place** on one line (get and put use the same format, including SFTP and chain ZMODEM). Phase and done lines use CRLF-safe output on raw TTY (Linux, macOS, Windows clients).
- Before `ait session --switch` from the local shell while still in SSH, run `!ait suspend` or use `--force` (does not save scrollback/cwd).

---

## Proxy chain mode

**ProxyJump / `-J` at login** reaches the final target through `DialChain`; in-session `!ait get`/`put` on that login **probe lrzsz first**, then **fall back to SFTP** when lrzsz is missing—the same strategy as standalone `ait get`/`put` (SFTP only). **Chain mode** applies when the **interactive shell** is deeper than the dial connection (usually after you manually `ssh` inside the session).

When you login to a bastion and then manually `ssh` to an internal host, AiT's direct connection only reaches the bastion. SFTP-based file transfer can only reach the dial target by default. Features that use the PTY pipe (`!!`, `!ait ai`, scrollback) work at any depth.

When chain mode is off and you are on the **same host as the dial target**, **`!ait get` / `!ait put` use SFTP** on the direct SSH connection (even when lrzsz is installed on the remote). **ZMODEM over PTY is used only in chain mode** when `sz`/`rz` are available. When lrzsz is missing on a direct hop, AiT **falls back to SFTP** with a notice. When you have manually SSH'd to a **deeper host**, **`!ait get` / `!ait put` automatically detect** the hostname mismatch (compared to both the resolved dial host and your original login destination alias) and enable chain mode with a one-line notice.

**`!ait chain on`** enables chain mode manually, which switches to PTY-pipe injection:

- **File transfer**: **lrzsz first** — ZMODEM over the PTY when `sz`/`rz` are available. In chain mode without lrzsz, AiT shows install instructions and, on an interactive terminal, asks **`Install lrzsz on <host>? [y/N]`** before running an OS-specific install command on the active shell (Linux package managers, Homebrew, choco/winget; may need remote `sudo`). Progress shows `handshake…` and **`copying… NN%`** (same label as SFTP; in-place on TTY). AiT client (Linux/macOS/Windows Terminal) needs no local lrzsz. Standalone `ait get/put` still uses SFTP.
- **Tmux**: Tmux operations are injected through the PTY pipe instead of opening a new SSH session.

**`!ait chain off`** returns to the default SFTP/direct-session mode.

**`!ait chain`** (no argument) shows the current chain mode status.

All AiT-injected commands are silent — they do not appear in shell history or arrow-key navigation.

```text
AiT user@bastion> ssh internal-server
…logged into internal-server via bastion…
AiT user@bastion> !ait get /var/log/app.log ./app.log
  chain mode auto-enabled: active shell is on internal-server (connected to bastion).
  File transfer will use PTY-pipe injection.
  Use !ait chain off to disable.
AiT user@bastion> !ait chain   # shows on
AiT user@bastion> !ait chain off
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
