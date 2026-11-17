# 仓内 testenv mock（SSH + HTTP）

依赖外部环境（sshd、OpenAI chat、Ollama `/api/tags`、Nominatim）的验收，用仓内回环 mock：不需要真实主机、API key 或个人基建。

`ait shell` 的 `install` 另有 Releases mock：[`mock/testrelease`](../../mock/testrelease)、[`mock/aitrelease-mock`](../../mock/aitrelease-mock)（含 `aitkeep` / `plugins/keep`；Windows 资源出 zip）。下文覆盖 SSH/SFTP 与 HTTP。

## 包

| 路径 | 职责 |
|------|------|
| [`mock/testssh`](../../mock/testssh) | 回环 SSH 守护进程（密码 + 可选公钥），支持 exec、交互 shell 与 SFTP |
| [`mock/testhttp`](../../mock/testhttp) | 回环 HTTP：`/v1/chat/completions`、`/api/tags`、`/reverse` |
| [`mock/testenv-mock`](../../mock/testenv-mock) | 进程包装：打印一行 JSON banner，然后等待 SIGINT/SIGTERM |
| [`tests/lib/mockenv.py`](../../tests/lib/mockenv.py) | 启动 `bin/testenv-mock` 的 Pytest 辅助 |

默认值（仅测试）：用户 `dev`、密码 `devpass`、主机名 `mockhost.example.com`、绑定 `127.0.0.1:0`。不要把真实主机名、端口或机密放进仓库树。

## 运行

Go（`make test` 始终覆盖）：

```bash
go test ./mock/testssh/ ./mock/testhttp/
go test ./internal/pkg/ssh/client/ ./internal/pkg/transfer/ ./internal/pkg/geocode/ ./internal/pkg/ollama/ ./internal/pkg/ai/client/
```

对着守护进程的 CLI smoke（CI 安全；无需真实 SSH 主机）：

```bash
make test-e2e-smoke
# GitHub CI / 全量 mock 套件（login、session、get/put、meta）：
make test-e2e-mock
# 或：AIT_TEST_USE_MOCK=1 pytest tests/e2e -v -m "not destructive and not tmux and not zmodem"
```

手动启动守护进程：

```bash
go build -o bin/testenv-mock ./mock/testenv-mock
./bin/testenv-mock
# stdout 首行为 JSON {ssh, http}；Ctrl+C 停止
```

Flags：`-ssh-addr`、`-http-addr`、`-root`、`-user`、`-password`、`-reject-password`、`-http-scenario`。

HTTP scenarios：`ok`、`unauthorized`、`rate_limited`、`server_error`、`not_found`、`malformed`、`geocode_unable`、`hang`。

## SFTP 路径

`pkg/sftp` 将**绝对**远端路径视为 OS 路径（工作目录选项只给相对路径加前缀）。测试使用 `Server.RemotePath(name)` / JSON `welcome`（位于 mock root 下的绝对路径）。Exec `cat /welcome.txt` 会映射进该 root。

## 验收（MS*）

成功与错误路径均需覆盖。ID 见 [change-acceptance.mdc](../../.cursor/rules/change-acceptance.mdc)（**MS1–MS23**）。

GitHub Actions `e2e-smoke` 运行 `make test-e2e-mock`（`AIT_TEST_USE_MOCK=1`）：pytest 将 mock 注入为 `ssh.primary`，并对它执行 `@ssh` login / session / get/put / meta。tmux、zmodem、嵌套跳与 destructive 测试保持跳过。
