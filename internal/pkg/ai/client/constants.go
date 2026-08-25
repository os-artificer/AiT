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

package client

import "time"

const (
	errEmptyPrompt         = "empty prompt"
	errClientNotConfigured = "ai client not configured"

	aiModeTieredLabel = "tiered (simple prompts local)"

	anthropicMessagesPath      = "/v1/messages"
	anthropicMaxTokens         = 4096
	anthropicHeaderContentType = "Content-Type"
	anthropicContentType       = "application/json"
	anthropicHeaderAPIKey      = "x-api-key"
	anthropicHeaderVersion     = "anthropic-version"

	anthropicBlockThinking         = "thinking"
	anthropicBlockRedactedThinking = "redacted_thinking"
	anthropicBlockText             = "text"

	anthropicRoleUser      = "user"
	anthropicRoleAssistant = "assistant"

	openAIChatCompletionsPath = "/chat/completions"
	openAIHeaderContentType   = "Content-Type"
	openAIContentType         = "application/json"
	openAIHeaderAuthorization = "Authorization"
	openAIBearerPrefix        = "Bearer "

	modelsListHTTPTimeout = 3 * time.Second

	httpStatusTooManyRequests = 429
	httpStatusInternalError   = 500

	statusAIProviderPrefix = "AI provider: "
	statusTieredOnSuffix   = ", tiered=on (simple prompts use local ollama)"
	statusSessionOverride  = ", session override: %s"

	hintNoProvidersConfigured = "Set cloud API keys or enable tiered Ollama routing."
	hintNoCloudAPIKeys        = "Set an API key for a cloud provider."

	errNoCloudProvidersWithKeys = "no enabled cloud AI providers with API keys configured"
	errCannotSetOllamaSession   = "cannot set session provider to ollama (local routing is automatic when tiered is enabled)"
	errUnknownCloudProviderFmt  = "unknown cloud provider %q"
	errProviderDisabledFmt      = "provider %q is disabled (ai provider enable %s)"
	errProviderNoAPIKeyFmt      = "provider %q has no API key (%s)"
	errAPIKeyNotSetFmt          = "%s API key not set"
	errAPIKeyNotSetWithHintFmt  = "%s API key not set (%s)"
	errProviderOverrideCloudFmt = "provider override must be a cloud provider (%s or %s)"
)
