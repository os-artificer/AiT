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
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func yamlBytesForSave(cfg *Config) ([]byte, error) {
	normalizeAI(cfg)
	clearLegacyDeepseekForSave(cfg)
	return yaml.Marshal(cfg)
}

func hasRootDeepseekKey(yamlText string) bool {
	for _, line := range strings.Split(yamlText, "\n") {
		if len(line) > 0 && line[0] == 'd' && strings.HasPrefix(line, "deepseek:") {
			return true
		}
	}
	return false
}

func TestSaveYAML_NoRootDeepseek(t *testing.T) {
	cfg := defaults()
	cfg.AI.Fallback = []string{ProviderMinimax}

	raw, err := yamlBytesForSave(cfg)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if hasRootDeepseekKey(text) {
		t.Fatalf("root deepseek should be omitted:\n%s", text)
	}
	if !strings.Contains(text, "providers:") || !strings.Contains(text, "deepseek") {
		t.Fatalf("expected ai.providers.deepseek in:\n%s", text)
	}
}

func TestLegacyDeepseekYAML_LoadThenSave(t *testing.T) {
	var cfg Config
	const legacy = "session:\n  auth_ttl: 168h\ndeepseek:\n  model: legacy-model\n  base_url: https://api.deepseek.com\n"
	if err := yaml.Unmarshal([]byte(legacy), &cfg); err != nil {
		t.Fatal(err)
	}
	normalizeAI(&cfg)

	p, ok := cfg.Provider(ProviderDeepseek)
	if !ok || p.Model != "legacy-model" {
		t.Fatalf("provider=%+v ok=%v", p, ok)
	}

	raw, err := yamlBytesForSave(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if hasRootDeepseekKey(text) {
		t.Fatalf("save yaml must not contain root deepseek:\n%s", text)
	}
	p, ok = cfg.Provider(ProviderDeepseek)
	if !ok || p.Model != "legacy-model" {
		t.Fatalf("provider after save prep=%+v ok=%v", p, ok)
	}
}
