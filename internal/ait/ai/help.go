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

const aiIntroLocal = "Interactive AI mode on this machine (outside SSH). Commands run locally only after you confirm."
const aiIntroRemote = "Interactive AI mode for the active SSH session. Shell commands run on the remote host via run after you confirm."
const aiNote = "Requires ai.<provider>.api_key for cloud providers. " +
	"With ai.tiered.enabled, simple prompts use local Ollama. " +
	"Use --model to override cloud provider for this session."

const aiREPLCommands = `  help              show AI mode help
  exit, quit        leave AI mode
  clear             clear chat history (not input history; see history)
  history           list or manage input history (history -c, -d, -s, -a, -r, -w)
  mode ask|agent    Ask = Q&A only; Agent = plan and run commands (default: ask)
  approve each|session   how often to confirm runs (default: each)
  trust             same as approve session
  model             show current AI routing (provider, tiered)
  model list        list configured providers (same as ait ai model list)
  model <provider>  set cloud provider for this session (deepseek|minimax)
  run <command>     run a shell command after confirmation
  (deprecated: policy per-cmd|once — use approve each|session)
`

const aiREPLCommandsLocalExtra = `  ait <subcmd>      run an ait CLI subcommand after confirmation (local only)
`

// WriteAIHelp prints ait ai --help to w.
func WriteAIHelp(w io.Writer) error {
	b := pkghelpdoc.NewBody(w)
	if err := b.WriteIntro(aiIntroLocal, aiNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai [--model deepseek|minimax] [REPL-command ...]"); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionFlags, aiHelpFlagsBody); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionREPLCommands, aiREPLCommands+aiREPLCommandsLocalExtra); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionSubcommands, formatAISubcommands()); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionTieredRouting, aiHelpTieredRoutingBody); err != nil {
		return err
	}
	if err := b.WriteSection(pkghelpdoc.SectionInSSH, aiHelpInSSHBody); err != nil {
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
		exampleAIOllamaStatus,
		exampleAIProviderEnableDeepseek,
		exampleAIOllamaSetupYes,
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
