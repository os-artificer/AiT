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

import "io"

// LoginHelpFull is the canonical ait login --help text (golden; do not reword flags).
const LoginHelpFull = `Log in to a remote host (OpenSSH-compatible). Configuration is resolved via
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
  -p port         remote SSH port (overrides ssh_config and -o Port)
  -l user         remote username (same as ssh_config User)
  -i path         private key file (IdentityFile)
  -F path         ssh_config file (default: ~/.ssh/config)
  -J destination  ProxyJump host(s); comma-separated for a jump chain (e.g. -J bastion,internal)
  -o key=value    ssh_config option, repeatable (e.g. -o Port=22 -o User=deploy)
  -4, -6          address family hints (accepted; passed to ssh -G)
  -t, -T, -C, -v  accepted for ssh compatibility (no extra effect in ait login)
  -A, -a, -q, -Q  accepted for ssh compatibility

AiT flags:
  --save          save session to SQLite (no alias required; list by session id)
  --alias name    save session under this alias (implies --save); use with ait session --load <alias>
  --tmux          use remote tmux for shell persistence (interactive login only)
  --password P    SSH password (stored in OS keyring for session unlock)
  --host H        legacy: remote host (same as positional destination; do not combine)
  --user U        legacy: same as -l
  --port N        legacy: same as -p
  --identity P    legacy: same as -i

Not supported by ait login (use the system ssh client instead):
  -L, -R, -D, -X, -Y, -f, -N, -M, -S, -O, -W, -w

Examples:
  ait login myserver                              # Host alias from ~/.ssh/config
  ait login -p 22 root@host -i ~/.ssh/id_ed25519
  ait login -p22 deploy@10.0.0.5 -i ~/.ssh/id_ed25519
  ait login -F ~/.ssh/config.d/work -J bastion app.internal
  ait login -o Port=22 -o User=root legacy-host
  ait login user@host                             # interactive (not saved; use --save or --alias)
  ait login --save user@host                      # save session (no alias)
  ait login user@host uptime                      # one-shot (not saved unless --save or --alias)
  ait login --alias prod --tmux user@prod         # save as alias + tmux (no --save needed)
  ait login --host 192.168.1.1 --user root        # legacy flags

Flags:
  -h, --help      help for login
`

const loginHelpShort = `  login flags (see "ait login --help" for full detail):
    -p port  -l user  -i identity  -F config  -J jump  -o key=value
    --save  --alias name  --tmux  --password P
    --host H  --user U  --port N  --identity P  (legacy, same as ssh flags above)
`

// WriteLoginHelp prints full login usage, flags, and examples to w.
func WriteLoginHelp(w io.Writer) error {
	_, err := io.WriteString(w, LoginHelpFull)
	return err
}

// WriteLoginHelpShort prints a compact flag list for the ait shell REPL to w.
func WriteLoginHelpShort(w io.Writer) error {
	_, err := io.WriteString(w, loginHelpShort)
	return err
}
