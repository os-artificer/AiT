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

package testhttp

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTP_OpenAIChatOK(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioOK, ChatReply: "pong"})
	resp, err := http.Post(srv.BaseURL()+pathChatCompletions, "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "pong") {
		t.Fatalf("body=%s", body)
	}
}

func TestHTTP_OpenAIUnauthorized(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioUnauthorized})
	resp, err := http.Post(srv.BaseURL()+pathChatCompletions, "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestHTTP_OpenAIRateLimited(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioRateLimited})
	resp, err := http.Post(srv.BaseURL()+pathChatCompletions, "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestHTTP_OllamaTagsOK(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioOK, OllamaModel: "qwen2.5:7b"})
	resp, err := http.Get(srv.BaseURL() + pathAPITags)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var parsed struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Models) != 1 || parsed.Models[0].Name != "qwen2.5:7b" {
		t.Fatalf("parsed=%+v", parsed)
	}
}

func TestHTTP_ServerError(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioServerError})
	resp, err := http.Get(srv.ReverseURL())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestHTTP_NominatimOK(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioOK})
	resp, err := http.Get(srv.ReverseURL() + "?lat=31.24&lon=121.49")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "外滩") {
		t.Fatalf("body=%s", body)
	}
}

func TestHTTP_NominatimUnable(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioGeocodeUnable})
	resp, err := http.Get(srv.ReverseURL())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Unable to geocode") {
		t.Fatalf("body=%s", body)
	}
}

func TestHTTP_MalformedJSON(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioMalformed})
	resp, err := http.Post(srv.BaseURL()+pathChatCompletions, "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if json.Valid(body) {
		t.Fatalf("expected invalid JSON, got %s", body)
	}
}

func TestHTTP_NotFound(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioNotFound})
	resp, err := http.Get(srv.BaseURL() + pathAPITags)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status=%d", resp.StatusCode)
	}
}

func TestHTTP_HangTimesOut(t *testing.T) {
	srv := Start(t, Config{Scenario: ScenarioHang})
	client := &http.Client{Timeout: 200 * time.Millisecond}
	_, err := client.Get(srv.BaseURL() + pathAPITags)
	if err == nil {
		t.Fatal("expected client timeout against hang scenario")
	}
}
