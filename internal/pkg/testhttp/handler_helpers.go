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
)

func serveScenario(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Scenario == ScenarioHang {
			<-r.Context().Done()
			return
		}
		if cfg.Scenario == ScenarioServerError {
			http.Error(w, "mock server error", http.StatusInternalServerError)
			return
		}
		if cfg.Scenario == ScenarioNotFound {
			http.NotFound(w, r)
			return
		}

		switch r.URL.Path {
		case pathChatCompletions:
			writeChat(w, cfg)
		case pathAPITags:
			writeTags(w, cfg)
		case pathReverse:
			writeReverse(w, cfg)
		default:
			http.NotFound(w, r)
		}
	}
}

func writeChat(w http.ResponseWriter, cfg Config) {
	w.Header().Set("Content-Type", "application/json")
	switch cfg.Scenario {
	case ScenarioUnauthorized:
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"invalid api key"}}`)
	case ScenarioRateLimited:
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":{"message":"rate limited"}}`)
	case ScenarioMalformed:
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `not-json`)
	default:
		body, _ := json.Marshal(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": cfg.ChatReply}}},
		})
		_, _ = w.Write(body)
	}
}

func writeTags(w http.ResponseWriter, cfg Config) {
	w.Header().Set("Content-Type", "application/json")
	if cfg.Scenario == ScenarioMalformed {
		_, _ = io.WriteString(w, `{`)
		return
	}
	body, _ := json.Marshal(map[string]any{
		"models": []map[string]string{{"name": cfg.OllamaModel}},
	})
	_, _ = w.Write(body)
}

func writeReverse(w http.ResponseWriter, cfg Config) {
	w.Header().Set("Content-Type", "application/json")
	switch cfg.Scenario {
	case ScenarioGeocodeUnable:
		_, _ = io.WriteString(w, `{"error":"Unable to geocode"}`)
	case ScenarioMalformed:
		_, _ = io.WriteString(w, `not-json`)
	default:
		body, _ := json.Marshal(map[string]string{"display_name": cfg.DisplayName})
		_, _ = w.Write(body)
	}
}
