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
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai/router"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func TestResolveTryOrder_SkipsDisabled(t *testing.T) {
	testAIEnv(t)
	cfg := testCfgDefaultDeepseek()
	if err := creds.SetProviderAPIKey(config.ProviderMinimax, "k-mini"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetProviderEnabled(config.ProviderMinimax, false); err != nil {
		t.Fatal(err)
	}

	keyed := map[ProviderID]struct{}{
		config.ProviderDeepseek: {},
	}
	order := router.ResolveTryOrder(cfg, "", keyed)
	for _, id := range order {
		if id == config.ProviderMinimax {
			t.Fatalf("order=%v should not include disabled minimax", order)
		}
	}
}

func TestNewClient_SkipsDisabledCloud(t *testing.T) {
	testAIEnv(t)
	cfg := testCfgDefaultDeepseek()
	if err := creds.SetProviderAPIKey(config.ProviderMinimax, "k"); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetKey("ai.default", config.ProviderMinimax); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetProviderEnabled(config.ProviderDeepseek, false); err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := client.slots[config.ProviderDeepseek]; ok {
		t.Fatal("disabled deepseek should not be registered")
	}
	if _, ok := client.slots[config.ProviderMinimax]; !ok {
		t.Fatal("enabled minimax should be registered")
	}
}
