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

package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

func validateProviderBaseURL(providerID, apiStyle, raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}

	if providerID == ProviderOllama {
		return validateOllamaBaseURL(u)
	}

	if os.Getenv("AIT_ALLOW_CUSTOM_AI_URL") == "1" {
		if u.Scheme != "https" || u.Host == "" {
			return errors.New("base_url must be a valid https URL")
		}
		return nil
	}

	if u.Scheme != "https" {
		return errors.New("base_url must use https")
	}

	host := strings.ToLower(u.Hostname())
	switch providerID {
	case ProviderDeepseek:
		if host != "api.deepseek.com" {
			return errors.New("deepseek base_url host must be api.deepseek.com (set AIT_ALLOW_CUSTOM_AI_URL=1 to override)")
		}
	case ProviderMinimax:
		if host != "api.minimaxi.com" {
			return errors.New("minimax base_url host must be api.minimaxi.com (set AIT_ALLOW_CUSTOM_AI_URL=1 to override)")
		}
	default:
		return errors.New("unknown provider: " + providerID)
	}
	_ = apiStyle
	return nil
}

func validateOllamaBaseURL(u *url.URL) error {
	if os.Getenv("AIT_ALLOW_CUSTOM_AI_URL") == "1" {
		if u.Scheme != "http" && u.Scheme != "https" || u.Host == "" {
			return errors.New("ollama base_url must be a valid http(s) URL")
		}
		return nil
	}
	if u.Scheme != "http" {
		return errors.New("ollama base_url must use http")
	}
	if !isLocalhostHost(u.Hostname()) {
		return errors.New("ollama base_url host must be localhost or 127.0.0.1 (set AIT_ALLOW_CUSTOM_AI_URL=1 to override)")
	}
	return nil
}

func isLocalhostHost(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func defaultBaseURL(providerID, apiStyle string) string {
	switch providerID {
	case ProviderDeepseek:
		return DefaultDeepseekBase
	case ProviderMinimax:
		if apiStyle == APIStyleAnthropic {
			return DefaultMinimaxBaseAnthropic
		}
		return DefaultMinimaxBaseOpenAI
	case ProviderOllama:
		return DefaultOllamaBase
	default:
		return ""
	}
}

// DefaultProviderConfig returns built-in defaults for a known provider id.
func DefaultProviderConfig(id string) ProviderConfig {
	return defaultProviderConfig(id)
}

func defaultProviderConfig(id string) ProviderConfig {
	style := APIStyleOpenAI
	model := DefaultDeepseekModel
	switch id {
	case ProviderMinimax:
		model = DefaultMinimaxModel
	case ProviderOllama:
		model = DefaultOllamaModel
	}
	return ProviderConfig{
		APIStyle: style,
		BaseURL:  defaultBaseURL(id, style),
		Model:    model,
	}
}

func normalizeTiered(cfg *Config) {
	t := &cfg.AI.Tiered
	if t.LocalProvider == "" {
		t.LocalProvider = ProviderOllama
	}
	if t.MaxPromptChars <= 0 {
		t.MaxPromptChars = DefaultTieredMaxPromptChars
	}
	if t.MaxHistoryTurns < 0 {
		t.MaxHistoryTurns = DefaultTieredMaxHistoryTurns
	}
	if len(t.ComplexKeywords) == 0 {
		t.ComplexKeywords = append([]string(nil), DefaultTieredComplexKeywords...)
	}
}

func normalizeAI(cfg *Config) {
	ensureProviders(cfg)

	if cfg.Deepseek.BaseURL != "" || cfg.Deepseek.Model != "" {
		p := cfg.AI.Providers[ProviderDeepseek]
		if cfg.Deepseek.BaseURL != "" {
			p.BaseURL = cfg.Deepseek.BaseURL
		}
		if cfg.Deepseek.Model != "" {
			p.Model = cfg.Deepseek.Model
		}
		if p.APIStyle == "" {
			p.APIStyle = APIStyleOpenAI
		}
		cfg.AI.Providers[ProviderDeepseek] = p
	}

	for _, id := range AllProviderIDs() {
		p, ok := cfg.AI.Providers[id]
		if !ok {
			if id == ProviderOllama {
				continue
			}
			cfg.AI.Providers[id] = defaultProviderConfig(id)
			continue
		}
		if p.APIStyle == "" {
			p.APIStyle = APIStyleOpenAI
		}
		if p.BaseURL == "" {
			p.BaseURL = defaultBaseURL(id, p.APIStyle)
		}
		if p.Model == "" {
			p.Model = defaultProviderConfig(id).Model
		}
		cfg.AI.Providers[id] = p
	}

	if cfg.AI.Default == "" {
		cfg.AI.Default = ProviderDeepseek
	}
	if !IsCloudProvider(cfg.AI.Default) {
		cfg.AI.Default = ProviderDeepseek
	}

	normalizeTiered(cfg)

	if cfg.AI.Tiered.Enabled {
		if _, ok := cfg.AI.Providers[ProviderOllama]; !ok {
			cfg.AI.Providers[ProviderOllama] = defaultProviderConfig(ProviderOllama)
		}
	}

	cfg.AI.Agent.Normalize()
}

// clearLegacyDeepseekForSave zeroes the root deepseek block so Save does not duplicate ai.providers.deepseek.
func clearLegacyDeepseekForSave(cfg *Config) {
	if cfg == nil {
		return
	}
	cfg.Deepseek.BaseURL = ""
	cfg.Deepseek.Model = ""
}

func validateCloudDefault(id string) error {
	if !IsCloudProvider(id) {
		return fmt.Errorf("ai.default must be a cloud provider (%s or %s), not %q", ProviderDeepseek, ProviderMinimax, id)
	}
	return nil
}
