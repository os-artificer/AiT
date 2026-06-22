/**
 * Copyright 2026 AiT authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package login

import (
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	helpLoginHead = `Log in to a remote host (OpenSSH-compatible). Configuration is resolved via
the local ssh -G command; the data connection uses the built-in Go SSH client.
Requires the OpenSSH client (ssh) on PATH (Linux/macOS/Windows OpenSSH).

Usage:
  ai-ssh login [flags] destination [command [argument ...]]

Arguments:
  destination    [user@]hostname, host:port, ssh://user@host:port, or Host alias from ~/.ssh/config
  command ...    optional remote command (one-shot mode, like "ssh host cmd"; default: interactive shell)

Flag order:
  Options may appear before or after destination, but before the first remote command argument.
  After the remote command starts, tokens such as -p are passed to the command (not parsed as login flags).
  Glued forms are supported: -p22, -i~/.ssh/id_ed25519

OpenSSH-compatible flags:
`
	helpLoginAIT = `
AiT flags:
  --save          save session to SQLite (no alias required; list by session id)
  --alias name    save session under this alias (implies --save); use with ai-ssh session --load <alias>
  --tmux          use remote tmux for shell persistence (interactive only; implies --save)
  --password P    SSH password (stored in OS keyring for session unlock)
  --host H        legacy: remote host (same as positional destination; do not combine)
  --user U        legacy: same as -l
  --port N        legacy: same as -p
  --identity P    legacy: same as -i

Persistence:
  Session rows are keyed by username, host, port, and alias (empty when --alias is omitted).
  Different --alias values on the same user@host:port are separate saved sessions.
  If an empty-alias row for the connection is already in ai-ssh session --list, login updates
  that row and refreshes auth TTL even without --save or --alias. Use --save or --alias
  for first-time bookmarks.

`
	helpLoginNotSupported = `Not supported by ai-ssh login (use the system ssh client instead):
`
	helpLoginExamples = `
Examples:
  ai-ssh login myserver                              # Host alias from ~/.ssh/config
  ai-ssh login -p 22 root@host -i ~/.ssh/id_ed25519
  ai-ssh login -p22 deploy@10.0.0.5 -i ~/.ssh/id_ed25519
  ai-ssh login -F ~/.ssh/config.d/work -J bastion app.internal
  ai-ssh login -o Port=22 -o User=root legacy-host
  ai-ssh login user@host                             # interactive (new host: not saved unless --save/--alias)
  ai-ssh login --save user@host                      # save session (no alias)
  ai-ssh login user@host uptime                      # one-shot (not saved unless --save/--alias or known connection)
  ai-ssh login --tmux devhost                        # ssh config Host + tmux (auto-saved)
  ai-ssh login --alias prod --tmux user@prod         # save as alias + tmux (no --save needed)
  ai-ssh login --host 192.168.1.1 --user root        # legacy flags

Flags:
  -h, --help      help for login
`
)

// HelpFull is the canonical ai-ssh login --help text (golden; do not reword flags).
var HelpFull = helpLoginHead + pkghelpdoc.SectionOpenSSHFlags + helpLoginAIT +
	helpLoginNotSupported + pkghelpdoc.SectionOpenSSHNotSupported + helpLoginExamples

// LoginHelpFull is a deprecated alias for HelpFull (tests and golden).
var LoginHelpFull = HelpFull
