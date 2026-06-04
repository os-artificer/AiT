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
	"strings"
	"time"
)

const (
	DefaultAuthTTL       = 168 * time.Hour // 7 days
	MinAuthTTL           = time.Hour
	MaxAuthTTL           = 30 * 24 * time.Hour
	DefaultDeepseekModel = "deepseek-chat"
	DefaultDeepseekBase  = "https://api.deepseek.com"

	RemotePromptPathShort    = "short"
	RemotePromptPathAbsolute = "absolute"
)

// RemotePromptConfig holds SSH remote shell prompt display options.
type RemotePromptConfig struct {
	Path string `yaml:"path"`
}

// Config holds user configuration.
type Config struct {
	Session struct {
		AuthTTL      string             `yaml:"auth_ttl"`
		RemotePrompt RemotePromptConfig `yaml:"remote_prompt"`
	} `yaml:"session"`
	AI AIConfig `yaml:"ai"`
	// Deepseek is legacy YAML only; merged into ai.providers on Load. Not written on Save.
	Deepseek struct {
		BaseURL string `yaml:"base_url,omitempty"`
		Model   string `yaml:"model,omitempty"`
	} `yaml:"deepseek,omitempty"`
}

// AuthTTL parses session.auth_ttl as a duration. Nil or empty config uses DefaultAuthTTL.
// Invalid duration strings return DefaultAuthTTL and the parse error; values are clamped to
// MinAuthTTL and MaxAuthTTL.
func (c *Config) AuthTTL() (time.Duration, error) {
	if c == nil || c.Session.AuthTTL == "" {
		return DefaultAuthTTL, nil
	}
	d, err := time.ParseDuration(c.Session.AuthTTL)
	if err != nil {
		return DefaultAuthTTL, err
	}

	if d < MinAuthTTL {
		d = MinAuthTTL
	}
	if d > MaxAuthTTL {
		d = MaxAuthTTL
	}
	return d, nil
}

// RemotePromptPathAbsolute reports whether session.remote_prompt.path is absolute.
// Empty or unknown values use short (basename) mode.
func (c *Config) RemotePromptPathAbsolute() bool {
	if c == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(c.Session.RemotePrompt.Path), RemotePromptPathAbsolute)
}
