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

package cli

import (
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func parseAPIKeyConfigKey(key string) (provider string, ok bool) {
	switch key {
	case "deepseek.api_key":
		return config.ProviderDeepseek, true
	case "minimax.api_key":
		return config.ProviderMinimax, true
	}
	if strings.HasPrefix(key, "ai.") && strings.HasSuffix(key, ".api_key") {
		rest := strings.TrimPrefix(key, "ai.")
		rest = strings.TrimSuffix(rest, ".api_key")
		if config.IsKnownProvider(rest) {
			return rest, true
		}
	}
	return "", false
}

func setProviderAPIKey(provider, value string) error {
	return creds.SetProviderAPIKey(provider, value)
}

func getProviderAPIKey(provider string) (string, error) {
	return creds.GetProviderAPIKey(provider)
}

func apiKeyStoredMessage(provider string) string {
	return fmt.Sprintf("AiT: ai.%s API key stored in keyring", provider)
}
