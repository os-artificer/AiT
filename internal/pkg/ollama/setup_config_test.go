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
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

func TestApplyTieredConfig_OCFG(t *testing.T) {
	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderMinimax
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {Model: "deepseek-chat"},
		config.ProviderMinimax:  {Model: "MiniMax-M2.5"},
	}

	if err := ApplyTieredConfig(cfg, "qwen2.5", "http://127.0.0.1:11434/v1"); err != nil {
		t.Fatal(err)
	}
	if !cfg.TieredEnabled() {
		t.Fatal("tiered should be enabled")
	}
	pc, ok := cfg.Provider(config.ProviderOllama)
	if !ok || pc.Model != "qwen2.5" || !cfg.ProviderEnabled(config.ProviderOllama) {
		t.Fatalf("ollama=%+v enabled=%v", pc, cfg.ProviderEnabled(config.ProviderOllama))
	}
	if cfg.DefaultProviderID() != config.ProviderMinimax {
		t.Fatalf("default changed to %q", cfg.DefaultProviderID())
	}
	if ds, _ := cfg.Provider(config.ProviderDeepseek); ds.Model != "deepseek-chat" {
		t.Fatal("deepseek provider overwritten")
	}
}
