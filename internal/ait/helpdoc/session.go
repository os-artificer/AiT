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

const sessionIntro = "List, load, suspend, switch, or delete saved SSH sessions in the local AiT database."
const sessionNote = "Inside an active SSH session, use !ait suspend to suspend (not ait session --suspend)."

const sessionFlags = `  --list              list saved sessions
  --load <id|alias>   reconnect to a saved session
  --suspend           suspend from local CLI (prefer !ait suspend in SSH)
  --switch [id|alias] switch to another session (omit value to pick interactively)
  --delete <id|alias|all>   delete one session or all sessions
  --tmux              use remote tmux for --load or --switch
  --password P        SSH password when session auth has expired
  --force             force switch or skip delete confirmation
`

const sessionExampleDeleteAllForce = "ait session --delete all --force"

// WriteSessionHelp prints ait session --help to w.
func WriteSessionHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(sessionIntro, sessionNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait session [--list | --load <id|alias> | --switch [id|alias] | --suspend | --delete <id|alias|all>]"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionAITFlags, sessionFlags); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait session --list",
		"ait session --load prod --tmux",
		"ait session --switch",
		"ait session --switch prod --force",
		sessionExampleDeleteAllForce,
		"!ait suspend   # inside SSH session",
	); err != nil {
		return err
	}
	return b.Finish("session")
}
