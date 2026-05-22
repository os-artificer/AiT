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
	if err := b.WriteSection("Flags", "  --model   Override AI provider for this session (uses provider model/base_url from config)\n"); err != nil {
		return err
	}
	if err := b.WriteSection("REPL commands", aiREPLCommands+aiREPLCommandsLocalExtra); err != nil {
		return err
	}
	if err := b.WriteSection("Subcommands", "  chat [message...]   one-shot question (non-interactive)\n  provider enable|disable <id>\n  ollama status|setup|start|stop|disable\n"); err != nil {
		return err
	}
	if err := b.WriteSection("Tiered routing", "When ai.tiered.enabled in config.yaml, short first-turn prompts use local Ollama; longer or complex prompts use cloud providers. Setup: ait ai ollama setup. List: ait ai model list\n"); err != nil {
		return err
	}
	if err := b.WriteSection("In SSH", "Use !ait ai or !! for interactive AI; !ait ai chat <msg> or !! chat <msg> for one-shot. !ait ai as a meta subcommand is blocked."); err != nil {
		return err
	}
	if err := b.WriteExamples(
		"ait ai",
		"ait ai chat \"summarize this\"",
		"ait ai model list",
		"ait ai ollama status",
		"ait ai provider enable deepseek",
		"ait ai ollama setup --yes",
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
	if err := b.WriteSection("Commands", cmds); err != nil {
		return err
	}
	if err := b.WriteExamples(examples); err != nil {
		return err
	}
	return b.Finish("help")
}
