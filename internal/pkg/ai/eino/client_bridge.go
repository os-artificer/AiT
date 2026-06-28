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

// ChatModel wraps an Eino BaseChatModel for Ask/Agent single-step calls.
type ChatModel struct {
	model.BaseChatModel
}

// GenerateChat delegates to package GenerateChat.
func (m *ChatModel) GenerateChat(ctx context.Context, messages []*schema.Message) (reply, thinking string, err error) {
	return GenerateChat(ctx, m.BaseChatModel, messages)
}

// GenerateAgent delegates to package GenerateAgent.
func (m *ChatModel) GenerateAgent(ctx context.Context, messages []*schema.Message) (reply, thinking string, err error) {
	return GenerateAgent(ctx, m.BaseChatModel, messages)
}
