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

package creds

import (
	"fmt"

	"github.com/99designs/keyring"
	"github.com/os-artificer/AiT/internal/pkg/config"
)

func providerAPIKeyKey(provider string) string {
	return fmt.Sprintf("ai/%s/api_key", provider)
}

// SetProviderAPIKey stores the API key for provider in the OS keyring (key ai/<provider>/api_key).
// provider must be a known ID such as deepseek or minimax.
// Returns keyring open or Set errors.
func SetProviderAPIKey(provider, key string) error {
	kr, err := openRing()
	if err != nil {
		return err
	}

	return kr.Set(keyring.Item{
		Key:  providerAPIKeyKey(provider),
		Data: []byte(key),
	})
}

// GetProviderAPIKey returns the API key for provider from the keyring.
// For deepseek, falls back to legacy deepseek/api_key and migrates to ai/deepseek/api_key.
// Returns keyring open/Get errors, including keyring.ErrKeyNotFound when unset.
func GetProviderAPIKey(provider string) (string, error) {
	kr, err := openRing()
	if err != nil {
		return "", err
	}

	item, err := kr.Get(providerAPIKeyKey(provider))
	if err == nil {
		return string(item.Data), nil
	}

	if provider != config.ProviderDeepseek {
		return "", err
	}

	legacy, legErr := kr.Get("deepseek/api_key")
	if legErr != nil {
		return "", err
	}

	key := string(legacy.Data)
	_ = SetProviderAPIKey(config.ProviderDeepseek, key)
	return key, nil
}

// SetMinimaxAPIKey stores the MiniMax API key (alias for SetProviderAPIKey minimax).
func SetMinimaxAPIKey(key string) error {
	return SetProviderAPIKey(config.ProviderMinimax, key)
}

// GetMinimaxAPIKey returns the MiniMax API key.
func GetMinimaxAPIKey() (string, error) {
	return GetProviderAPIKey(config.ProviderMinimax)
}
