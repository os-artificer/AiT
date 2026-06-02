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
	"strings"
)

// EnsureModel pulls model when missing from /api/tags.
func EnsureModel(ctx context.Context, client *http.Client, baseURL, model string, skipPull bool, modelsDir string) error {
	if skipPull || model == "" {
		return nil
	}
	if client == nil {
		client = &http.Client{Timeout: httpClientTimeout}
	}
	tags, _, err := FetchTags(ctx, client, baseURL, model)
	if err != nil {
		return err
	}
	if modelPresent(tags, model) {
		return nil
	}
	return runOllama(ctx, modelsDir, "pull", model)
}

func modelPresent(tags []string, model string) bool {
	for _, t := range tags {
		name := t
		if i := strings.Index(t, " *"); i >= 0 {
			name = t[:i]
		}
		if name == model || strings.HasPrefix(name, model+":") {
			return true
		}
	}
	return false
}
