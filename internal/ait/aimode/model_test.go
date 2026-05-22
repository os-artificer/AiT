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

package aimode

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func TestSetModel_M12(t *testing.T) {
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_ALLOW_CUSTOM_AI_URL", "1")

	if err := creds.SetProviderAPIKey(config.ProviderMinimax, "k"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderMinimax: {APIStyle: config.APIStyleOpenAI, BaseURL: "http://127.0.0.1/v1", Model: "m"},
	}

	client, err := ai.NewClient(cfg, ai.WithProviderOverride(config.ProviderMinimax))
	if err != nil {
		t.Fatal(err)
	}

	var showBuf bytes.Buffer
	if err := setModel(context.Background(), Options{AI: client, Out: &showBuf}, "", &showBuf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(showBuf.String(), "minimax") {
		t.Fatalf("show: %q", showBuf.String())
	}

	var setBuf bytes.Buffer
	if err := setModel(context.Background(), Options{AI: client, Out: &setBuf}, config.ProviderMinimax, &setBuf); err != nil {
		t.Fatal(err)
	}
	if client.SessionProvider() != config.ProviderMinimax {
		t.Fatalf("session provider=%q", client.SessionProvider())
	}
}

func TestSetModel_ErrNoAI(t *testing.T) {
	var buf bytes.Buffer
	err := setModel(context.Background(), Options{Out: &buf}, "", &buf)
	if err != ErrAINotConfigured {
		t.Fatalf("got %v", err)
	}
}

func TestMain(m *testing.M) {
	_ = os.Setenv("AIT_KEYRING_BACKEND", "file")
	os.Exit(m.Run())
}
