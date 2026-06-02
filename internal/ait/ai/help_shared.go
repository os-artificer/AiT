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

const (
	exampleAIProviderEnableDeepseek  = "ait ai provider enable deepseek"
	exampleAIProviderEnableOllamaYes = "ait ai provider enable ollama --yes"
	exampleAIProviderDisableMinimax  = "ait ai provider disable minimax"
	exampleAIOllamaStatus            = "ait ai ollama status"
	exampleAIOllamaSetupYes          = "ait ai ollama setup --yes"
	aiHelpFlagsBody                  = "  --model   Override AI provider for this session (uses provider model/base_url from config)\n"
	aiHelpTieredRoutingBody          = "When ai.tiered.enabled in config.yaml, short first-turn prompts use local Ollama; " +
		"longer or complex prompts use cloud providers. Setup: ait ai ollama setup. List: ait ai model list\n"
	aiHelpInSSHBody = "Use !ait ai or !! for interactive AI; " +
		"!ait ai ask <msg> or !! ask <msg> for one-shot; " +
		"!ait ai agent <msg> or !! agent <msg> for one-shot agent. " +
		"!ait ai as a meta subcommand is blocked."
)
