# xiwh/zmodem (vendored)

Vendored copy of [github.com/xiwh/zmodem](https://github.com/xiwh/zmodem) **v1.0.4**.

The upstream repository is no longer available on GitHub (HTTP 404), so `proxy.golang.org` cannot serve the module and `go install github.com/os-artificer/AiT/cmd/ait@latest` fails when this dependency is external.

AiT uses the `zmodem` package for lrzsz-compatible PTY file transfer in chain mode (`internal/pkg/transfer/zmodem`).

Packages included:

- `zmodem/` — protocol engine
- `byteutil/` — block buffer I/O
- `collectionutil/` — slice helpers
- `myioutil/` — I/O helpers

Example CLI (`main.go`) from upstream is not included.

## AiT patches (relative to upstream v1.0.4)

- **`ZModemConsumer.OnUploadProgress`**: optional callback invoked during send (`sender.go` ZRPOS) after each 8KB chunk so chain-mode `!ait put` can report inline `copying… N%` progress.
