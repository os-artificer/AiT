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

package search

import (
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func visionEndpoint(cfg *config.Config) (baseURL, apiKey string) {
	if cfg == nil {
		return "", ""
	}
	id := cfg.AI.Default
	if id == "" {
		id = config.ProviderDeepseek
	}
	p, ok := cfg.Provider(id)
	if !ok {
		return "", ""
	}
	baseURL = p.BaseURL
	if k, err := creds.GetProviderAPIKey(id); err == nil {
		apiKey = k
	}
	return baseURL, apiKey
}
