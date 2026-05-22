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

package ollama

import (
	"github.com/os-artificer/AiT/internal/pkg/config"
)

// ApplyTieredConfig enables tiered routing and merges providers.ollama without overwriting cloud providers.
func ApplyTieredConfig(cfg *config.Config, model, baseURL string) error {
	if cfg == nil {
		return nil
	}
	if model == "" {
		model = config.DefaultOllamaModel
	}
	if baseURL == "" {
		baseURL = config.DefaultOllamaBase
	}

	cfg.AI.Tiered.Enabled = true
	enabled := true
	p, ok := cfg.Provider(config.ProviderOllama)
	if !ok {
		p = config.DefaultProviderConfig(config.ProviderOllama)
	}
	existingDir := p.ModelsDir
	p.Enabled = &enabled
	p.APIStyle = config.APIStyleOpenAI
	p.BaseURL = baseURL
	p.Model = model
	if existingDir != "" {
		p.ModelsDir = existingDir
	}
	if cfg.AI.Providers == nil {
		cfg.AI.Providers = make(map[string]config.ProviderConfig)
	}
	cfg.AI.Providers[config.ProviderOllama] = p
	return nil
}
