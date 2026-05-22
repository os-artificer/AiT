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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// SetKey sets a dotted configuration key on c.
func (c *Config) SetKey(key, value string) error {
	if key == "session.auth_ttl" {
		if _, err := time.ParseDuration(value); err != nil {
			return err
		}
		c.Session.AuthTTL = value
		return nil
	}

	if key == "ai.default" {
		if !IsCloudProvider(value) {
			return validateCloudDefault(value)
		}
		c.AI.Default = value
		return nil
	}

	if key == "ai.fallback" {
		c.AI.Fallback = parseFallbackList(value)
		return nil
	}

	if field, ok := parseTieredKey(key); ok {
		return c.setTieredField(field, value)
	}

	if id, field, ok := parseProviderKey(key); ok {
		return c.setProviderField(id, field, value)
	}

	return errors.New("unknown config key: " + key)
}

// GetKey returns the string value for a dotted key.
func (c *Config) GetKey(key string) (string, error) {
	switch key {
	case "session.auth_ttl":
		return c.Session.AuthTTL, nil
	case "ai.default":
		return c.DefaultProviderID(), nil
	case "ai.fallback":
		return strings.Join(c.AI.Fallback, ","), nil
	}

	if field, ok := parseTieredKey(key); ok {
		return c.getTieredField(field)
	}

	if id, field, ok := parseProviderKey(key); ok {
		return c.getProviderField(id, field)
	}

	return "", errors.New("unknown config key: " + key)
}

func parseTieredKey(key string) (field string, ok bool) {
	const prefix = "ai.tiered."
	if !strings.HasPrefix(key, prefix) {
		return "", false
	}
	return strings.TrimPrefix(key, prefix), true
}

func (c *Config) setTieredField(field, value string) error {
	normalizeTiered(c)
	switch field {
	case "enabled":
		v, err := parseBool(value)
		if err != nil {
			return err
		}
		c.AI.Tiered.Enabled = v
	case "local_provider":
		if value != ProviderOllama {
			return fmt.Errorf("local_provider must be %q", ProviderOllama)
		}
		c.AI.Tiered.LocalProvider = value
	case "max_prompt_chars":
		n, err := strconv.Atoi(value)
		if err != nil || n <= 0 {
			return errors.New("max_prompt_chars must be a positive integer")
		}
		c.AI.Tiered.MaxPromptChars = n
	case "max_history_turns":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return errors.New("max_history_turns must be a non-negative integer")
		}
		c.AI.Tiered.MaxHistoryTurns = n
	case "complex_keywords":
		c.AI.Tiered.ComplexKeywords = parseKeywordList(value)
	default:
		return errors.New("unknown tiered field: " + field)
	}
	return nil
}

func (c *Config) getTieredField(field string) (string, error) {
	normalizeTiered(c)
	t := c.AI.Tiered
	switch field {
	case "enabled":
		return strconv.FormatBool(t.Enabled), nil
	case "local_provider":
		return t.LocalProvider, nil
	case "max_prompt_chars":
		return strconv.Itoa(t.MaxPromptChars), nil
	case "max_history_turns":
		return strconv.Itoa(t.MaxHistoryTurns), nil
	case "complex_keywords":
		return strings.Join(t.ComplexKeywords, ","), nil
	default:
		return "", errors.New("unknown tiered field: " + field)
	}
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
	if strings.HasPrefix(key, "ai.providers.") {
		rest := strings.TrimPrefix(key, "ai.providers.")
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

func (c *Config) setProviderField(id, field, value string) error {
	ensureProviders(c)
	p := c.AI.Providers[id]

	switch field {
	case "enabled":
		v, err := parseBool(value)
		if err != nil {
			return err
		}
		p.Enabled = &v
	case "base_url":
		if err := validateProviderBaseURL(id, p.APIStyle, value); err != nil {
			return err
		}
		p.BaseURL = value
	case "model":
		p.Model = value
	case "api_style":
		style := strings.ToLower(strings.TrimSpace(value))
		if style != APIStyleOpenAI && style != APIStyleAnthropic {
			return fmt.Errorf("api_style must be openai or anthropic")
		}
		if id == ProviderDeepseek && style == APIStyleAnthropic {
			return errors.New("deepseek provider only supports openai api_style")
		}
		if id == ProviderOllama && style != APIStyleOpenAI {
			return errors.New("ollama provider only supports openai api_style")
		}
		p.APIStyle = style
		if p.BaseURL == "" {
			p.BaseURL = defaultBaseURL(id, style)
		}
	case "models_dir":
		if id != ProviderOllama {
			return errors.New("models_dir is only valid for provider ollama")
		}
		dir, err := expandAbsPath(value)
		if err != nil {
			return err
		}
		p.ModelsDir = dir
	default:
		return errors.New("unknown provider field: " + field)
	}

	c.AI.Providers[id] = p
	return nil
}

func (c *Config) getProviderField(id, field string) (string, error) {
	p, ok := c.Provider(id)
	if !ok {
		p = defaultProviderConfig(id)
	}

	switch field {
	case "enabled":
		return strconv.FormatBool(c.ProviderEnabled(id)), nil
	case "base_url":
		return p.BaseURL, nil
	case "model":
		return p.Model, nil
	case "api_style":
		if p.APIStyle == "" {
			return APIStyleOpenAI, nil
		}
		return p.APIStyle, nil
	case "models_dir":
		return p.ModelsDir, nil
	default:
		return "", errors.New("unknown provider field: " + field)
	}
}

func expandAbsPath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("models_dir must not be empty")
	}
	if strings.HasPrefix(raw, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if raw == "~" {
			raw = home
		} else if strings.HasPrefix(raw, "~/") {
			raw = filepath.Join(home, strings.TrimPrefix(raw, "~/"))
		} else {
			return "", fmt.Errorf("invalid path %q", raw)
		}
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	if !filepath.IsAbs(abs) {
		return "", errors.New("models_dir must be an absolute path")
	}
	return filepath.Clean(abs), nil
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
