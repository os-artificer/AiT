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
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// MockChatModel is a test double for model.BaseChatModel / ToolCallingChatModel.
type MockChatModel struct {
	Reply     string
	Thinking  string
	ToolCalls []schema.ToolCall
}

// Generate returns a canned assistant message for tests.
func (m *MockChatModel) Generate(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	return &schema.Message{
		Role:             schema.Assistant,
		Content:          m.Reply,
		ReasoningContent: m.Thinking,
		ToolCalls:        m.ToolCalls,
	}, nil
}

// Stream is unused in AiT tests; satisfies ToolCallingChatModel when embedded.
func (m *MockChatModel) Stream(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

// WithTools returns the same mock as ToolCallingChatModel.
func (m *MockChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	_ = tools
	return m, nil
}
