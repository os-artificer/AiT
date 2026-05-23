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
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TagsBaseURL strips /v1 from an OpenAI-compatible Ollama base URL.
func TagsBaseURL(base string) string {
	return strings.TrimSuffix(strings.TrimSuffix(base, "/v1"), "/")
}

// FetchTags GETs /api/tags and returns tag lines, a user note on failure, or an error.
func FetchTags(ctx context.Context, client *http.Client, baseURL, configuredModel string) ([]string, string, error) {
	url := TagsBaseURL(baseURL) + pathAPITags
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, unreachableNote, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, unreachableNote, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, unreachableNote, fmt.Errorf("status %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", err
	}

	var parsed struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, "", err
	}

	var tags []string
	for _, m := range parsed.Models {
		line := m.Name
		if configuredModel != "" && (m.Name == configuredModel || strings.HasPrefix(m.Name, configuredModel+":")) {
			line += " * (configured)"
		}
		tags = append(tags, line)
	}
	return tags, "", nil
}
