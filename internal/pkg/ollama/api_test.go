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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"models":[{"name":"llama3.2"}]}`))
	}))
	defer srv.Close()

	tags, note, err := FetchTags(context.Background(), srv.Client(), srv.URL+"/v1", "llama3.2")
	if err != nil || note != "" {
		t.Fatalf("err=%v note=%q", err, note)
	}
	if len(tags) != 1 || !strings.Contains(tags[0], "configured") {
		t.Fatalf("tags=%v", tags)
	}
}

func TestTagsBaseURL(t *testing.T) {
	if got := TagsBaseURL("http://127.0.0.1:11434/v1"); got != "http://127.0.0.1:11434" {
		t.Fatalf("got %q", got)
	}
}
