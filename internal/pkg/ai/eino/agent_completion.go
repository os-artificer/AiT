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

// GenerateChat runs a single Ask turn via Eino and returns clean reply + thinking (E9: thinking not for history).
func GenerateChat(ctx context.Context, chatModel model.BaseChatModel, messages []*schema.Message) (reply, thinking string, err error) {
	out, err := chatModel.Generate(ctx, messages)
	if err != nil {
		return "", "", err
	}
	reply, thinking = ExtractReply(out.Content)
	if rc := out.ReasoningContent; rc != "" {
		thinking = rc
	}
	return reply, thinking, nil
}

// GenerateAgent runs a single Agent turn via Eino with tools bound.
func GenerateAgent(ctx context.Context, chatModel model.BaseChatModel, messages []*schema.Message) (reply, thinking string, err error) {
	tools, err := AgentToolInfos()
	if err != nil {
		return "", "", err
	}

	var agentModel model.BaseChatModel = chatModel
	if tc, ok := chatModel.(model.ToolCallingChatModel); ok {
		withTools, bindErr := tc.WithTools(tools)
		if bindErr != nil {
			return "", "", bindErr
		}
		agentModel = withTools
	}

	out, err := agentModel.Generate(ctx, messages)
	if err != nil {
		return "", "", err
	}
	return ProcessAgentMessage(out)
}
