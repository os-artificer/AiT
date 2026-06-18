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

package meta

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const (
	metaHelpIntro    = "AiT in-session commands (!ait prefix):"
	metaHelpUsage    = "!ait <subcommand> [args...]"
	metaHelpTailNote = "\nlogin/load/switch suspend the current session (auto-save if needed) and connect to the target.\n"
)

const metaCommands = `  help              show this help
  ai [question]     remote AI mode (!ait ai); optional first question
  ai ask <msg>     one-shot AI reply (!ait ai ask); does not enter AI mode
  ai agent [msg]   Agent REPL when no msg; one-shot when msg given
  !! [question]     alias for !ait ai
  !! ask <msg>     one-shot AI reply
  !! agent [msg]   Agent REPL when no msg; one-shot when msg given
  login [flags] dest   suspend current session; log in to another host
  load [id|alias]      suspend current session; load a saved session
  session --switch [id|alias]   suspend current; switch to another session
  local             local mode (prompt: AiT Local>)
  remote            return to remote shell (prompt: AiT user@host:dir[:branch]>)
  chain [on|off]    toggle chain mode (PTY-pipe transfer for SSH hops)
  suspend           suspend session and disconnect (auto-save if needed)
  get <remote> <local>   download file (reuses SSH, no re-auth)
  put <local> <remote>   upload file (reuses SSH, no re-auth)
`

// WriteMetaHelp prints in-session !ait help to w (caller applies CRLF for SSH).
func WriteMetaHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(metaHelpIntro); err != nil {
		return err
	}
	if err := b.WriteUsage(metaHelpUsage); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionCommands, metaCommands); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"!ait list",
		"!ait login user@hostB",
		"!ait session --switch prod",
		"!ait config get session.auth_ttl",
		"!ait session delete <id|alias>",
	); err != nil {
		return err
	}
	if _, err := io.WriteString(w, metaHelpTailNote); err != nil {
		return err
	}
	return b.Finish("help")
}
