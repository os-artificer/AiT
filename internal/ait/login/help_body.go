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
  ait login [flags] destination [command [argument ...]]

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
  --alias name    save session under this alias (implies --save); use with ait session --load <alias>
  --tmux          use remote tmux for shell persistence (interactive login only)
  --password P    SSH password (stored in OS keyring for session unlock)
  --host H        legacy: remote host (same as positional destination; do not combine)
  --user U        legacy: same as -l
  --port N        legacy: same as -p
  --identity P    legacy: same as -i

Persistence:
  If user@host:port is already in ait session --list, login updates that row and
  refreshes auth TTL even without --save or --alias. Use --save or --alias for first-time bookmarks.

`
	helpLoginNotSupported = `Not supported by ait login (use the system ssh client instead):
`
	helpLoginExamples = `
Examples:
  ait login myserver                              # Host alias from ~/.ssh/config
  ait login -p 22 root@host -i ~/.ssh/id_ed25519
  ait login -p22 deploy@10.0.0.5 -i ~/.ssh/id_ed25519
  ait login -F ~/.ssh/config.d/work -J bastion app.internal
  ait login -o Port=22 -o User=root legacy-host
  ait login user@host                             # interactive (new host: not saved unless --save/--alias)
  ait login --save user@host                      # save session (no alias)
  ait login user@host uptime                      # one-shot (not saved unless --save/--alias or known connection)
  ait login --alias prod --tmux user@prod         # save as alias + tmux (no --save needed)
  ait login --host 192.168.1.1 --user root        # legacy flags

Flags:
  -h, --help      help for login
`
)

// HelpFull is the canonical ait login --help text (golden; do not reword flags).
var HelpFull = helpLoginHead + pkghelpdoc.SectionOpenSSHFlags + helpLoginAIT +
	helpLoginNotSupported + pkghelpdoc.SectionOpenSSHNotSupported + helpLoginExamples

// LoginHelpFull is a deprecated alias for HelpFull (tests and golden).
var LoginHelpFull = HelpFull
