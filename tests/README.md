# AiT Python E2E Tests

Black-box end-to-end tests for the `aitssh` CLI using pytest and pexpect. Complements Go unit tests (`make test`); does not replace them.

## Prerequisites

- Go 1.25+ and `make build` → `bin/aitssh`
- Python 3.10+
- OpenSSH client (`ssh`) on PATH (for `ssh -G` resolution)
- For SSH-marked tests: reachable host and credentials in `config.local.yaml`
- For AI-marked tests: API keys and/or local Ollama per config
- For tmux/zmodem tests: enable features in config and install tools on remote hosts

## Setup

```bash
cd /path/to/AiT
make build
cp tests/config.example.yaml tests/config.local.yaml
# Edit config.local.yaml with your SSH host, keys, and feature flags
make test-e2e-smoke   # creates tests/.venv and installs deps automatically
```

Or manually:

```bash
python3 -m venv tests/.venv
tests/.venv/bin/pip install -r tests/requirements.txt
cd tests && ../tests/.venv/bin/pytest e2e -v -m smoke
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

## Run

```bash
# Full E2E (skips unconfigured / disabled features)
make test-e2e

# Quick smoke only (help, non-interactive CLI)
make test-e2e-smoke

# Direct pytest
pytest tests/e2e -v
pytest tests/e2e -m smoke -v
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
| login / prompt | `test_login_ssh.py` | ssh |
| persistence | `test_session_persist.py` | ssh |
| get/put CLI | `test_transfer_cli.py` | ssh |
| meta in session | `test_meta_in_session.py` | ssh |
| chain login/switch | `test_meta_chain_login.py` | ssh |
| nested ssh exit | `test_nested_ssh.py` | ssh |
| local AI | `test_ai_local.py` | smoke / ai |
| remote AI | `test_ai_remote.py` | ssh, ai |
| agent REPL | `test_ai_repl_agent.py` | ai |
| provider / ollama | `test_ai_provider.py` | smoke / ai |
| tmux login | `test_tmux.py` | ssh, tmux |
| ZMODEM chain | `test_zmodem_chain.py` | ssh, zmodem |

## Isolation

Each test uses an isolated `AIT_HOME` under pytest `tmp_path`. Keyring uses file backend with a test password. Destructive tests (`delete all`) are marked `@pytest.mark.destructive` and excluded by default in `make test-e2e`.
