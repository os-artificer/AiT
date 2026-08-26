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

// AI help strings — single source for ai plugin, ask, agent, ollama, and in-REPL help.
var (
	ExampleAIProviderEnableDeepseek  = toolname.AICLI("provider enable deepseek")
	ExampleAIProviderEnableOllamaYes = toolname.AICLI("provider enable ollama --yes")
	ExampleAIProviderDisableMinimax  = toolname.AICLI("provider disable minimax")
	ExampleAIOllamaStatus            = toolname.AICLI("ollama status")
	ExampleAIOllamaSetupYes          = toolname.AICLI("ollama setup --yes")
	ExampleAIOllamaStartWait         = toolname.AICLI("ollama start --wait")
	ExampleAIOllamaDisableStop       = toolname.AICLI("ollama disable --stop")

	ExampleAIAskLocal   = toolname.AICLI(`ask "how do I find large files"`)
	ExampleAIAgentLocal = toolname.AICLI(`agent "check disk usage"`)
	ExampleMetaSuspend  = toolname.MetaLine("suspend")
	ExampleMetaLocal    = toolname.MetaLine("local")

	AIHelpFlagsBody     = "  --model   Override AI provider for this session (uses provider model/base_url from config)\n"
	AIChatHelpFlagsBody = "  --model   Override cloud provider (deepseek|minimax); inherited from " + toolname.AICLI("") + "\n"

	AIHelpTieredRoutingBody = "When ai.tiered.enabled in config.yaml, short first-turn prompts use local Ollama; " +
		"longer or complex prompts use cloud providers. Setup: " + toolname.AICLI("ollama setup") +
		". List: " + toolname.AICLI("model list") + "\n"
	AIHelpInSSHBody = "AI is provided by the " + toolname.AI + " plugin (experimental, off by default). " +
		"From the toolbox REPL: " + toolname.AI + ". " +
		"Remote Agent: " + toolname.AICLI("connect user@host") + ". " +
		"In SSH sessions, !! and !ait ai are not available. " +
		"Enable with --ai-experimental on ai, AIT_AI_EXPERIMENTAL=1, or ai.experimental.enabled."

	AIIntroLocal = "Ops AI Agent plugin (experimental, off by default): one-shot ask/agent, model list, " +
		"provider/ollama, and remote connect. Interactive Agent Loop: " + toolname.AIT + " agent."
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
