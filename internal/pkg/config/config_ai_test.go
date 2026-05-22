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
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNormalizeAI_LegacyDeepseekM5(t *testing.T) {
	cfg := &Config{}
	cfg.Deepseek.BaseURL = "https://api.deepseek.com"
	cfg.Deepseek.Model = "deepseek-chat"
	normalizeAI(cfg)

	if cfg.DefaultProviderID() != ProviderDeepseek {
		t.Fatalf("default=%q", cfg.DefaultProviderID())
	}
	p, ok := cfg.Provider(ProviderDeepseek)
	if !ok || p.Model != "deepseek-chat" {
		t.Fatalf("provider=%+v ok=%v", p, ok)
	}
}

func TestValidateMinimaxURLM6(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	if err := cfg.SetKey("minimax.base_url", "https://api.minimax.io/v1"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetKey("minimax.base_url", "https://evil.example/v1"); err == nil {
		t.Fatal("expected reject illegal minimax host")
	}
}

func TestSetGetAIDefaultFallbackP1P2(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)

	if err := cfg.SetKey("ai.default", ProviderMinimax); err != nil {
		t.Fatal(err)
	}
	v, err := cfg.GetKey("ai.default")
	if err != nil || v != ProviderMinimax {
		t.Fatalf("default=%q err=%v", v, err)
	}

	if err := cfg.SetKey("ai.fallback", "minimax,deepseek"); err != nil {
		t.Fatal(err)
	}
	fb, err := cfg.GetKey("ai.fallback")
	if err != nil || fb != "minimax,deepseek" {
		t.Fatalf("fallback=%q err=%v", fb, err)
	}
}

func TestUnmarshalLegacyYAMLM5(t *testing.T) {
	var cfg Config
	if err := yaml.Unmarshal([]byte("deepseek:\n  model: legacy-model\n"), &cfg); err != nil {
		t.Fatal(err)
	}
	normalizeAI(&cfg)
	p, ok := cfg.Provider(ProviderDeepseek)
	if !ok || p.Model != "legacy-model" {
		t.Fatalf("model=%q ok=%v", p.Model, ok)
	}
	if cfg.DefaultProviderID() != ProviderDeepseek {
		t.Fatalf("default=%q", cfg.DefaultProviderID())
	}
}

func TestValidateOllamaURLT5(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	if err := cfg.SetKey("ollama.base_url", "http://127.0.0.1:11434/v1"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetKey("ollama.base_url", "http://evil.example:11434/v1"); err == nil {
		t.Fatal("expected reject non-localhost ollama host")
	}
}

func TestAIDefaultRejectsOllamaC1(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	if err := cfg.SetKey("ai.default", ProviderOllama); err == nil {
		t.Fatal("expected reject ollama as ai.default")
	}
}

func TestOllamaModelsDirSetGet(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	dir := filepath.Join(t.TempDir(), "ollama-models")
	if err := cfg.SetKey("ollama.models_dir", dir); err != nil {
		t.Fatal(err)
	}
	got, err := cfg.GetKey("ai.providers.ollama.models_dir")
	if err != nil || got != dir {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestProviderEnabledSetGet(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	if err := cfg.SetKey("ai.providers.deepseek.enabled", "false"); err != nil {
		t.Fatal(err)
	}
	if cfg.ProviderEnabled(ProviderDeepseek) {
		t.Fatal("expected deepseek disabled")
	}
	v, err := cfg.GetKey("ai.providers.deepseek.enabled")
	if err != nil || v != "false" {
		t.Fatalf("enabled=%q err=%v", v, err)
	}
	if err := cfg.SetProviderEnabled(ProviderMinimax, true); err != nil {
		t.Fatal(err)
	}
	if !cfg.ProviderEnabled(ProviderMinimax) {
		t.Fatal("expected minimax enabled after SetProviderEnabled")
	}
}

func TestTieredKeysC5(t *testing.T) {
	cfg := &Config{}
	normalizeAI(cfg)
	if err := cfg.SetKey("ai.tiered.enabled", "true"); err != nil {
		t.Fatal(err)
	}
	v, err := cfg.GetKey("ai.tiered.enabled")
	if err != nil || v != "true" {
		t.Fatalf("enabled=%q err=%v", v, err)
	}
}

func TestYAMLRoundTripAIBlock(t *testing.T) {
	cfg := defaults()
	cfg.AI.Fallback = []string{ProviderMinimax}
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Config
	if err := yaml.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	normalizeAI(&decoded)
	if decoded.AI.Default != ProviderDeepseek {
		t.Fatalf("default=%q", decoded.AI.Default)
	}
}
