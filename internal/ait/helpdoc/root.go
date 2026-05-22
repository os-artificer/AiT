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
	"io"
	"strings"
)

// rootSubcommands mirrors cli/root.go AddCommand order; update both when adding commands.
var rootSubcommands = []rootSubcommand{
	{"shell", "interactive AiT> REPL"},
	{"ai", "local AI REPL; provider enable/disable; ollama setup/status"},
	{"login", "log in and save session (OpenSSH-compatible)"},
	{"list", "list saved sessions"},
	{"load", "reconnect to a saved session"},
	{"suspend", "suspend current session (prefer !ait suspend in SSH)"},
	{"get", "download a file via SFTP"},
	{"put", "upload a file via SFTP"},
	{"config", "read/write configuration"},
	{"session", "switch or delete saved sessions"},
}

const (
	rootHelpIntroMain = "AiT is an AI-enhanced SSH client with session management, optional tmux persistence, and Deepseek assistant."
	rootHelpIntroHint = "Run 'ait <command> --help' for command-specific usage and examples."
	rootHelpUsage     = "ait [command]"
)

type rootSubcommand struct {
	name string
	desc string
}

func formatRootSubcommands() string {
	var b strings.Builder
	for _, sc := range rootSubcommands {
		fmt.Fprintf(&b, "  %-10s %s\n", sc.name, sc.desc)
	}
	return b.String()
}

// WriteRootHelp prints ait --help to w.
func WriteRootHelp(w io.Writer) error {
	body := newHelpBody(w)
	if err := body.WriteIntro(rootHelpIntroMain, rootHelpIntroHint); err != nil {
		return err
	}
	if err := body.WriteUsage(rootHelpUsage); err != nil {
		return err
	}
	if err := body.WriteSection(helpSectionCommands, formatRootSubcommands()); err != nil {
		return err
	}
	if err := body.WriteExamples(
		"ait shell",
		"ait login myserver",
		"ait list",
	); err != nil {
		return err
	}
	return body.Finish("ait")
}
