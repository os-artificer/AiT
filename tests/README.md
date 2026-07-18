# AiT Python E2E Tests

Black-box end-to-end tests for the `aitssh` CLI using pytest and pexpect. Complements Go unit tests (`make test`); does not replace them.

## CI vs local

| Target | Command | What runs |
|--------|---------|-----------|
| **GitHub Actions** | `make test-e2e-smoke` | `@smoke` only: help, config, session CLI, shell REPL, AI experimental gate. **No SSH / no real hosts.** |
| **Local full** | `make test-e2e` | Smoke + `@ssh` / `@ai` / optional tmux·zmodem when configured in `config.local.yaml`. |

On CI (`CI=true` / `GITHUB_ACTIONS`), `@pytest.mark.ssh` tests are always skipped even if a host config is present.

Real hostnames, ports, and keys belong only in **gitignored** `tests/config.local.yaml` (or `AIT_TEST_*` env on your machine). Never commit personal domains into the repo.

## Prerequisites

- Go 1.25+ and `make build` → `bin/aitssh`
- Python 3.10+
- OpenSSH client (`ssh`) on PATH (for `ssh -G` resolution)
- For SSH-marked tests (local): reachable host and credentials in `config.local.yaml`
- For AI-marked tests: API keys and/or local Ollama per config; experimental AI via `AIT_AI_EXPERIMENTAL=1` (fixture `ai_runner` sets this)
- For tmux/zmodem tests: enable features in config and install tools on remote hosts

## Setup

```bash
cd /path/to/AiT
make build
cp tests/config.example.yaml tests/config.local.yaml
# Edit config.local.yaml with your SSH host, keys, and feature flags (local only)
make test-e2e-smoke   # CI-equivalent; creates tests/.venv automatically
make test-e2e         # full local suite when config.local.yaml is filled
```

Optional environment overrides:

| Variable | Purpose |
|----------|---------|
| `AIT_TEST_CONFIG` | Path to YAML config (default: `tests/config.local.yaml`) |
| `AIT_TEST_SSH_HOST` | Override `ssh.primary.host` |
| `AIT_TEST_SSH_PORT` | Override `ssh.primary.port` |
| `AIT_TEST_SSH_INNER_HOST` | Override `ssh.inner.host` |
| `AIT_TEST_DEEPSEEK_KEY` | Override `ai.deepseek_key` |
| `AIT_KEYRING_FILE_PASSWORD` | Keyring file backend password |
| `AIT_AI_EXPERIMENTAL` | Enable experimental AI (set by `ai_runner` fixture) |

## Run

```bash
# CI-safe smoke
make test-e2e-smoke

# Full local E2E (skips unconfigured / disabled features; never on CI hosts)
make test-e2e

# Direct pytest
pytest tests/e2e -v -m smoke
pytest tests/e2e -m "ssh and not destructive" -v
```

Without `config.local.yaml`, SSH/AI integration tests are **skipped** with a message (not failed).

## Coverage matrix

| Area | Test module | Marker |
|------|-------------|--------|
| version / help | `test_version_help.py` | smoke |
| config | `test_config.py` | smoke |
| shell REPL | `test_shell_repl.py` | smoke |
| session CLI | `test_session_cli.py` | smoke |
| login / prompt | `test_login_ssh.py` | ssh (local) |
| persistence | `test_session_persist.py` | ssh (local) |
| get/put CLI | `test_transfer_cli.py` | ssh (local) |
| meta in session | `test_meta_in_session.py` | ssh (local) |
| chain login/switch | `test_meta_chain_login.py` | ssh (local) |
| nested ssh exit | `test_nested_ssh.py` | ssh (local) |
| local AI | `test_ai_local.py` | smoke / ai |
| remote AI | `test_ai_remote.py` | ssh, ai (local) |
| agent REPL | `test_ai_repl_agent.py` | ai |
| provider / ollama | `test_ai_provider.py` | smoke / ai |
| tmux login | `test_tmux.py` | ssh, tmux (local) |
| ZMODEM chain | `test_zmodem_chain.py` | ssh, zmodem (local) |

## Isolation

Each test uses an isolated `AIT_HOME` under pytest `tmp_path`. Keyring uses file backend with a test password. Destructive tests (`delete all`) are marked `@pytest.mark.destructive` and excluded by default in `make test-e2e`.
