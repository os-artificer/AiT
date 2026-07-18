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

	ollamasub "github.com/os-artificer/AiT/internal/aissh/ai/ollama/subregistry"
	providersub "github.com/os-artificer/AiT/internal/aissh/ai/provider/subregistry"
	"github.com/os-artificer/AiT/internal/aissh/ai/subregistry"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	pkgsubregistry "github.com/os-artificer/AiT/internal/pkg/subregistry"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const aiNote = "Experimental (off by default): enable with --ai-experimental, " +
	"AIT_AI_EXPERIMENTAL=1, or ai.experimental.enabled=true. " +
	"Requires ai.<provider>.api_key for cloud providers. " +
	"With ai.tiered.enabled, simple prompts use local Ollama. " +
	"Use --model to override cloud provider for this session."

var aiUsageLine = toolname.CLI("ai [--model deepseek|minimax] [REPL-command ...]")

// WriteAIHelp prints aitssh ai --help to w.
func WriteAIHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(pkghelpdoc.AIIntroLocal, aiNote); err != nil {
		return err
	}
	if err := b.WriteUsage(aiUsageLine); err != nil {
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
		toolname.CLI("ai"),
		toolname.CLI(`ai ask "summarize this"`),
		toolname.CLI(`ai agent "find large files in /var/log"`),
		toolname.CLI("ai model list"),
		pkghelpdoc.ExampleAIOllamaStatus,
		pkghelpdoc.ExampleAIProviderEnableDeepseek,
		pkghelpdoc.ExampleAIOllamaSetupYes,
		toolname.CLI("ai --model minimax"),
		"run echo hello",
		"mode agent",
		"approve session",
	); err != nil {
		return err
	}
	return b.Finish("ai")
}

func formatAISubcommands() string {
	providerChildren := pkgsubregistry.JoinChildNames(providersub.Entries(), "|")
	ollamaChildren := pkgsubregistry.JoinChildNames(ollamasub.Entries(), "|")

	var b strings.Builder
	for _, e := range subregistry.Entries() {
		switch e.Name {
		case "ask", "agent":
			fmt.Fprintf(&b, "  %s [message...]     %s\n", e.Name, e.Short)
		case "provider":
			fmt.Fprintf(&b, "  %s %s <id>\n", e.Name, providerChildren)
		case "ollama":
			fmt.Fprintf(&b, "  %s %s\n", e.Name, ollamaChildren)
		default:
			fmt.Fprintf(&b, "  %s   %s\n", e.Name, e.Short)
		}
	}
	return b.String()
}
