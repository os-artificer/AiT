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
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/testhttp"
)

func TestFetchTags_MockOK(t *testing.T) {
	srv := testhttp.Start(t, testhttp.Config{Scenario: testhttp.ScenarioOK, OllamaModel: "llama3.2:latest"})
	tags, note, err := FetchTags(context.Background(), srv.HTTPClient(), srv.BaseURL(), "llama3.2")
	if err != nil {
		t.Fatal(err)
	}
	if note != "" {
		t.Fatalf("note=%q", note)
	}
	if len(tags) != 1 || !strings.Contains(tags[0], "llama3.2:latest") {
		t.Fatalf("tags=%v", tags)
	}
}

func TestFetchTags_MockServerError(t *testing.T) {
	srv := testhttp.Start(t, testhttp.Config{Scenario: testhttp.ScenarioServerError})
	_, note, err := FetchTags(context.Background(), srv.HTTPClient(), srv.BaseURL(), "")
	if err == nil {
		t.Fatal("expected error")
	}
	if note == "" {
		t.Fatal("expected unreachable note")
	}
}

func TestFetchTags_MockMalformed(t *testing.T) {
	srv := testhttp.Start(t, testhttp.Config{Scenario: testhttp.ScenarioMalformed})
	_, _, err := FetchTags(context.Background(), srv.HTTPClient(), srv.BaseURL(), "")
	if err == nil {
		t.Fatal("expected JSON error")
	}
}
