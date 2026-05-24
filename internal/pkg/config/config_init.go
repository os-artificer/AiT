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
	"reflect"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	"gopkg.in/yaml.v3"
)

// EnsureFileResult describes the outcome of EnsureFile.
type EnsureFileResult struct {
	Path      string
	Created   bool
	Updated   bool
	AddedKeys []string
}

// EnsureFile creates or updates config.yaml with default values for missing keys only.
// Existing non-empty values are preserved. When dryRun is true, reports would-be changes without writing.
func EnsureFile(dryRun bool) (EnsureFileResult, error) {
	path, err := platform.ConfigPath()
	if err != nil {
		return EnsureFileResult{}, err
	}

	res := EnsureFileResult{Path: path}
	_, statErr := os.Stat(path)
	res.Created = errors.Is(statErr, os.ErrNotExist)
	if statErr != nil && !res.Created {
		return res, statErr
	}

	tieredInFile := hadTieredInFile(path)

	cfg, err := Load()
	if err != nil {
		return res, err
	}

	before := cloneConfig(cfg)
	added := mergeMissingDefaults(cfg, tieredInFile || cfg.AI.Tiered.Enabled)
	normalizeAI(cfg)

	if len(added) > 0 {
		res.AddedKeys = added
		res.Updated = !res.Created
	} else if !res.Created && !reflect.DeepEqual(before, cfg) {
		res.Updated = true
	}

	if res.Created || res.Updated {
		if dryRun {
			return res, nil
		}
		if err := Save(cfg); err != nil {
			return res, err
		}
	}
	return res, nil
}

func hadTieredInFile(path string) bool {
	b, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var raw map[string]any
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return false
	}
	ai, _ := raw["ai"].(map[string]any)
	if ai == nil {
		return false
	}
	_, ok := ai["tiered"]
	return ok
}

func mergeMissingDefaults(cfg *Config, fillTiered bool) []string {
	def := defaults()
	var added []string

	if cfg.Session.AuthTTL == "" {
		cfg.Session.AuthTTL = def.Session.AuthTTL
		added = append(added, "session.auth_ttl")
	}
	if cfg.AI.Default == "" {
		cfg.AI.Default = def.AI.Default
		added = append(added, "ai.default")
	}

	ensureProviders(cfg)
	for _, id := range CloudProviderIDs() {
		added = append(added, mergeProvider(cfg, def, id)...)
	}

	if fillTiered {
		added = append(added, mergeTieredDefaults(cfg)...)
	}

	return added
}

func mergeProvider(cfg, def *Config, id string) []string {
	var added []string
	p, ok := cfg.AI.Providers[id]
	if !ok {
		cfg.AI.Providers[id] = def.AI.Providers[id]
		return []string{"ai.providers." + id}
	}

	defP := def.AI.Providers[id]
	if p.APIStyle == "" {
		p.APIStyle = defP.APIStyle
		added = append(added, "ai.providers."+id+".api_style")
	}
	if p.BaseURL == "" {
		p.BaseURL = defP.BaseURL
		added = append(added, "ai.providers."+id+".base_url")
	}
	if p.Model == "" {
		p.Model = defP.Model
		added = append(added, "ai.providers."+id+".model")
	}
	cfg.AI.Providers[id] = p
	return added
}

func mergeTieredDefaults(cfg *Config) []string {
	var added []string
	t := &cfg.AI.Tiered
	def := defaults().AI.Tiered

	if t.LocalProvider == "" {
		t.LocalProvider = def.LocalProvider
		added = append(added, "ai.tiered.local_provider")
	}
	if t.MaxPromptChars <= 0 {
		t.MaxPromptChars = def.MaxPromptChars
		added = append(added, "ai.tiered.max_prompt_chars")
	}
	if t.MaxHistoryTurns < 0 {
		t.MaxHistoryTurns = def.MaxHistoryTurns
		added = append(added, "ai.tiered.max_history_turns")
	}
	if len(t.ComplexKeywords) == 0 {
		t.ComplexKeywords = append([]string(nil), def.ComplexKeywords...)
		added = append(added, "ai.tiered.complex_keywords")
	}
	return added
}

func cloneConfig(cfg *Config) *Config {
	if cfg == nil {
		return nil
	}
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return cfg
	}
	out := &Config{}
	if err := yaml.Unmarshal(b, out); err != nil {
		return cfg
	}
	return out
}
