# AiT Python E2E 测试

面向 AiT CLI（`aitssh`、`ait` 及相关插件）的黑盒端到端测试，使用 pytest 与 pexpect。补充 Go 单元测试（`make test`），并不替代它们。

## CI 与本地

| 目标 | 命令 | 运行内容 |
|--------|---------|-----------|
| **GitHub Actions** | `make test-e2e-mock` | 对仓内 `testenv-mock` 跑全量非主机相关套件（`AIT_TEST_USE_MOCK=1`）：`@smoke` + `@mock` + `@ssh` 的 login/session/get/put/meta。跳过 tmux、zmodem、destructive。无密钥时 Nested/AI 仍会跳过。**不使用真实 SSH 主机。** |
| **仅 CI smoke** | `make test-e2e-smoke` | `@smoke`，含 mock SSH/HTTP oneshot 场景。 |
| **本地全量** | `make test-e2e` | Smoke + `@ssh` / `@ai` / 可选 tmux·zmodem（当 `config.local.yaml` 已配置时）。 |

CI 在未设置 `AIT_TEST_USE_MOCK=1` 时会跳过 `@pytest.mark.ssh`。设置该 mock 环境变量后（GitHub CI 默认），`@ssh` 对着回环 mock sshd 运行。

真实主机名、端口与密钥只应放在 **gitignore** 的 `tests/config.local.yaml`（或本机 `AIT_TEST_*` 环境变量）。切勿把个人域名提交进仓库。

## 前置条件

- Go 1.25+ 与 `make build` → `bin/ait`、`bin/aitssh`、…（`make test-e2e-smoke` / `make test-e2e-mock` 也会构建 `bin/testenv-mock`）
- Python 3.10+
- PATH 上有 OpenSSH 客户端（`ssh`），用于 `ssh -G` 解析
- SSH 标记测试（本地、无 mock）：`config.local.yaml` 中需有可达主机与凭据
- `@ai` 测试：按配置提供 API key 和/或本机 Ollama；`ai_runner` fixture 以 **`ait agent`** 为 argv 前缀运行
- tmux/zmodem 测试：在配置中启用功能，并在远端主机安装对应工具

## 准备

```bash
cd /path/to/AiT
make build
cp tests/config.example.yaml tests/config.local.yaml
# 仅在本地编辑 config.local.yaml：填入 SSH 主机、密钥与功能开关
make test-e2e-smoke   # 仅 @smoke
make test-e2e-mock    # 等价于 CI 的 testenv-mock 全量套件
make test-e2e         # 填好 config.local.yaml 后的本地全量
```

可选环境变量覆盖：

| 变量 | 用途 |
|----------|---------|
| `AIT_TEST_CONFIG` | YAML 配置路径（默认：`tests/config.local.yaml`） |
| `AIT_TEST_SSH_HOST` | 覆盖 `ssh.primary.host` |
| `AIT_TEST_SSH_PORT` | 覆盖 `ssh.primary.port` |
| `AIT_TEST_SSH_INNER_HOST` | 覆盖 `ssh.inner.host` |
| `AIT_TEST_DEEPSEEK_KEY` | 覆盖 `ai.deepseek_key` |
| `AIT_KEYRING_FILE_PASSWORD` | 密钥环文件后端密码 |
| `AIT_AI_EXPERIMENTAL` | 遗留环境变量（可选）；Agent 默认由 `ait` / `ait agent` 开启 |
| `AIT_TEST_USE_MOCK` | 将仓内 `testenv-mock` 注入为 `ssh.primary`，并对它跑 `@ssh` |

## 运行

```bash
# CI 安全的 smoke
make test-e2e-smoke

# 等价于 CI 的全量 mock 套件
make test-e2e-mock

# 本地全量 E2E（跳过未配置 / 未启用功能；从不在 CI 主机上跑）
make test-e2e

# 直接 pytest
pytest tests/e2e -v -m smoke
pytest tests/e2e -v -m mock
AIT_TEST_USE_MOCK=1 pytest tests/e2e -m "not destructive and not tmux and not zmodem" -v
pytest tests/e2e -m "ssh and not destructive" -v
```

没有 `config.local.yaml` 且未设置 `AIT_TEST_USE_MOCK=1` 时，SSH/AI 集成测试会被 **跳过** 并给出提示（不会失败）。

## 覆盖矩阵

| 领域 | 测试模块 | Marker |
|------|-------------|--------|
| version / help | `test_version_help.py` | smoke |
| config | `test_config.py` | smoke |
| connect_flags / `--password` | `test_connect_flags.py` | smoke |
| shell REPL | `test_shell_repl.py` | smoke |
| session CLI | `test_session_cli.py` | smoke |
| login / prompt | `test_login_ssh.py` | ssh（CI 用 mock；否则用本地主机） |
| persistence | `test_session_persist.py` | ssh |
| get/put CLI | `test_transfer_cli.py` | ssh |
| meta in session | `test_meta_in_session.py` | ssh |
| chain login/switch | `test_meta_chain_login.py` | ssh |
| nested ssh exit | `test_nested_ssh.py` | ssh（需要 inner host；mock 上跳过） |
| local AI | `test_ai_local.py` | smoke / ai |
| remote AI | `test_ai_remote.py` | ssh, ai |
| agent REPL | `test_ai_repl_agent.py` | ai |
| provider / ollama | `test_ai_provider.py` | smoke / ai |
| tmux login | `test_tmux.py` | ssh, tmux（本地） |
| ZMODEM chain | `test_zmodem_chain.py` | ssh, zmodem（本地） |
| mock SSH oneshot / get/put / auth errors | `test_mock_ssh.py` | smoke, mock |
| mock HTTP Nominatim / Ollama / OpenAI | `test_mock_http.py` | smoke, mock |

回环 mock（无需 `config.local.yaml`）：[docs/architecture/testenv-mock.md](../docs/architecture/testenv-mock.md)。

## 隔离

每个测试在 pytest `tmp_path` 下使用独立的 `AIT_HOME`。密钥环使用带测试密码的文件后端。破坏性测试（`delete all`）标记为 `@pytest.mark.destructive`，在 `make test-e2e` / `make test-e2e-mock` 中默认排除。
