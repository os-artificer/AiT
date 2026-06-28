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
	"net/http"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

type providerSlot struct {
	id       ProviderID
	apiStyle string
	baseURL  string
	model    string
	key      string
	http     *http.Client
}

func newProviderSlot(id ProviderID, pc config.ProviderConfig, key string, httpClient *http.Client) *providerSlot {
	style := pc.APIStyle
	if style == "" {
		style = config.APIStyleOpenAI
	}
	slot := &providerSlot{
		id:       id,
		apiStyle: style,
		baseURL:  pc.BaseURL,
		model:    pc.Model,
		key:      key,
		http:     httpClient,
	}
	if slot.http == nil && style == config.APIStyleAnthropic {
		slot.http = http.DefaultClient
	}
	return slot
}

func initTieredSlot(cfg *config.Config, localID string, httpClient *http.Client) *providerSlot {
	pc, ok := cfg.Provider(localID)
	if !ok {
		return nil
	}
	return newProviderSlot(localID, pc, ollamaPlaceholderKey, httpClient)
}
