# In-repo testenv mock (SSH + HTTP)

AiT acceptance tests that need an **external** environment (sshd, OpenAI-compatible chat, Ollama `/api/tags`, Nominatim reverse geocode) use loopback mocks checked into this repository. They do **not** require a real host, API key, or personal infrastructure.

GitHub Releases for `ait> install` already have a separate mock: `internal/pkg/aitrelease/testrelease` and `cmd/aitrelease-mock`. This document covers SSH/SFTP and HTTP.

## Packages

| Path | Role |
|------|------|
| [`internal/pkg/testssh`](../internal/pkg/testssh) | Loopback SSH daemon (password + optional public key) with a small exec set and SFTP |
| [`internal/pkg/testhttp`](../internal/pkg/testhttp) | Loopback HTTP: `/v1/chat/completions`, `/api/tags`, `/reverse` |
| [`cmd/testenv-mock`](../cmd/testenv-mock) | Process wrapper: prints one JSON banner line, then waits for SIGINT/SIGTERM |
| [`tests/lib/mockenv.py`](../tests/lib/mockenv.py) | Pytest helper that starts `bin/testenv-mock` |

Defaults (tests only): user `dev`, password `devpass`, hostname `mockhost.example.com`, bind `127.0.0.1:0`. Do not put real hostnames, ports, or secrets in the tree.

## Run

Go (always on `make test`):

```bash
go test ./internal/pkg/testssh/ ./internal/pkg/testhttp/
go test ./internal/pkg/ssh/client/ ./internal/pkg/transfer/ ./internal/pkg/geocode/ ./internal/pkg/ollama/ ./internal/pkg/ai/client/
```

CLI smoke against the daemon (CI-safe; no real SSH host):

```bash
make test-e2e-smoke
# or: pytest tests/e2e -v -m mock
```

Manual daemon:

```bash
go build -o bin/testenv-mock ./cmd/testenv-mock
./bin/testenv-mock
# first stdout line is JSON {ssh, http}; Ctrl+C to stop
```

Flags: `-ssh-addr`, `-http-addr`, `-root`, `-user`, `-password`, `-reject-password`, `-http-scenario`.

HTTP scenarios: `ok`, `unauthorized`, `rate_limited`, `server_error`, `not_found`, `malformed`, `geocode_unable`, `hang`.

## SFTP paths

`pkg/sftp` treats **absolute** remote paths as OS paths (the working-directory option only prefixes relative paths). Tests use `Server.RemotePath(name)` / JSON `welcome` (absolute under the mock root). Exec `cat /welcome.txt` is mapped into that root.

## Acceptance (MS*)

Success and error paths are both required. IDs live in [change-acceptance.mdc](../.cursor/rules/change-acceptance.mdc) (**MS1–MS19**).
