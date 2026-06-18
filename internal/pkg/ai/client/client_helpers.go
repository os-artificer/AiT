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

	"github.com/os-artificer/AiT/internal/pkg/ai/models"
	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/ollama"
	"github.com/os-artificer/AiT/internal/pkg/utf8tail"
)

// ProviderDisabledErr returns a user-visible error for a disabled provider override.
func ProviderDisabledErr(id string) error {
	return provider.DisabledErr(id)
}

func truncate(s string, max int) string {
	return utf8tail.Truncate(s, max)
}

// cloudKeyedSet returns the set of cloud provider IDs with configured slots.
func cloudKeyedSet(slots map[ProviderID]*providerSlot) map[ProviderID]struct{} {
	out := make(map[ProviderID]struct{})
	for _, id := range config.CloudProviderIDs() {
		if _, ok := slots[id]; ok {
			out[id] = struct{}{}
		}
	}
	return out
}

// hasCloudSlot reports whether at least one cloud provider has a configured slot.
func hasCloudSlot(slots map[ProviderID]*providerSlot) bool {
	for _, id := range config.CloudProviderIDs() {
		if _, ok := slots[id]; ok {
			return true
		}
	}
	return false
}

// providerReachable returns a display label for whether the provider is reachable.
func (c *Client) providerReachable(ctx context.Context, httpClient *http.Client, id ProviderID, enabled bool) string {
	if !enabled {
		return models.KeyNo
	}
	pc, ok := c.cfg.Provider(id)
	if !ok {
		return models.KeyNo
	}
	if id == ProviderOllama {
		_, note, err := ollama.FetchTags(ctx, httpClient, pc.BaseURL, pc.Model)
		if err == nil && note == "" {
			return models.KeyYes
		}
		return models.KeyNo
	}
	if _, err := creds.GetProviderAPIKey(id); err != nil {
		return models.KeyNo
	}
	return models.KeyYes
}
