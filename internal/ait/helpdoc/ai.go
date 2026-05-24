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

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

const aiIntroLocal = "Interactive AI mode on this machine (outside SSH). Commands run locally only after you confirm."
const aiIntroRemote = "Interactive AI mode for the active SSH session. Shell commands run on the remote host via run after you confirm."
const aiNote = "Requires ai.<provider>.api_key for cloud providers. With ai.tiered.enabled, simple prompts use local Ollama. Use --model to override cloud provider for this session."

const aiREPLCommands = `  help              show AI mode help
  exit, quit        leave AI mode
  clear             clear chat history
  model             show current AI routing (provider, tiered)
  model list        list configured providers (same as ait ai model list)
  model <provider>  set cloud provider for this session (deepseek|minimax)
  policy per-cmd    confirm each run (default)
  policy once       confirm once per AI session (alias: trust)
  run <command>     run a shell command after confirmation
`

const aiREPLCommandsLocalExtra = `  ait <subcmd>      run an ait CLI subcommand after confirmation (local only)
`

// WriteAIHelp prints ait ai --help to w.
func WriteAIHelp(w io.Writer) error {
	b := newHelpBody(w)
	if err := b.WriteIntro(aiIntroLocal, aiNote); err != nil {
		return err
	}
	if err := b.WriteUsage("ait ai [--model deepseek|minimax] [REPL-command ...]"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionFlags, aiHelpFlagsBody); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionREPLCommands, aiREPLCommands+aiREPLCommandsLocalExtra); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionSubcommands, aiHelpSubcommandsBody); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionTieredRouting, aiHelpTieredRoutingBody); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionInSSH, aiHelpInSSHBody); err != nil {
		return err
	}
	if err := WriteTTYKeysHelp(w, TTYKeysAI); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai",
		"ait ai chat \"summarize this\"",
		"ait ai model list",
		exampleAIOllamaStatus,
		exampleAIProviderEnableDeepseek,
		exampleAIOllamaSetupYes,
		"ait ai --model minimax",
		"run echo hello",
		"policy once",
	); err != nil {
		return err
	}
	return b.Finish("ai")
}

// WriteAIREPLHelp prints in-AI-mode help for the given scope.
func WriteAIREPLHelp(w io.Writer, scope ai.Scope) error {
	b := newHelpBody(w)
	intro := aiIntroLocal
	cmds := aiREPLCommands + aiREPLCommandsLocalExtra
	examples := "run ls\nait list\npolicy per-cmd\n"
	if scope == ai.ScopeRemote {
		intro = aiIntroRemote
		cmds = aiREPLCommands
		examples = "run uptime\npolicy once\n"
	}
	if err := b.WriteIntro(intro); err != nil {
		return err
	}
	if err := b.WriteUsage("help | model [provider] | run <cmd> | policy per-cmd|once | clear | exit"); err != nil {
		return err
	}
	if err := b.WriteSection(helpSectionCommands, cmds); err != nil {
		return err
	}
	if err := WriteTTYKeysHelp(w, TTYKeysAI); err != nil {
		return err
	}
	if err := b.WriteExamples(examples); err != nil {
		return err
	}
	return b.Finish("help")
}
