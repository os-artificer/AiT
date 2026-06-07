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
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func providerConfigForList(cfg *config.Config, id string) (config.ProviderConfig, bool) {
	pc, inCfg := cfg.Provider(id)
	if inCfg {
		return pc, true
	}
	if id == ProviderOllama {
		return pc, false
	}
	if _, err := creds.GetProviderAPIKey(id); err != nil {
		return pc, false
	}
	return config.DefaultProviderConfig(id), true
}
