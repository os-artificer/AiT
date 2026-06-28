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

// GenerateWithModel runs a single chat completion on the given model.
func GenerateWithModel(ctx context.Context, chatModel model.BaseChatModel, messages []*schema.Message) (reply, thinking string, err error) {
	return GenerateChat(ctx, chatModel, messages)
}

// GenerateAgentWithModel runs a single agent completion on the given model.
func GenerateAgentWithModel(ctx context.Context, chatModel model.BaseChatModel, messages []*schema.Message) (reply, thinking string, err error) {
	return GenerateAgent(ctx, chatModel, messages)
}
