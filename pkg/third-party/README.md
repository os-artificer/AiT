# Third-party vendored Go sources

External Go libraries that cannot be fetched reliably from the public module proxy (removed upstream, private repos, etc.) are vendored here as first-party packages under the AiT module.

## Import prefix

```text
github.com/os-artificer/AiT/pkg/third-party/<upstream>/...
```

## Adding a new vendored library

1. Copy upstream source into `pkg/third-party/<name>/` (omit example `main`, upstream `go.mod`, and tests unless needed).
2. Rewrite internal imports to the `github.com/os-artificer/AiT/pkg/third-party/...` prefix.
3. Remove the external `require` from the root [`go.mod`](../../go.mod) and run `go mod tidy`.
4. Document the upstream URL, version, and reason in a README beside the copy and in the table below.

## Index

| Path | Upstream | Version | Reason |
|------|----------|---------|--------|
| [`xiwh/zmodem/`](xiwh/zmodem/) | [github.com/xiwh/zmodem](https://github.com/xiwh/zmodem) | v1.0.4 | Upstream GitHub 404; breaks `go install` via module proxy |
