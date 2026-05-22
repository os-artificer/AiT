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
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

func TestStartServe_AlreadyRunning(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"models": []any{}})
	}))
	defer srv.Close()

	err := StartServe(context.Background(), srv.Client(), srv.URL+"/v1", false, "")
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("got %v", err)
	}
}

func TestDisableProvider_SetsEnabledFalse(t *testing.T) {
	cfg := &config.Config{}
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderOllama: {Model: config.DefaultOllamaModel},
	}
	if err := DisableProvider(context.Background(), cfg, false); err != nil {
		t.Fatal(err)
	}
	if cfg.ProviderEnabled(config.ProviderOllama) {
		t.Fatal("expected ollama disabled")
	}
}

func TestBaseURLFromConfig_Default(t *testing.T) {
	if got := BaseURLFromConfig(nil); got != config.DefaultOllamaBase {
		t.Fatalf("got %q", got)
	}
}
