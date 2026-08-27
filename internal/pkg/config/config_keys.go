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

package config

import (
	"fmt"
	"strings"
	"time"
)

// SetKey sets a dotted configuration key on c.
func (c *Config) SetKey(key, value string) error {
	if key == keySessionAuthTTL {
		if _, err := time.ParseDuration(value); err != nil {
			return err
		}
		c.Session.AuthTTL = value
		return nil
	}

	if key == keySessionRemotePromptPath {
		v := strings.ToLower(strings.TrimSpace(value))
		if v != RemotePromptPathShort && v != RemotePromptPathAbsolute {
			return fmt.Errorf("session.remote_prompt.path must be %q or %q", RemotePromptPathShort, RemotePromptPathAbsolute)
		}
		c.Session.RemotePrompt.Path = v
		return nil
	}

	if key == keyAIDefault {
		if !IsCloudProvider(value) {
			return validateCloudDefault(value)
		}
		c.AI.Default = value
		return nil
	}

	if key == keyAIFallback {
		c.AI.Fallback = parseFallbackList(value)
		return nil
	}

	if key == keyAIExperimentalEnabled {
		v, err := parseBool(value)
		if err != nil {
			return fmt.Errorf("%s: %w", keyAIExperimentalEnabled, err)
		}
		c.AI.Experimental.Enabled = v
		return nil
	}

	if key == keyAIAgentCommandTimeout {
		if _, err := time.ParseDuration(value); err != nil {
			return err
		}
		c.AI.Agent.CommandTimeout = value
		return nil
	}

	if field, ok := parseTieredKey(key); ok {
		return c.setTieredField(field, value)
	}

	if id, field, ok := parseProviderKey(key); ok {
		return c.setProviderField(id, field, value)
	}

	return unknownConfigKeyErr(key)
}

// GetKey returns the string value for a dotted key.
func (c *Config) GetKey(key string) (string, error) {
	switch key {
	case keySessionAuthTTL:
		return c.Session.AuthTTL, nil
	case keySessionRemotePromptPath:
		if c.Session.RemotePrompt.Path == "" {
			return RemotePromptPathShort, nil
		}
		return c.Session.RemotePrompt.Path, nil
	case keyAIDefault:
		return c.DefaultProviderID(), nil
	case keyAIFallback:
		return strings.Join(c.AI.Fallback, ","), nil
	case keyAIExperimentalEnabled:
		if c.AI.Experimental.Enabled {
			return "true", nil
		}
		return "false", nil
	case keyAIAgentCommandTimeout:
		if strings.TrimSpace(c.AI.Agent.CommandTimeout) == "" {
			return DefaultAgentCommandTimeoutYAML, nil
		}
		return c.AI.Agent.CommandTimeout, nil
	}

	if field, ok := parseTieredKey(key); ok {
		return c.getTieredField(field)
	}

	if id, field, ok := parseProviderKey(key); ok {
		return c.getProviderField(id, field)
	}

	return "", unknownConfigKeyErr(key)
}

func parseTieredKey(key string) (field string, ok bool) {
	const prefix = keyAITieredPrefix
	if !strings.HasPrefix(key, prefix) {
		return "", false
	}
	return strings.TrimPrefix(key, prefix), true
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean %q", s)
	}
}

func parseKeywordList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseProviderKey(key string) (id, field string, ok bool) {
	if strings.HasPrefix(key, keyAIProvidersPrefix) {
		rest := strings.TrimPrefix(key, keyAIProvidersPrefix)
		parts := strings.SplitN(rest, ".", 2)
		if len(parts) == 2 && IsKnownProvider(parts[0]) {
			return parts[0], parts[1], true
		}
		return "", "", false
	}

	for _, id := range AllProviderIDs() {
		prefix := id + "."
		if strings.HasPrefix(key, prefix) {
			return id, strings.TrimPrefix(key, prefix), true
		}
	}
	return "", "", false
}

func parseFallbackList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part != "" && IsCloudProvider(part) {
			out = append(out, part)
		}
	}
	return out
}

func ensureProviders(c *Config) {
	if c.AI.Providers == nil {
		c.AI.Providers = make(map[string]ProviderConfig)
	}
}
