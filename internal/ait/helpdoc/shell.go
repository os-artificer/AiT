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

const shellIntro = "Start an interactive AiT command loop at the AiT> prompt."
const shellNote = "Type commands directly (no !ait prefix). Use exit or quit to leave."
const shellREPLCommands = `  help              show REPL help
  clear             clear terminal screen
  exit, quit        leave AiT shell
  ai                enter local AI mode (same as ait ai)
  login [flags] destination [command ...]  (--save or --alias to persist)
  load <id|alias> [--tmux] [--password]
  list              list saved sessions
  config, get, put  — same as: ait <command> ...
  session           — CLI: ait session --list | --load | --switch | --suspend | --delete
`

const (
	shellSectionPrompts     = "Prompts"
	shellREPLHelpIntro      = "AiT shell — type commands directly, press Enter to run (no !ait prefix):"
	shellREPLHelpUsage      = "help | clear | login [flags] destination | load <id|alias> | list | exit"
	shellREPLHelpNoteTTY    = "In a TTY, Ctrl+L clears the screen and redraws the prompt (line buffer kept)."
	shellExampleLoginServer = "login myserver"
)

const shellREPLPrompts = `  AiT>              not connected (this shell)
  AiT user@host>    remote SSH session
  AiT Local>        local mode inside SSH (!ait local)
`

// WriteShellHelp prints ait shell --help to w.
func WriteShellHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(shellIntro, shellNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait shell"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionREPLCommands, shellREPLCommands); err != nil {
		return err
	}
	if err := WriteTTYKeysHelp(w, TTYKeysShell); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait shell",
		shellExampleLoginServer,
		"load prod --tmux",
	); err != nil {
		return err
	}
	return b.Finish("shell")
}

// WriteShellREPLHelp prints in-shell help (Usage/Examples sections; caller applies CRLF).
func WriteShellREPLHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(shellREPLHelpIntro, shellREPLHelpNoteTTY); err != nil {
		return err
	}
	if err := b.WriteUsage(shellREPLHelpUsage); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionCommands, shellREPLCommands); err != nil {
		return err
	}
	if err := WriteLoginHelpShort(w); err != nil {
		return err
	}
	if err := b.WriteSection(shellSectionPrompts, shellREPLPrompts); err != nil {
		return err
	}
	if err := WriteTTYKeysHelp(w, TTYKeysShell); err != nil {
		return err
	}
	if err := b.WriteExamples(
		shellExampleLoginServer,
		"login -p 22 root@192.168.1.1 -i ~/.ssh/id_rsa",
		"load myserver",
		"list",
	); err != nil {
		return err
	}
	return b.Finish("help")
}
