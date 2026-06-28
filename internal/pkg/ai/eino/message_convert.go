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

package eino

import (
	"github.com/os-artificer/AiT/internal/pkg/ai/apitypes"

	"github.com/cloudwego/eino/schema"
)

const (
	// RoleSystem is the system message role string.
	RoleSystem = "system"
	// RoleUser is the user message role string.
	RoleUser = "user"
	// RoleAssistant is the assistant message role string.
	RoleAssistant = "assistant"
)

// PlainMessage is a provider-neutral chat message (role + content).
type PlainMessage struct {
	Role    string
	Content string
}

// PlainFromSchema converts Eino schema messages to plain role/content pairs.
func PlainFromSchema(msgs []*schema.Message) []PlainMessage {
	out := make([]PlainMessage, 0, len(msgs))
	for _, m := range msgs {
		role := RoleUser
		switch m.Role {
		case schema.System:
			role = RoleSystem
		case schema.Assistant:
			role = RoleAssistant
		}
		out = append(out, PlainMessage{Role: role, Content: m.Content})
	}
	return out
}

// BuildPlainChatMessages constructs plain role/content pairs for a chat round.
func BuildPlainChatMessages(scope apitypes.Scope, history []apitypes.ChatMessage, user string, env apitypes.AssistEnv, meta apitypes.PromptMeta) []PlainMessage {
	return PlainFromSchema(BuildChatMessages(scope, history, user, env, meta))
}

// SchemaFromPlain converts plain messages to Eino schema messages.
func SchemaFromPlain(msgs []PlainMessage) []*schema.Message {
	out := make([]*schema.Message, 0, len(msgs))
	for _, m := range msgs {
		switch m.Role {
		case RoleSystem:
			out = append(out, schema.SystemMessage(m.Content))
		case RoleAssistant:
			out = append(out, schema.AssistantMessage(m.Content, nil))
		default:
			out = append(out, schema.UserMessage(m.Content))
		}
	}
	return out
}
