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

// RootCommandEntry is one line in the root ait --help command list.
type RootCommandEntry struct {
	Name  string
	Short string
}

func formatRootSubcommands(entries []RootCommandEntry) string {
	var b strings.Builder
	for _, sc := range entries {
		fmt.Fprintf(&b, "  %-10s %s\n", sc.Name, sc.Short)
	}
	return b.String()
}

// WriteRootHelp prints ait --help to w using entries for the command list.
func WriteRootHelp(w io.Writer, entries []RootCommandEntry) error {
	const introMain = "AiT is an AI-enhanced SSH client with session management, optional tmux persistence, and Deepseek assistant."
	const introHint = "Run 'ait <command> --help' for command-specific usage and examples."
	const usage = "ait [command]"

	b := NewBody(w)
	if err := b.WriteIntro(introMain, introHint); err != nil {
		return err
	}
	if err := b.WriteUsage(usage); err != nil {
		return err
	}
	if err := b.WriteSection(SectionCommands, formatRootSubcommands(entries)); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait shell",
		"ait login myserver",
		"ait session --list",
	); err != nil {
		return err
	}
	return b.Finish("ait")
}
