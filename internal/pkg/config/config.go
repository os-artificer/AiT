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
	"time"
)

const (
	DefaultAuthTTL       = 168 * time.Hour // 7 days
	MinAuthTTL           = time.Hour
	MaxAuthTTL           = 30 * 24 * time.Hour
	DefaultDeepseekModel = "deepseek-chat"
	DefaultDeepseekBase  = "https://api.deepseek.com"
)

// Config holds user configuration.
type Config struct {
	Session struct {
		AuthTTL string `yaml:"auth_ttl"`
	} `yaml:"session"`
	Deepseek struct {
		BaseURL string `yaml:"base_url"`
		Model   string `yaml:"model"`
	} `yaml:"deepseek"`
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

// SetKey sets a dotted configuration key on c. Supported keys: session.auth_ttl,
// deepseek.base_url, deepseek.model. value is validated before assignment.
// Errors: time.ParseDuration for auth_ttl; validateDeepseekBaseURL errors; unknown key.
func (c *Config) SetKey(key, value string) error {
	switch key {
	case "session.auth_ttl":
		if _, err := time.ParseDuration(value); err != nil {
			return err
		}
		c.Session.AuthTTL = value
	case "deepseek.base_url":
		if err := validateDeepseekBaseURL(value); err != nil {
			return err
		}
		c.Deepseek.BaseURL = value
	case "deepseek.model":
		c.Deepseek.Model = value
	default:
		return errors.New("unknown config key: " + key)
	}
	return nil
}

// GetKey returns the string value for a dotted key (session.auth_ttl, deepseek.base_url,
// deepseek.model). Returns ("", errors.New("unknown config key")) for unsupported keys.
func (c *Config) GetKey(key string) (string, error) {
	switch key {
	case "session.auth_ttl":
		return c.Session.AuthTTL, nil
	case "deepseek.base_url":
		return c.Deepseek.BaseURL, nil
	case "deepseek.model":
		return c.Deepseek.Model, nil
	default:
		return "", errors.New("unknown config key: " + key)
	}
}
