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
	"net/http"

	"github.com/os-artificer/AiT/internal/pkg/config"

	openai "github.com/sashabaranov/go-openai"
)

type endpoint struct {
	id       ProviderID
	apiStyle string
	baseURL  string
	model    string
	key      string
	openai   *openai.Client
	http     *http.Client
}

func newEndpoint(id ProviderID, pc config.ProviderConfig, key string, httpClient *http.Client) *endpoint {
	style := pc.APIStyle
	if style == "" {
		style = config.APIStyleOpenAI
	}

	ep := &endpoint{
		id:       id,
		apiStyle: style,
		baseURL:  pc.BaseURL,
		model:    pc.Model,
		key:      key,
		http:     httpClient,
	}
	if style == config.APIStyleAnthropic {
		if ep.http == nil {
			ep.http = http.DefaultClient
		}
		return ep
	}

	ocfg := openai.DefaultConfig(key)
	ocfg.BaseURL = pc.BaseURL
	if httpClient != nil {
		ocfg.HTTPClient = httpClient
	}
	ep.openai = openai.NewClientWithConfig(ocfg)
	return ep
}
