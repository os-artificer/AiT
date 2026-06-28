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
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func sshConnectHeadBlock() string {
	return fmt.Sprintf(`Connect to a remote host (OpenSSH-compatible). Configuration is resolved via
the local ssh -G command; the data connection uses the built-in Go SSH client.
Requires the OpenSSH client (ssh) on PATH (Linux/macOS/Windows OpenSSH).

Usage:
  %s [flags] destination [command [argument ...]]

Arguments:
  destination    [user@]hostname, host:port, ssh://user@host:port, or Host alias from ~/.ssh/config
  command ...    optional remote command (one-shot mode, like "ssh host cmd"; default: interactive shell)

Flag order:
  Options may appear before or after destination, but before the first remote command argument.
  After the remote command starts, tokens such as -p are passed to the command (not parsed as flags).
  Glued forms are supported: -p22, -i~/.ssh/id_ed25519

OpenSSH-compatible flags:
`, toolname.AISsh)
}

func sshConnectAITBlock() string {
	return fmt.Sprintf(`
AiT flags:
  --save          save session to SQLite (no alias required; list by session id)
  --alias name    save session under this alias (implies --save); use with %s
  --tmux          use remote tmux for shell persistence (interactive only; implies --save)
  --password P    SSH password (stored in OS keyring for session unlock)
  --host H        legacy: remote host (same as positional destination; do not combine)
  --user U        legacy: same as -l
  --port N        legacy: same as -p
  --identity P    legacy: same as -i

Persistence:
  Session rows are keyed by username, host, port, and alias (empty when --alias is omitted).
  Different --alias values on the same user@host:port are separate saved sessions.
  If an empty-alias row for the connection is already in %s, connect updates
  that row and refreshes auth TTL even without --save or --alias. Use --save or --alias
  for first-time bookmarks.

`, toolname.CLI("session --load <alias>"), toolname.CLI("session --list"))
}

func sshConnectNotSupportedBlock() string {
	return fmt.Sprintf("Not supported by %s connect (use the system ssh client instead):\n",
		toolname.AISsh)
}

func sshConnectExamplesBlock() string {
	lines := []string{
		"",
		"Examples:",
		"  " + toolname.CLI("myserver") + "                              # Host alias from ~/.ssh/config",
		"  " + toolname.CLI("-p 22 root@host -i ~/.ssh/id_ed25519"),
		"  " + toolname.CLI("-p22 deploy@10.0.0.5 -i ~/.ssh/id_ed25519"),
		"  " + toolname.CLI("-F ~/.ssh/config.d/work -J bastion app.internal"),
		"  " + toolname.CLI("-o Port=22 -o User=root legacy-host"),
		"  " + toolname.CLI("user@host") + "                             # interactive (new host: not saved unless --save/--alias)",
		"  " + toolname.CLI("--save user@host") + "                      # save session (no alias)",
		"  " + toolname.CLI("user@host uptime") + "                      # one-shot (not saved unless --save/--alias or known connection)",
		"  " + toolname.CLI("--tmux devhost") + "                        # ssh config Host + tmux (auto-saved)",
		"  " + toolname.CLI("--alias prod --tmux user@prod") + "         # save as alias + tmux (no --save needed)",
		"  " + toolname.CLI("--host 192.168.1.1 --user root") + "        # legacy flags",
		"",
	}
	return strings.Join(lines, "\n")
}

func sshConnectHelpFull() string {
	return sshConnectHeadBlock() + SectionOpenSSHFlags + sshConnectAITBlock() +
		sshConnectNotSupportedBlock() + SectionOpenSSHNotSupported + sshConnectExamplesBlock()
}

func sshConnectShortBlock() string {
	return fmt.Sprintf(`  Connect flags (see "%s" for full detail):
    -p port  -l user  -i identity  -F config  -J jump  -o key=value
    --save  --alias name  --tmux  --password P
    --host H  --user U  --port N  --identity P  (legacy, same as ssh flags above)
`, toolname.CLI("--help"))
}
