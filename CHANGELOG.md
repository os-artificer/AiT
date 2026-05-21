# Changelog

**Languages:** English (this page) · [简体中文](CHANGELOG.zh-CN.md)

All notable **product and CLI** changes are listed here. Documentation-only updates are not included. When a release is tagged, move items from `[Unreleased]` into a dated version section.

## [Unreleased]

### Added

- AI-enhanced SSH client core: `login`, `list`, `load`, `suspend`, `session switch`, `session delete`
- SQLite session persistence, aliases, `session.auth_ttl`, and OS keyring credentials
- OpenSSH config resolution via `ssh -G` with Go `x/crypto/ssh` as the data plane
- Optional remote `--tmux` keep-alive (`ait-<session_id>`)
- In-session `!ait`: `ai` / `!!`, `local` / `remote`, `suspend`, and help
- Deepseek assistant (`!ait ai`)
- File transfer via `ait get` / `ait put` (SFTP, single file); in-session `!ait get` / `!ait put`
- Local entry `ait shell` REPL (`AiT>` prompt)

### Changed

- **Breaking:** removed `ait sync`; use `ait get` / `ait put` instead

### Improved

- CLI package layout consolidated under `internal/ait`
- Split `internal/ait` and `internal/pkg/ssh` packages for clearer responsibilities
- In-session `!ait` interaction and output behavior

### Fixed

- SSH login authentication, host key verification, and remote prompt initialization
- In-session `!ait` meta-command output issues
