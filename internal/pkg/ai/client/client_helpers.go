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
	"net/http"
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/ollama"
)

// ProviderDisabledErr returns a user-visible error for a disabled provider override.
func ProviderDisabledErr(id string) error {
	return provider.DisabledErr(id)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	i := len(s) - max
	for i < len(s) && !utf8.RuneStart(s[i]) {
		i++
	}
	return s[i:]
}

// cloudKeyedSet returns the set of provider IDs that have a configured cloud endpoint.
func cloudKeyedSet(endpoints map[ProviderID]*endpoint) map[ProviderID]struct{} {
	out := make(map[ProviderID]struct{})
	for _, id := range config.CloudProviderIDs() {
		if _, ok := endpoints[id]; ok {
			out[id] = struct{}{}
		}
	}
	return out
}

// hasCloudEndpoint reports whether at least one cloud provider has a configured endpoint.
func hasCloudEndpoint(endpoints map[ProviderID]*endpoint) bool {
	for _, id := range config.CloudProviderIDs() {
		if _, ok := endpoints[id]; ok {
			return true
		}
	}
	return false
}

// initTieredEndpoint creates a local-tiered endpoint when the provider config exists.
func initTieredEndpoint(cfg *config.Config, localID string, httpClient *http.Client) *endpoint {
	pc, ok := cfg.Provider(localID)
	if !ok {
		return nil
	}
	return newEndpoint(localID, pc, ollamaPlaceholderKey, httpClient)
}

// providerReachable returns a display label for whether the provider is reachable.
func (c *Client) providerReachable(ctx context.Context, httpClient *http.Client, id ProviderID, enabled bool) string {
	if !enabled {
		return "no"
	}
	ep, ok := c.endpoints[id]
	if !ok {
		return "no"
	}
	if id == ProviderOllama {
		_, note, err := ollama.FetchTags(ctx, httpClient, ep.baseURL, ep.model)
		if err == nil && note == "" {
			return "yes"
		}
		return "no"
	}
	return "yes"
}
