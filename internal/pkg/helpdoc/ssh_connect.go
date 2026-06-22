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

package helpdoc

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	sshConnectHead = `Connect to a remote host (OpenSSH-compatible). Configuration is resolved via
the local ssh -G command; the data connection uses the built-in Go SSH client.
Requires the OpenSSH client (ssh) on PATH (Linux/macOS/Windows OpenSSH).

Usage:
  ai-ssh [flags] destination [command [argument ...]]

Arguments:
  destination    [user@]hostname, host:port, ssh://user@host:port, or Host alias from ~/.ssh/config
  command ...    optional remote command (one-shot mode, like "ssh host cmd"; default: interactive shell)

Flag order:
  Options may appear before or after destination, but before the first remote command argument.
  After the remote command starts, tokens such as -p are passed to the command (not parsed as flags).
  Glued forms are supported: -p22, -i~/.ssh/id_ed25519

OpenSSH-compatible flags:
`
	sshConnectAIT = `
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
  If an empty-alias row for the connection is already in ai-ssh session --list, connect updates
  that row and refreshes auth TTL even without --save or --alias. Use --save or --alias
  for first-time bookmarks.

`
	sshConnectNotSupported = `Not supported by ai-ssh connect (use the system ssh client instead):
`
	sshConnectExamples = `
Examples:
  ai-ssh myserver                              # Host alias from ~/.ssh/config
  ai-ssh -p 22 root@host -i ~/.ssh/id_ed25519
  ai-ssh -p22 deploy@10.0.0.5 -i ~/.ssh/id_ed25519
  ai-ssh -F ~/.ssh/config.d/work -J bastion app.internal
  ai-ssh -o Port=22 -o User=root legacy-host
  ai-ssh user@host                             # interactive (new host: not saved unless --save/--alias)
  ai-ssh --save user@host                      # save session (no alias)
  ai-ssh user@host uptime                      # one-shot (not saved unless --save/--alias or known connection)
  ai-ssh --tmux devhost                        # ssh config Host + tmux (auto-saved)
  ai-ssh --alias prod --tmux user@prod         # save as alias + tmux (no --save needed)
  ai-ssh --host 192.168.1.1 --user root        # legacy flags
`
)

// SSHConnectHelpFull is the canonical SSH connect section for ai-ssh --help.
var SSHConnectHelpFull = sshConnectHead + SectionOpenSSHFlags + sshConnectAIT +
	sshConnectNotSupported + SectionOpenSSHNotSupported + sshConnectExamples

// WriteSSHConnectSection prints the default SSH connect usage block to w.
func WriteSSHConnectSection(w io.Writer) error {
	_, err := io.WriteString(w, SSHConnectHelpFull)
	return err
}

// WriteSSHConnectShort prints a compact flag list for shell REPL help.
func WriteSSHConnectShort(w io.Writer) error {
	const short = `  Connect flags (see "ai-ssh --help" for full detail):
    -p port  -l user  -i identity  -F config  -J jump  -o key=value
    --save  --alias name  --tmux  --password P
    --host H  --user U  --port N  --identity P  (legacy, same as ssh flags above)
`
	_, err := io.WriteString(w, short)
	return err
}

// RootUsageLine returns the top-level ai-ssh usage without subcommands.
func RootUsageLine() string {
	return toolname.AISsh + " [flags] destination [command ...]"
}
