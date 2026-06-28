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

// Package ai is the public facade for AiT AI features.
//
// The root package re-exports the client subpackage symbols used by the CLI and
// other internal packages, keeping implementation details in
// internal/pkg/ai/client.
package ai

import (
	"net/http"

	"github.com/os-artificer/AiT/internal/pkg/ai/client"
	"github.com/os-artificer/AiT/internal/pkg/ai/eino"
	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
	"github.com/os-artificer/AiT/internal/pkg/ai/router"
	"github.com/os-artificer/AiT/internal/pkg/config"
)

// ErrNotConfigured is returned when the AI client is not configured.
var ErrNotConfigured = provider.ErrNotConfigured

// ErrChatMessageRequired is returned when an empty chat prompt is supplied.
var ErrChatMessageRequired = provider.ErrChatMessageRequired

const (
	// ScopeLocal is local AI mode outside SSH.
	ScopeLocal = client.ScopeLocal
	// ScopeRemote is remote AI mode inside SSH.
	ScopeRemote = client.ScopeRemote
)

const (
	ProviderDeepseek = client.ProviderDeepseek
	ProviderMinimax  = client.ProviderMinimax
	ProviderOllama   = client.ProviderOllama
)

// Client is the shared AI client used by the rest of AiT.
type Client = client.Client

// ClientOption configures a new AI client instance.
type ClientOption = client.ClientOption

// Scope denotes local or remote AI mode.
type Scope = client.Scope

// ProviderID identifies an AI provider.
type ProviderID = client.ProviderID

// ChatMessage is one message in a conversation with AI.
type ChatMessage = client.ChatMessage

// AssistEnv carries optional context for assistant-style AI requests.
type AssistEnv = client.AssistEnv

// PromptMeta exposes routing metadata used in system prompts.
type PromptMeta = client.PromptMeta

// ModelRow is a single model entry in the AI model list.
type ModelRow = client.ModelRow

// ModelsReport is the formatted result of ListModels.
type ModelsReport = client.ModelsReport

// ExtractReply parses raw provider content into a clean user reply and optional thinking metadata.
func ExtractReply(raw string) (string, string) {
	return eino.ExtractReply(raw)
}

// WithProviderOverride sets the explicit cloud provider for a new AI client.
func WithProviderOverride(id ProviderID) ClientOption {
	return client.WithProviderOverride(id)
}

// WithHTTPClient configures the HTTP client used for provider requests.
func WithHTTPClient(c *http.Client) ClientOption {
	return client.WithHTTPClient(c)
}

// NewClient builds a multi-provider AI client with the given config.
func NewClient(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	return client.NewClient(cfg, opts...)
}

// ClassifyPrompt reports whether a prompt should use local tiered routing first.
func ClassifyPrompt(cfg *config.Config, user string, historyLen int, sessionOverride ProviderID, skipTiered bool) bool {
	return router.ClassifyPrompt(cfg, user, historyLen, sessionOverride, skipTiered)
}

// NotConfiguredErr returns a hint error when AI is not configured.
func NotConfiguredErr() error {
	return provider.NotConfiguredErr()
}

// ChatMessageRequiredErr returns a hint error for missing chat input.
func ChatMessageRequiredErr() error {
	return provider.ChatMessageRequiredErr()
}

// ProviderDisabledErr reports that a provider override is disabled.
func ProviderDisabledErr(id string) error {
	return provider.DisabledErr(id)
}
