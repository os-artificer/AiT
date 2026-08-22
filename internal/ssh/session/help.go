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

package session

import (
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const sessionIntro = "List, load, suspend, switch, or delete saved SSH sessions in the local AiT database."

var (
	sessionNote = "Inside an active SSH session, use " + toolname.MetaLine("suspend") +
		" to suspend (not " + toolname.CLI("session --suspend") + "). " +
		"List # column matches interactive pick by number; ID column is display-only unless you set an alias."
	sessionFlags = `  --list              list saved sessions (# column for pick-by-number)
  --list -v           include FULL_ID (complete UUID) for scripting
  --load [<id|alias>] reconnect (omit value to pick interactively by #)
  --suspend           suspend from local CLI (prefer ` + toolname.MetaLine("suspend") + ` in SSH)
  --switch [id|alias] switch to another session (omit value to pick interactively)
  --delete <id|alias|all>   delete one session or all sessions
  --tmux              use remote tmux for --load or --switch
  --password P        SSH password when session auth has expired
  --force             force switch or skip delete confirmation
`
	sessionUsageLine = toolname.CLI("session [--list [-v] | --load [<id|alias>] | --switch [id|alias] | --suspend | --delete <id|alias|all>]")
)

var sessionExampleDeleteAllForce = toolname.CLI("session --delete all --force")

// WriteSessionHelp prints aissh session --help to w.
func WriteSessionHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(sessionIntro, sessionNote); err != nil {
		return err
	}
	if err := b.WriteUsage(sessionUsageLine); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionAITFlags, sessionFlags); err != nil {
		return err
	}
	if err := b.WriteExamples(
		toolname.CLI("session --list"),
		toolname.CLI("session --list -v"),
		toolname.CLI("session --load"),
		toolname.CLI("session --load prod --tmux"),
		toolname.CLI("session --switch"),
		toolname.CLI("session --switch prod --force"),
		sessionExampleDeleteAllForce,
		pkghelpdoc.ExampleMetaSuspend+"   # inside SSH session",
	); err != nil {
		return err
	}
	return b.Finish("session")
}
