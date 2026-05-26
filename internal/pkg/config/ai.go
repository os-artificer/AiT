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

import "fmt"

// DefaultTieredComplexKeywords are substrings that force cloud routing when tiered routing is on.
var DefaultTieredComplexKeywords = []string{
	"implement", "refactor", "architecture", "debug", "架构", "重构", "实现",
}

// Provider identifiers for ai.providers and ai.default.
const (
	ProviderDeepseek = "deepseek"
	ProviderMinimax  = "minimax"
	ProviderOllama   = "ollama"
)

// APIStyle selects the HTTP protocol for a provider endpoint.
const (
	APIStyleOpenAI               = "openai"
	APIStyleAnthropic            = "anthropic"
	DefaultMinimaxModel          = "MiniMax-M2.5"
	DefaultMinimaxBaseOpenAI     = "https://api.minimaxi.com/v1"
	DefaultMinimaxBaseAnthropic  = "https://api.minimaxi.com/anthropic"
	DefaultOllamaModel           = "llama3.2"
	DefaultOllamaBase            = "http://127.0.0.1:11434/v1"
	DefaultTieredMaxPromptChars  = 280
	DefaultTieredMaxHistoryTurns = 0
)

// ProviderConfig holds per-provider AI endpoint settings.
type ProviderConfig struct {
	Enabled   *bool  `yaml:"enabled,omitempty"`
	APIStyle  string `yaml:"api_style"`
	BaseURL   string `yaml:"base_url"`
	Model     string `yaml:"model"`
	ModelsDir string `yaml:"models_dir,omitempty"` // Ollama only: host path for OLLAMA_MODELS
}

// TieredConfig controls local-first routing to Ollama for simple prompts.
type TieredConfig struct {
	Enabled         bool     `yaml:"enabled"`
	LocalProvider   string   `yaml:"local_provider"`
	MaxPromptChars  int      `yaml:"max_prompt_chars"`
	MaxHistoryTurns int      `yaml:"max_history_turns"`
	ComplexKeywords []string `yaml:"complex_keywords"`
}

// AIConfig groups default provider, optional fallback order, tiered routing, and provider map.
type AIConfig struct {
	Default   string                    `yaml:"default"`
	Fallback  []string                  `yaml:"fallback"`
	Tiered    TieredConfig              `yaml:"tiered"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// CloudProviderIDs returns cloud providers used for ai.default, fallback, and --model.
func CloudProviderIDs() []string {
	return []string{ProviderDeepseek, ProviderMinimax}
}

// AllProviderIDs returns every built-in provider ID including local Ollama.
func AllProviderIDs() []string {
	return []string{ProviderDeepseek, ProviderMinimax, ProviderOllama}
}

// IsCloudProvider reports whether id may be ai.default or session --model override.
func IsCloudProvider(id string) bool {
	return id == ProviderDeepseek || id == ProviderMinimax
}

// IsKnownProvider reports whether id is a built-in provider (including ollama).
func IsKnownProvider(id string) bool {
	return IsCloudProvider(id) || id == ProviderOllama
}

// LocalProviderID returns the tiered local provider ID (default ollama).
func (c *Config) LocalProviderID() string {
	if c == nil || c.AI.Tiered.LocalProvider == "" {
		return ProviderOllama
	}
	return c.AI.Tiered.LocalProvider
}

// TieredEnabled reports whether local-first routing is active.
func (c *Config) TieredEnabled() bool {
	return c != nil && c.AI.Tiered.Enabled
}

// Provider returns config for id, or false if unknown.
func (c *Config) Provider(id string) (ProviderConfig, bool) {
	if c == nil || c.AI.Providers == nil {
		return ProviderConfig{}, false
	}
	p, ok := c.AI.Providers[id]
	return p, ok
}

// ProviderEnabled reports whether id is configured and not explicitly disabled.
// Omitted enabled in YAML is treated as true.
func (c *Config) ProviderEnabled(id string) bool {
	p, ok := c.Provider(id)
	if !ok {
		return false
	}
	if p.Enabled == nil {
		return true
	}
	return *p.Enabled
}

// SetProviderEnabled sets ai.providers.<id>.enabled and ensures the provider entry exists.
func (c *Config) SetProviderEnabled(id string, enabled bool) error {
	if !IsKnownProvider(id) {
		return fmt.Errorf("unknown provider %q", id)
	}
	ensureProviders(c)
	p, ok := c.AI.Providers[id]
	if !ok {
		p = defaultProviderConfig(id)
	}
	p.Enabled = &enabled
	c.AI.Providers[id] = p
	return nil
}

// DefaultProviderID returns ai.default or ProviderDeepseek when unset.
func (c *Config) DefaultProviderID() string {
	if c == nil || c.AI.Default == "" {
		return ProviderDeepseek
	}
	return c.AI.Default
}
