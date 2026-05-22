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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

func TestBuildStatus_TieredAndAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]string{{"name": "llama3.2"}},
		})
	}))
	defer srv.Close()

	cfg := &config.Config{}
	cfg.AI.Tiered.Enabled = true
	disabled := false
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderOllama: {
			BaseURL: srv.URL + "/v1",
			Model:   "llama3.2",
			Enabled: &disabled,
		},
	}

	report := BuildStatus(context.Background(), cfg, srv.Client())
	if !report.TieredEnabled {
		t.Fatal("expected tiered enabled")
	}
	if report.ProviderEnabled {
		t.Fatal("expected provider disabled")
	}
	if !report.APIOK {
		t.Fatal("expected API ok")
	}
	if len(report.PulledModels) == 0 {
		t.Fatal("expected pulled models")
	}
}

func TestPrintStatus_Format(t *testing.T) {
	var buf bytes.Buffer
	cfg := defaultsStatusCfg()
	if err := PrintStatus(context.Background(), cfg, &buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, sub := range []string{"Ollama status:", "ai.tiered.enabled", "ai.providers.ollama.enabled"} {
		if !strings.Contains(out, sub) {
			t.Errorf("missing %q in %q", sub, out)
		}
	}
}

func defaultsStatusCfg() *config.Config {
	cfg := &config.Config{}
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderOllama: {BaseURL: config.DefaultOllamaBase, Model: config.DefaultOllamaModel},
	}
	return cfg
}
