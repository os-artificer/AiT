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

// Package apitypes holds shared AI DTOs used by client and eino without import cycles.
package apitypes

import "github.com/os-artificer/AiT/internal/pkg/ai/provider"

const (
	// ScopeLocal is aitssh ai on the local machine (non-SSH).
	ScopeLocal Scope = iota
	// ScopeRemote is !ait ai inside an active SSH session.
	ScopeRemote
)

// Scope selects local versus remote AI mode context.
type Scope int

// ProviderID identifies an AI provider.
type ProviderID = provider.ID

// ChatMessage is one turn in a multi-turn AI conversation.
type ChatMessage struct {
	Role    string
	Content string
}

// AssistEnv carries optional terminal context for the first user turn in a chat round.
type AssistEnv struct {
	Scope       Scope
	HostLabel   string
	RemoteCwd   string
	Scrollback  string
	AgentMode   bool
	ApproveEach bool
}

// PromptMeta carries factual provider routing fields injected into the system prompt.
type PromptMeta struct {
	Provider ProviderID
	Model    string
	Mode     string
}
