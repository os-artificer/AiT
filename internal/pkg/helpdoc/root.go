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

	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// RootCommandEntry is one line in the root ai-ssh --help command list.
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

// WriteRootHelp prints ai-ssh --help to w using entries for the command list.
func WriteRootHelp(w io.Writer, entries []RootCommandEntry) error {
	introMain := "AiT is an AI-enhanced SSH client with session management, optional tmux persistence, and Deepseek assistant."
	introHint := fmt.Sprintf("Run '%s <command> --help' for command-specific usage and examples.", toolname.AISsh)
	usage := toolname.AISsh + " [command]"

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
		toolname.CLI("shell"),
		toolname.CLI("login myserver"),
		clihints.TrySessionList,
	); err != nil {
		return err
	}
	return b.Finish(toolname.AISsh)
}
