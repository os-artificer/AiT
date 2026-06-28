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
	"fmt"
	"io"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const shellIntro = "Start an interactive AiT command loop at the AiT> prompt."

var shellNote = "Type commands directly (no " + toolname.MetaPrefix + " prefix). Use exit or quit to leave."

var shellREPLCommands = fmt.Sprintf(`  help              show REPL help
  clear             clear terminal screen
  history           list or manage input history (history -c, -d, -s, -a, -r, -w)
  exit, quit        leave AiT shell
  ai                enter local AI mode (same as %s)
  [flags] destination [command ...]  connect (ssh-style; --save or --alias to persist)
  load <id|alias> [--tmux] [--password]
  list              list saved sessions
  ll [path]         long directory listing (ls -la style; %s only)
  config, get, put  — same as: %s <command> ...
  session           — CLI: %s | %s | %s | %s | %s
`, toolname.CLI("ai"), toolname.CLI("shell"), toolname.AISsh,
	toolname.CLI("session --list"), toolname.CLI("session --load"),
	toolname.CLI("session --switch"), toolname.CLI("session --suspend"),
	toolname.CLI("session --delete"))

var shellREPLExtensions = fmt.Sprintf(`  ll [path]             ls -la style listing (builtin; not in %s or %s)
  <AppRoot>/bin/<name>  user extensions (checked before %s CLI and PATH)
  <name> on PATH        system commands when not an %s subcommand
`, toolname.CLI("ai"), toolname.MetaPrefix, toolname.AISsh, toolname.AISsh)

var shellREPLCommandsNoLoad = fmt.Sprintf(`  help              show REPL help
  clear             clear terminal screen
  history           list or manage input history (history -c, -d, -s, -a, -r, -w)
  exit, quit        leave AiT shell
  ai                enter local AI mode (same as %s)
  list              list saved sessions
  ll [path]         long directory listing (ls -la style; %s only)
  config, get, put  — same as: %s <command> ...
  session           — CLI: %s | %s | %s | %s | %s
`, toolname.CLI("ai"), toolname.CLI("shell"), toolname.AISsh,
	toolname.CLI("session --list"), toolname.CLI("session --load"),
	toolname.CLI("session --switch"), toolname.CLI("session --suspend"),
	toolname.CLI("session --delete"))

const (
	shellSectionExtensions    = "Extensions"
	shellSectionPrompts       = "Prompts"
	shellREPLHelpUsage        = "help | clear | history | [flags] destination | load <id|alias> | list | ll [path] | exit"
	shellREPLHelpUsageNoLoad  = "help | clear | history | list | ll [path] | exit"
	shellREPLHelpNoteTTY      = "In a TTY, Ctrl+L clears the screen and redraws the prompt (line buffer kept)."
	shellExampleConnectServer = "myserver"
)

var (
	shellREPLHelpIntro = "AiT shell — type commands directly, press Enter to run (no " + toolname.MetaPrefix + " prefix):"
	shellREPLPrompts   = `  AiT>              not connected (this shell)
  AiT user@host:dir[:branch]>  remote SSH session (cwd + Git branch when in repo)
  AiT Local>        local mode inside SSH (` + pkghelpdoc.ExampleMetaLocal + `)
`
)

const (
	shellSectionActiveSession = "Active session"
	shellActiveSessionBlock   = "  SSH connect and load are unavailable while a session is active.\n" +
		"  Use suspend to save the current session, or session --switch to change.\n"
)

// WriteShellHelp prints aitssh shell --help to w.
func WriteShellHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(shellIntro, shellNote); err != nil {
		return err
	}
	if err := b.WriteUsage(toolname.CLI("shell")); err != nil {
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
		toolname.CLI("shell"),
		shellExampleConnectServer,
		"load prod --tmux",
	); err != nil {
		return err
	}
	return b.Finish("shell")
}

// WriteShellREPLHelp prints in-shell help (Usage/Examples sections; caller applies CRLF).
// When hasActiveSession is true, connect and load are omitted from commands and usage.
func WriteShellREPLHelp(w io.Writer, hasActiveSession bool) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(shellREPLHelpIntro, shellREPLHelpNoteTTY); err != nil {
		return err
	}
	usage := shellREPLHelpUsage
	if hasActiveSession {
		usage = shellREPLHelpUsageNoLoad
	}
	if err := b.WriteUsage(usage); err != nil {
		return err
	}
	cmds := shellREPLCommands
	if hasActiveSession {
		cmds = shellREPLCommandsNoLoad
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
		if err := pkghelpdoc.WriteSSHConnectShort(w); err != nil {
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
		shellExampleConnectServer,
		"-p 22 root@192.168.1.1 -i ~/.ssh/id_rsa",
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
