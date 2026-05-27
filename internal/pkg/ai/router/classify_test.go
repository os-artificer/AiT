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

package router

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

func tieredCfg() *config.Config {
	cfg := &config.Config{}
	cfg.AI.Tiered.Enabled = true
	cfg.AI.Tiered.MaxPromptChars = 280
	cfg.AI.Tiered.MaxHistoryTurns = 0
	cfg.AI.Tiered.ComplexKeywords = []string{"implement", "debug"}
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderOllama: {Model: "llama3.2"},
	}
	return cfg
}

func TestClassifyPrompt_H1_Disabled(t *testing.T) {
	cfg := tieredCfg()
	cfg.AI.Tiered.Enabled = false
	if ClassifyPrompt(cfg, "hi", 0, "", false) {
		t.Fatal("want false when tiered disabled")
	}
}

func TestClassifyPrompt_H2_Simple(t *testing.T) {
	if !ClassifyPrompt(tieredCfg(), "what is ls?", 0, "", false) {
		t.Fatal("want true for short simple prompt")
	}
}

func TestClassifyPrompt_H3_Long(t *testing.T) {
	long := strings.Repeat("a", 300)
	if ClassifyPrompt(tieredCfg(), long, 0, "", false) {
		t.Fatal("want false for long prompt")
	}
}

func TestClassifyPrompt_H4_History(t *testing.T) {
	if ClassifyPrompt(tieredCfg(), "hi", 1, "", false) {
		t.Fatal("want false when history exceeds max_history_turns")
	}
}

func TestClassifyPrompt_H5_CodeFence(t *testing.T) {
	if ClassifyPrompt(tieredCfg(), "fix ```go\nmain\n```", 0, "", false) {
		t.Fatal("want false for code fence")
	}
}

func TestClassifyPrompt_H6_Keyword(t *testing.T) {
	if ClassifyPrompt(tieredCfg(), "please implement a cache", 0, "", false) {
		t.Fatal("want false for complex keyword")
	}
}

func TestClassifyPrompt_H7_SessionOverride(t *testing.T) {
	if ClassifyPrompt(tieredCfg(), "hi", 0, config.ProviderMinimax, false) {
		t.Fatal("want false when session override set")
	}
}

func TestClassifyPrompt_OllamaDisabled(t *testing.T) {
	cfg := tieredCfg()
	disabled := false
	p := cfg.AI.Providers[config.ProviderOllama]
	p.Enabled = &disabled
	cfg.AI.Providers[config.ProviderOllama] = p
	if ClassifyPrompt(cfg, "hi", 0, "", false) {
		t.Fatal("want false when ollama provider disabled")
	}
}
