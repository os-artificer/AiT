# AiT Python E2E Tests

Black-box end-to-end tests for the AiT CLIs (`aissh`, `ai`, `ait`) using pytest and pexpect. Complements Go unit tests (`make test`); does not replace them.

## CI vs local

| Target | Command | What runs |
|--------|---------|-----------|
| **GitHub Actions** | `make test-e2e-mock` | Full non-host-specific suite against in-repo `testenv-mock` (`AIT_TEST_USE_MOCK=1`): `@smoke` + `@mock` + `@ssh` login/session/get/put/meta. Skips tmux, zmodem, destructive. Nested/AI still skip without keys. **No real SSH hosts.** |
| **CI smoke only** | `make test-e2e-smoke` | `@smoke` including mock SSH/HTTP oneshot scenarios. |
| **Local full** | `make test-e2e` | Smoke + `@ssh` / `@ai` / optional tmux·zmodem when configured in `config.local.yaml`. |

On CI without `AIT_TEST_USE_MOCK=1`, `@pytest.mark.ssh` tests are skipped. With the mock env var (GitHub CI default), `@ssh` runs against loopback mock sshd.

Real hostnames, ports, and keys belong only in **gitignored** `tests/config.local.yaml` (or `AIT_TEST_*` env on your machine). Never commit personal domains into the repo.

## Prerequisites

- Go 1.25+ and `make build` → `bin/aissh`, `bin/ai`, `bin/ait` (`make test-e2e-smoke` / `make test-e2e-mock` also build `bin/testenv-mock`)
- Python 3.10+
- OpenSSH client (`ssh`) on PATH (for `ssh -G` resolution)
- For SSH-marked tests (local, no mock): reachable host and credentials in `config.local.yaml`
- For AI-marked tests: API keys and/or local Ollama per config; experimental AI via `AIT_AI_EXPERIMENTAL=1` (fixture `ai_runner` targets `bin/ai` and sets this)
- For tmux/zmodem tests: enable features in config and install tools on remote hosts

## Setup

```bash
cd /path/to/AiT
make build
cp tests/config.example.yaml tests/config.local.yaml
# Edit config.local.yaml with your SSH host, keys, and feature flags (local only)
make test-e2e-smoke   # @smoke only
make test-e2e-mock    # CI-equivalent full suite against testenv-mock
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
| `AIT_TEST_USE_MOCK` | Inject in-repo `testenv-mock` as `ssh.primary` and run `@ssh` against it |

## Run

```bash
# CI-safe smoke
make test-e2e-smoke

# CI-equivalent full mock-backed suite
make test-e2e-mock

# Full local E2E (skips unconfigured / disabled features; never on CI hosts)
make test-e2e

# Direct pytest
pytest tests/e2e -v -m smoke
pytest tests/e2e -v -m mock
AIT_TEST_USE_MOCK=1 pytest tests/e2e -m "not destructive and not tmux and not zmodem" -v
pytest tests/e2e -m "ssh and not destructive" -v
```

Without `config.local.yaml` and without `AIT_TEST_USE_MOCK=1`, SSH/AI integration tests are **skipped** with a message (not failed).

## Coverage matrix

| Area | Test module | Marker |
|------|-------------|--------|
| version / help | `test_version_help.py` | smoke |
| config | `test_config.py` | smoke |
| connect_flags / `--password` | `test_connect_flags.py` | smoke |
| shell REPL | `test_shell_repl.py` | smoke |
| session CLI | `test_session_cli.py` | smoke |
| login / prompt | `test_login_ssh.py` | ssh (mock on CI; local host otherwise) |
| persistence | `test_session_persist.py` | ssh |
| get/put CLI | `test_transfer_cli.py` | ssh |
| meta in session | `test_meta_in_session.py` | ssh |
| chain login/switch | `test_meta_chain_login.py` | ssh |
| nested ssh exit | `test_nested_ssh.py` | ssh (needs inner host; skipped on mock) |
| local AI | `test_ai_local.py` | smoke / ai |
| remote AI | `test_ai_remote.py` | ssh, ai |
| agent REPL | `test_ai_repl_agent.py` | ai |
| provider / ollama | `test_ai_provider.py` | smoke / ai |
| tmux login | `test_tmux.py` | ssh, tmux (local) |
| ZMODEM chain | `test_zmodem_chain.py` | ssh, zmodem (local) |
| mock SSH oneshot / get/put / auth errors | `test_mock_ssh.py` | smoke, mock |
| mock HTTP Nominatim / Ollama / OpenAI | `test_mock_http.py` | smoke, mock |

Loopback mocks (no `config.local.yaml`): [docs/testenv-mock.md](../docs/testenv-mock.md).

## Isolation

Each test uses an isolated `AIT_HOME` under pytest `tmp_path`. Keyring uses file backend with a test password. Destructive tests (`delete all`) are marked `@pytest.mark.destructive` and excluded by default in `make test-e2e` / `make test-e2e-mock`.
