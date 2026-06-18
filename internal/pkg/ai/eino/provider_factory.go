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
	"net/http"

	"github.com/os-artificer/AiT/internal/pkg/ai/apitypes"
	"github.com/os-artificer/AiT/internal/pkg/config"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

const ollamaPlaceholderKey = "ollama"

// NewProviderChatModel builds an Eino OpenAI-compatible ChatModel for provider id.
func NewProviderChatModel(
	ctx context.Context, id apitypes.ProviderID, pc config.ProviderConfig, apiKey string, httpClient *http.Client,
) (model.BaseChatModel, error) {
	if pc.APIStyle == config.APIStyleAnthropic {
		return nil, errLegacyAnthropic
	}

	key := apiKey
	if id == config.ProviderOllama {
		key = ollamaPlaceholderKey
	}

	cfg := &einoopenai.ChatModelConfig{
		APIKey:  key,
		BaseURL: pc.BaseURL,
		Model:   pc.Model,
	}
	if httpClient != nil {
		cfg.HTTPClient = httpClient
	}

	return einoopenai.NewChatModel(ctx, cfg)
}
