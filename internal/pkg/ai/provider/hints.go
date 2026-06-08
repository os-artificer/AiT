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

import "fmt"

const (
	hintSetCloudKeysOrOllama   = "Set a cloud provider API key or enable tiered Ollama routing."
	hintSetDeepseekAPIKey      = "ait config set ai.deepseek.api_key <key>"
	hintSetMinimaxAPIKey       = "ait config set ai.minimax.api_key <key>"
	hintEnableOllama           = "ait ai provider enable ollama"
	hintAIHelp                 = "ait ai --help"
	hintOllamaSetup            = "ait ai ollama setup"
	hintSetCloudAPIKey         = "Set an API key for a cloud provider."
	hintStoreAPIKeyInKeyring   = "Store the provider API key in the OS keyring."
	hintProvideChatMessage     = "Provide a message on the command line or via stdin."
	hintAIAskWithMessage       = "ait ai ask <message>"
	hintAIAskStdin             = "echo ping | ait ai ask"
	hintMetaAIAsk              = "!ait ai ask <message>"
	hintAIAskHelp              = "ait ai ask --help"
	hintEnableProviderOverride = "Enable the provider before using it as --model override."

	// HintAIHelp is the standard ait ai --help Try line (shared with client).
	HintAIHelp = hintAIHelp
	// HintOllamaSetup is the ait ai ollama setup Try line.
	HintOllamaSetup = hintOllamaSetup
	// HintStoreAPIKeyInKeyring explains keyring storage in API key errors.
	HintStoreAPIKeyInKeyring = hintStoreAPIKeyInKeyring
	// HintSetCloudAPIKey is the one-line cloud API key setup hint.
	HintSetCloudAPIKey = hintSetCloudAPIKey
)

// ConfigSetAPIKeyCmd returns the ait config set command for a provider API key.
func ConfigSetAPIKeyCmd(id string) string {
	return fmt.Sprintf("ait config set ai.%s.api_key <key>", id)
}

// ConfigGetAPIKeyCmd returns the ait config get command for a provider API key.
func ConfigGetAPIKeyCmd(id string) string {
	return fmt.Sprintf("ait config get ai.%s.api_key", id)
}

// ProviderEnableCmd returns the ait ai provider enable command for id.
func ProviderEnableCmd(id string) string {
	return fmt.Sprintf("ait ai provider enable %s", id)
}
