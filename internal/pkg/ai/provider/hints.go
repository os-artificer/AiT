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

package provider

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

var hintMetaAIAsk = toolname.AICLI("ask <message>")

const (
	hintSetCloudKeysOrOllama   = "Set a cloud provider API key or enable tiered Ollama routing."
	hintSetCloudAPIKey         = "Set an API key for a cloud provider."
	hintStoreAPIKeyInKeyring   = "Store the provider API key in the OS keyring."
	hintProvideChatMessage     = "Provide a message on the command line or via stdin."
	hintEnableProviderOverride = "Enable the provider before using it as --model override."
)

var (
	hintSetDeepseekAPIKey = toolname.CLI("config set ai.deepseek.api_key <key>")
	hintSetMinimaxAPIKey  = toolname.CLI("config set ai.minimax.api_key <key>")
	hintEnableOllama      = toolname.AICLI("provider enable ollama")
	hintAIHelp            = toolname.AICLI("--help")
	hintOllamaSetup       = toolname.AICLI("ollama setup")
	hintAIAskWithMessage  = toolname.AICLI("ask <message>")
	hintAIAskStdin        = "echo ping | " + toolname.AICLI("ask")
	hintAIAskHelp         = toolname.AICLI("ask --help")

	// HintAIHelp is the standard ai --help Try line (shared with client).
	HintAIHelp = hintAIHelp
	// HintOllamaSetup is the ai ollama setup Try line.
	HintOllamaSetup = hintOllamaSetup
	// HintStoreAPIKeyInKeyring explains keyring storage in API key errors.
	HintStoreAPIKeyInKeyring = hintStoreAPIKeyInKeyring
	// HintSetCloudAPIKey is the one-line cloud API key setup hint.
	HintSetCloudAPIKey = hintSetCloudAPIKey
)

// ConfigSetAPIKeyCmd returns the aissh config set command for a provider API key.
func ConfigSetAPIKeyCmd(id string) string {
	return toolname.CLI(fmt.Sprintf("config set ai.%s.api_key <key>", id))
}

// ConfigGetAPIKeyCmd returns the aissh config get command for a provider API key.
func ConfigGetAPIKeyCmd(id string) string {
	return toolname.CLI(fmt.Sprintf("config get ai.%s.api_key", id))
}

// ProviderEnableCmd returns the ai provider enable command for id.
func ProviderEnableCmd(id string) string {
	return toolname.AICLI("provider enable " + id)
}
