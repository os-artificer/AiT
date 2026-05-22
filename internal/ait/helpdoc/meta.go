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

const metaCommands = `  help              show this help
  ai [question]     remote AI mode (!ait ai); optional first question
  ai chat <msg>     one-shot AI reply (!ait ai chat); does not enter AI mode
  !! [question]     alias for !ait ai
  !! chat <msg>     one-shot AI reply
  local             local mode (prompt: AiT Local>)
  remote            return to remote shell (prompt: AiT user@host>)
  suspend           suspend session and disconnect
  get <remote> <local>   download file (reuses SSH, no re-auth)
  put <local> <remote>   upload file (reuses SSH, no re-auth)
`

// WriteMetaHelp prints in-session !ait help to w (caller applies CRLF for SSH).
func WriteMetaHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro("AiT in-session commands (!ait prefix):"); err != nil {
		return err
	}
	if err := b.WriteUsage("!ait <subcommand> [args...]"); err != nil {
		return err
	}
	if err := b.WriteSection("Commands", metaCommands); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"!ait list",
		"!ait config get session.auth_ttl",
		"!ait session delete <id|alias>",
	); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\nFor login/load use: ait shell (AiT> prompt), not from inside SSH.\n"); err != nil {
		return err
	}
	return b.Finish("help")
}
