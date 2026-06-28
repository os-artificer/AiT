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

import (
	"github.com/os-artificer/AiT/internal/pkg/ai/apitypes"
	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
)

const (
	ProviderDeepseek = provider.Deepseek
	ProviderMinimax  = provider.Minimax
	ProviderOllama   = provider.Ollama

	ScopeLocal  = apitypes.ScopeLocal
	ScopeRemote = apitypes.ScopeRemote

	ollamaPlaceholderKey = "ollama"

	aiModeLocalFirst = "local_first"
	aiModeCloudOnly  = "cloud_only"

	errEmptyResponseFmt = "empty response from %s"
)

// Scope selects local versus remote AI mode context.
type Scope = apitypes.Scope

// ChatMessage is one turn in a multi-turn AI conversation.
type ChatMessage = apitypes.ChatMessage

// AssistEnv carries optional terminal context for the first user turn in a chat round.
type AssistEnv = apitypes.AssistEnv

// PromptMeta carries factual provider routing fields injected into the system prompt.
type PromptMeta = apitypes.PromptMeta

// ProviderID identifies an AI provider (alias of provider.ID).
type ProviderID = apitypes.ProviderID
