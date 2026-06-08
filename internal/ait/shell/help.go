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

package shell

import (
	"io"

	"github.com/os-artificer/AiT/internal/ait/login"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const shellIntro = "Start an interactive AiT command loop at the AiT> prompt."
const shellNote = "Type commands directly (no !ait prefix). Use exit or quit to leave."
const shellREPLCommands = `  help              show REPL help
  clear             clear terminal screen
  history           list or manage input history (history -c, -d, -s, -a, -r, -w)
  exit, quit        leave AiT shell
  ai                enter local AI mode (same as ait ai)
  login [flags] destination [command ...]  (--save or --alias to persist)
  load <id|alias> [--tmux] [--password]
  list              list saved sessions
  ll [path]         long directory listing (ls -la style; ait shell only)
  config, get, put  — same as: ait <command> ...
  session           — CLI: ait session --list | --load | --switch | --suspend | --delete
`

const shellREPLExtensions = `  ll [path]             ls -la style listing (builtin; not in ait ai or !ait)
  <AppRoot>/bin/<name>  user extensions (checked before ait CLI and PATH)
  <name> on PATH        system commands when not an ait subcommand
`

const shellREPLCommandsNoLogin = `  help              show REPL help
  clear             clear terminal screen
  history           list or manage input history (history -c, -d, -s, -a, -r, -w)
  exit, quit        leave AiT shell
  ai                enter local AI mode (same as ait ai)
  list              list saved sessions
  ll [path]         long directory listing (ls -la style; ait shell only)
  config, get, put  — same as: ait <command> ...
  session           — CLI: ait session --list | --load | --switch | --suspend | --delete
`

const (
	shellSectionExtensions    = "Extensions"
	shellSectionPrompts       = "Prompts"
	shellREPLHelpIntro        = "AiT shell — type commands directly, press Enter to run (no !ait prefix):"
	shellREPLHelpUsage        = "help | clear | history | login [flags] destination | load <id|alias> | list | ll [path] | exit"
	shellREPLHelpUsageNoLogin = "help | clear | history | list | ll [path] | exit"
	shellREPLHelpNoteTTY      = "In a TTY, Ctrl+L clears the screen and redraws the prompt (line buffer kept)."
	shellExampleLoginServer   = "login myserver"
)

const shellREPLPrompts = `  AiT>              not connected (this shell)
  AiT user@host:dir[:branch]>  remote SSH session (cwd + Git branch when in repo)
  AiT Local>        local mode inside SSH (!ait local)
`

const (
	shellSectionActiveSession = "Active session"
	shellActiveSessionBlock   = "  login and load are unavailable while a session is active.\n" +
		"  Use suspend to save the current session, or session --switch to change.\n"
)

// WriteShellHelp prints ait shell --help to w.
func WriteShellHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(shellIntro, shellNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait shell"); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionREPLCommands, shellREPLCommands); err != nil {
		return err
	}
	if err := b.WriteSection(shellSectionExtensions, shellREPLExtensions); err != nil {
		return err
	}
	if err := pkghelpdoc.WriteTTYKeysHelp(w, pkghelpdoc.TTYKeysShell); err != nil {
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
// When hasActiveSession is true, login and load are omitted from commands and usage.
func WriteShellREPLHelp(w io.Writer, hasActiveSession bool) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(shellREPLHelpIntro, shellREPLHelpNoteTTY); err != nil {
		return err
	}
	usage := shellREPLHelpUsage
	if hasActiveSession {
		usage = shellREPLHelpUsageNoLogin
	}
	if err := b.WriteUsage(usage); err != nil {
		return err
	}
	cmds := shellREPLCommands
	if hasActiveSession {
		cmds = shellREPLCommandsNoLogin
	}
	if err := b.WriteSection(pkghelpdoc.SectionCommands, cmds); err != nil {
		return err
	}
	if err := b.WriteSection(shellSectionExtensions, shellREPLExtensions); err != nil {
		return err
	}
	if hasActiveSession {
		if err := b.WriteSection(shellSectionActiveSession, shellActiveSessionBlock); err != nil {
			return err
		}
	} else {
		if err := login.WriteHelpShort(w); err != nil {
			return err
		}
	}
	if err := b.WriteSection(shellSectionPrompts, shellREPLPrompts); err != nil {
		return err
	}
	if err := pkghelpdoc.WriteTTYKeysHelp(w, pkghelpdoc.TTYKeysShell); err != nil {
		return err
	}
	examples := []string{
		shellExampleLoginServer,
		"login -p 22 root@192.168.1.1 -i ~/.ssh/id_rsa",
		"load myserver",
		"list",
	}
	if hasActiveSession {
		examples = []string{"list"}
	}
	if err := b.WriteExamples(examples...); err != nil {
		return err
	}
	return b.Finish("help")
}
