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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	"gopkg.in/yaml.v3"
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

// Load reads config from disk or returns defaults.
func Load() (*Config, error) {
	cfg := defaults()
	dir, err := platform.ConfigDir()
	if err != nil {
		return cfg, err
	}
	path := filepath.Join(dir, "config.yaml")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	if cfg.Deepseek.BaseURL == "" {
		cfg.Deepseek.BaseURL = DefaultDeepseekBase
	}
	if cfg.Deepseek.Model == "" {
		cfg.Deepseek.Model = DefaultDeepseekModel
	}
	return cfg, nil
}

func validateDeepseekBaseURL(raw string) error {
	if os.Getenv("AIT_ALLOW_CUSTOM_AI_URL") == "1" {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return errors.New("deepseek.base_url must be a valid https URL")
		}
		return nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return errors.New("deepseek.base_url must use https")
	}
	host := strings.ToLower(u.Hostname())
	if host != "api.deepseek.com" {
		return errors.New("deepseek.base_url host must be api.deepseek.com (set AIT_ALLOW_CUSTOM_AI_URL=1 to override)")
	}
	return nil
}

func defaults() *Config {
	cfg := &Config{}
	cfg.Session.AuthTTL = DefaultAuthTTL.String()
	cfg.Deepseek.BaseURL = DefaultDeepseekBase
	cfg.Deepseek.Model = DefaultDeepseekModel
	return cfg
}

// AuthTTL parses session.auth_ttl from config.
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

// Save writes config to disk atomically.
func Save(cfg *Config) error {
	dir, err := platform.ConfigDir()
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	final := filepath.Join(dir, "config.yaml")
	tmp := filepath.Join(dir, ".config.yaml.tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, final)
}

// SetKey sets a dotted key like session.auth_ttl or deepseek.model.
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

// GetKey returns a config value by dotted key.
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
