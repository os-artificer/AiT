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

package ai

import (
	"fmt"
	"io"
	"strings"

	"github.com/os-artificer/AiT/internal/ait/ai/subregistry"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
)

const aiNote = "Requires ai.<provider>.api_key for cloud providers. " +
	"With ai.tiered.enabled, simple prompts use local Ollama. " +
	"Use --model to override cloud provider for this session."

// WriteAIHelp prints ait ai --help to w.
func WriteAIHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(pkghelpdoc.AIIntroLocal, aiNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai [--model deepseek|minimax] [REPL-command ...]"); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlags, pkghelpdoc.AIHelpFlagsBody); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionREPLCommands, pkghelpdoc.AIReplCommands+pkghelpdoc.AIReplCommandsLocalExtra); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionSubcommands, formatAISubcommands()); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionTieredRouting, pkghelpdoc.AIHelpTieredRoutingBody); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionInSSH, pkghelpdoc.AIHelpInSSHBody); err != nil {
		return err
	}
	if err := pkghelpdoc.WriteTTYKeysHelp(w, pkghelpdoc.TTYKeysAI); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai",
		"ait ai ask \"summarize this\"",
		"ait ai agent \"find large files in /var/log\"",
		"ait ai model list",
		pkghelpdoc.ExampleAIOllamaStatus,
		pkghelpdoc.ExampleAIProviderEnableDeepseek,
		pkghelpdoc.ExampleAIOllamaSetupYes,
		"ait ai --model minimax",
		"run echo hello",
		"mode agent",
		"approve session",
	); err != nil {
		return err
	}
	return b.Finish("ai")
}

func formatAISubcommands() string {
	var b strings.Builder
	for _, e := range subregistry.Entries() {
		switch e.Name {
		case "ask", "agent":
			fmt.Fprintf(&b, "  %s [message...]     %s\n", e.Name, e.Short)
		case "provider":
			fmt.Fprintf(&b, "  %s enable|disable <id>\n", e.Name)
		case "ollama":
			fmt.Fprintf(&b, "  %s status|setup|start|stop|disable\n", e.Name)
		default:
			fmt.Fprintf(&b, "  %s   %s\n", e.Name, e.Short)
		}
	}
	return b.String()
}
