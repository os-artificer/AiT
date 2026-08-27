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
	"context"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/testhttp"
)

func TestAssist_MockHTTP_OK(t *testing.T) {
	c := newMockHTTPClient(t, testhttp.Config{Scenario: testhttp.ScenarioOK, ChatReply: "mock-chat-ok"})
	reply, err := c.Assist(context.Background(), "ping", "", "")
	if err != nil || reply != "mock-chat-ok" {
		t.Fatalf("reply=%q err=%v", reply, err)
	}
}

func TestAssist_MockHTTP_Unauthorized(t *testing.T) {
	c := newMockHTTPClient(t, testhttp.Config{Scenario: testhttp.ScenarioUnauthorized})
	_, err := c.Assist(context.Background(), "ping", "", "")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestAssist_MockHTTP_RateLimited(t *testing.T) {
	c := newMockHTTPClient(t, testhttp.Config{Scenario: testhttp.ScenarioRateLimited})
	_, err := c.Assist(context.Background(), "ping", "", "")
	if err == nil {
		t.Fatal("expected rate-limit error")
	}
}

func newMockHTTPClient(t *testing.T, httpCfg testhttp.Config) *Client {
	t.Helper()
	testAIEnv(t)

	srv := testhttp.Start(t, httpCfg)
	cfg := &config.Config{}
	cfg.AI.Default = config.ProviderDeepseek
	cfg.AI.Providers = map[string]config.ProviderConfig{
		config.ProviderDeepseek: {
			APIStyle: config.APIStyleOpenAI,
			BaseURL:  srv.ChatURL(),
			Model:    "deepseek-chat",
		},
	}
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "test-key"); err != nil {
		t.Fatal(err)
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
