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

// Common — shared help section titles and argv tokens.
const (
	SectionSubcommands       = "Subcommands"
	SectionFlags             = "Flags"
	SectionArguments         = "Arguments"
	SectionAITFlags          = "AiT flags"
	SectionCommands          = "Commands"
	SectionKeys              = "Keys"
	SectionProviders         = "Providers"
	SectionREPLCommands      = "REPL commands"
	SectionInput             = "Input"
	SectionPaths             = "Paths"
	SectionFlagOrder         = "Flag order"
	SectionOpenSSHFlagsTitle = "OpenSSH-compatible flags"
	SectionTieredRouting     = "Tiered routing"
	SectionInSSH             = "In SSH"
	SectionOutput            = "Output"
	SectionFlagsOllama       = "Flags (ollama)"
	helpArgH                 = "-h"
	helpArgHelp              = "--help"
	helpArgCmd               = "help"
)

// OpenSSH — blocks embedded by login and transfer help.
const (
	// SectionOpenSSHFlags is the shared OpenSSH-compatible flags block (same as ait login --help).
	SectionOpenSSHFlags = `  -p port         remote SSH port (overrides ssh_config and -o Port)
  -l user         remote username (same as ssh_config User)
  -i path         private key file (IdentityFile)
  -F path         ssh_config file (default: ~/.ssh/config)
  -J destination  ProxyJump host(s); comma-separated for a jump chain (e.g. -J bastion,internal)
  -o key=value    ssh_config option, repeatable (e.g. -o Port=22 -o User=deploy)
  -4, -6          address family hints (accepted; passed to ssh -G)
  -t, -T, -C, -v  accepted for ssh compatibility (no extra effect in ait login)
  -A, -a, -q, -Q  accepted for ssh compatibility
`

	// SectionOpenSSHNotSupported is the shared not-supported flags list for SSH-style commands.
	SectionOpenSSHNotSupported = `  -L, -R, -D, -X, -Y, -f, -N, -M, -S, -O, -W, -w
`
)

// AI — examples and bodies reused across ai, provider, and ollama help.
const (
	exampleAIProviderEnableDeepseek  = "ait ai provider enable deepseek"
	exampleAIProviderEnableOllamaYes = "ait ai provider enable ollama --yes"
	exampleAIProviderDisableMinimax  = "ait ai provider disable minimax"
	exampleAIOllamaStatus            = "ait ai ollama status"
	exampleAIOllamaSetupYes          = "ait ai ollama setup --yes"
	exampleAIOllamaStartWait         = "ait ai ollama start --wait"
	exampleAIOllamaDisableStop       = "ait ai ollama disable --stop"
	aiHelpSubcommandsBody            = "  ask [message...]     one-shot question (non-interactive)\n" +
		"  agent [message...]   one-shot agent with tool execution (non-interactive)\n" +
		"  provider enable|disable <id>\n" +
		"  ollama status|setup|start|stop|disable\n"
	aiHelpFlagsBody         = "  --model   Override AI provider for this session (uses provider model/base_url from config)\n"
	aiHelpTieredRoutingBody = "When ai.tiered.enabled in config.yaml, short first-turn prompts use local Ollama; " +
		"longer or complex prompts use cloud providers. Setup: ait ai ollama setup. List: ait ai model list\n"
	aiHelpInSSHBody = "Use !ait ai or !! for interactive AI; " +
		"!ait ai ask <msg> or !! ask <msg> for one-shot; " +
		"!ait ai agent <msg> or !! agent <msg> for one-shot agent. " +
		"!ait ai as a meta subcommand is blocked."
	aiChatHelpFlagsBody  = "  --model   Override cloud provider (deepseek|minimax); inherited from ait ai\n"
	aiAgentHelpFlagsBody = "  --model   Override cloud provider (deepseek|minimax); inherited from ait ai\n"
)
