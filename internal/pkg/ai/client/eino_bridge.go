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
	"context"
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/ai/eino"
	"github.com/os-artificer/AiT/internal/pkg/config"

	"github.com/cloudwego/eino/components/model"
)

func (c *Client) completeEino(ctx context.Context, slot *providerSlot, messages []completionMessage, agent bool) (string, string, error) {
	chatModel, err := c.newEinoChatModel(ctx, slot)
	if err != nil {
		return "", "", err
	}

	schemaMsgs := eino.SchemaFromPlain(completionToPlain(messages))
	var reply, thinking string
	if agent {
		reply, thinking, err = eino.GenerateAgent(ctx, chatModel, schemaMsgs)
	} else {
		reply, thinking, err = eino.GenerateChat(ctx, chatModel, schemaMsgs)
	}
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", slot.id, err)
	}
	if reply == "" && thinking == "" {
		return "", "", fmt.Errorf(errEmptyResponseFmt, slot.id)
	}
	return reply, thinking, nil
}

func (c *Client) newEinoChatModel(ctx context.Context, slot *providerSlot) (model.BaseChatModel, error) {
	pc := config.ProviderConfig{
		APIStyle: slot.apiStyle,
		BaseURL:  slot.baseURL,
		Model:    slot.model,
	}
	return eino.NewProviderChatModel(ctx, slot.id, pc, slot.key, c.httpClient)
}

func completionToPlain(msgs []completionMessage) []eino.PlainMessage {
	out := make([]eino.PlainMessage, len(msgs))
	for i, m := range msgs {
		out[i] = eino.PlainMessage{Role: m.Role, Content: m.Content}
	}
	return out
}
