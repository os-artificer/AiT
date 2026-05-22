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
	"os"
	"path/filepath"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	"gopkg.in/yaml.v3"
)

// Load reads config.yaml from the platform config directory, or returns defaults when missing.
// Applies DefaultDeepseekBase and DefaultDeepseekModel when those fields are empty after parse.
// Errors: platform.ConfigDir failure; os.ReadFile (other than not exist); yaml.Unmarshal.
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

	normalizeAI(cfg)
	return cfg, nil
}

func defaults() *Config {
	cfg := &Config{}
	cfg.Session.AuthTTL = DefaultAuthTTL.String()
	cfg.AI.Default = ProviderDeepseek
	cfg.AI.Providers = map[string]ProviderConfig{
		ProviderDeepseek: defaultProviderConfig(ProviderDeepseek),
		ProviderMinimax:  defaultProviderConfig(ProviderMinimax),
	}
	return cfg
}

// Save writes cfg to config.yaml via a temporary file and atomic rename (mode 0600).
// cfg must be non-nil. Errors: platform.ConfigDir; yaml.Marshal; WriteFile; Rename.
func Save(cfg *Config) error {
	dir, err := platform.ConfigDir()
	if err != nil {
		return err
	}

	normalizeAI(cfg)
	clearLegacyDeepseekForSave(cfg)

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
