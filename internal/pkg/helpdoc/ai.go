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
	"io"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// AI help strings — single source for aitssh ai, ask, agent, ollama, and in-REPL help.
var (
	ExampleAIProviderEnableDeepseek  = toolname.CLI("ai provider enable deepseek")
	ExampleAIProviderEnableOllamaYes = toolname.CLI("ai provider enable ollama --yes")
	ExampleAIProviderDisableMinimax  = toolname.CLI("ai provider disable minimax")
	ExampleAIOllamaStatus            = toolname.CLI("ai ollama status")
	ExampleAIOllamaSetupYes          = toolname.CLI("ai ollama setup --yes")
	ExampleAIOllamaStartWait         = toolname.CLI("ai ollama start --wait")
	ExampleAIOllamaDisableStop       = toolname.CLI("ai ollama disable --stop")

	ExampleAIAskSSH    = toolname.MetaLine("ai ask how do I find large files")
	ExampleAIAgentSSH  = toolname.MetaLine("ai agent how do I find large files")
	ExampleMetaSuspend = toolname.MetaLine("suspend")
	ExampleMetaLocal   = toolname.MetaLine("local")

	AIHelpFlagsBody     = "  --model   Override AI provider for this session (uses provider model/base_url from config)\n"
	AIChatHelpFlagsBody = "  --model   Override cloud provider (deepseek|minimax); inherited from " + toolname.CLI("ai") + "\n"

	AIHelpTieredRoutingBody = "When ai.tiered.enabled in config.yaml, short first-turn prompts use local Ollama; " +
		"longer or complex prompts use cloud providers. Setup: " + toolname.CLI("ai ollama setup") +
		". List: " + toolname.CLI("ai model list") + "\n"
	AIHelpInSSHBody = "Use " + toolname.MetaLine("ai") + " or !! for interactive AI; " +
		toolname.MetaLine("ai ask <msg>") + " or !! ask <msg> for one-shot; " +
		toolname.MetaLine("ai agent") + " for interactive Agent REPL, or " +
		toolname.MetaLine("ai agent <msg>") + " / !! agent <msg> for one-shot agent. " +
		toolname.MetaLine("ai") + " as a meta subcommand is blocked."

	AIIntroLocal  = "Interactive AI mode on this machine (outside SSH). Commands run locally only after you confirm."
	AIIntroRemote = "Interactive AI mode for the active SSH session. Shell commands run on the remote host via run after you confirm."

	AIReplCommands = `  help              show AI mode help
  exit, quit        leave AI mode
  clear             clear chat history (not input history; see history)
  history           list or manage input history (history -c, -d, -s, -a, -r, -w)
  mode ask|agent    Ask = Q&A only; Agent = plan and run commands (default: agent)
  approve each|session   how often to confirm runs (default: each)
  trust             same as approve session
  model             show current AI routing (provider, tiered)
  model list        list configured providers (same as ` + toolname.CLI("ai model list") + `)
  model <provider>  set cloud provider for this session (deepseek|minimax)
  run <command>     run a shell command after confirmation
  (deprecated: policy per-cmd|once — use approve each|session)
`

	AIReplCommandsLocalExtra = "  " + toolname.AISsh + " <subcmd>      run an " + toolname.AISsh + " CLI subcommand after confirmation (local only)\n"
)

// WriteAIChatSubcommandHelp prints one-shot ask/agent --help using shared layout.
// cmdName is the subcommand name for the footer (e.g. "ask", "agent").
func WriteAIChatSubcommandHelp(w io.Writer, cmdName, introMain, introSSH, usageLine, inputBody string, examples ...string) error {
	b := NewBody(w)
	if err := b.WriteIntro(introMain, introSSH); err != nil {
		return err
	}
	if err := b.WriteUsage(usageLine); err != nil {
		return err
	}
	if err := b.WriteSection(SectionInput, inputBody); err != nil {
		return err
	}
	if err := b.WriteSection(SectionFlags, AIChatHelpFlagsBody); err != nil {
		return err
	}
	if err := b.WriteExamples(examples...); err != nil {
		return err
	}
	return b.Finish(cmdName)
}
