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

// Package shared holds helpers shared by aissh config subcommands.
package shared

import (
	"fmt"
	"strings"

	pkgconfig "github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

const (
	keyPrefixAI     = "ai."
	keySuffixAPIKey = ".api_key"
)

// ParseAPIKeyConfigKey reports whether key is an API key path and returns the provider id.
func ParseAPIKeyConfigKey(key string) (provider string, ok bool) {
	switch key {
	case "deepseek.api_key":
		return pkgconfig.ProviderDeepseek, true
	case "minimax.api_key":
		return pkgconfig.ProviderMinimax, true
	}
	if strings.HasPrefix(key, keyPrefixAI) && strings.HasSuffix(key, keySuffixAPIKey) {
		rest := strings.TrimPrefix(key, keyPrefixAI)
		rest = strings.TrimSuffix(rest, keySuffixAPIKey)
		if pkgconfig.IsKnownProvider(rest) {
			return rest, true
		}
	}
	return "", false
}

// SetProviderAPIKey stores an API key in the OS keyring.
func SetProviderAPIKey(provider, value string) error {
	return creds.SetProviderAPIKey(provider, value)
}

// GetProviderAPIKey reads an API key from the OS keyring.
func GetProviderAPIKey(provider string) (string, error) {
	return creds.GetProviderAPIKey(provider)
}

// APIKeyStoredMessage is printed after storing a key in the keyring.
func APIKeyStoredMessage(provider string) string {
	return fmt.Sprintf("AiT: ai.%s API key stored in keyring", provider)
}

// MaskSecret masks a secret for display.
func MaskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "..."
}
