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

package ollama

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/platform"
)

const envModelsDir = "OLLAMA_MODELS"

// SuggestedModelsDirOption is one interactive choice for setup.
type SuggestedModelsDirOption struct {
	Label string
	Path  string
}

// SuggestedModelsDirs returns preset paths for setup prompts.
func SuggestedModelsDirs(cfg *config.Config) ([]SuggestedModelsDirOption, error) {
	aitDir, err := AitDataModelsDir()
	if err != nil {
		return nil, err
	}
	upstream, err := DefaultUpstreamModelsDir()
	if err != nil {
		return nil, err
	}
	opts := []SuggestedModelsDirOption{
		{Label: "AiT data directory (recommended)", Path: aitDir},
		{Label: "Ollama default (~/.ollama/models)", Path: upstream},
	}
	if cfg != nil {
		if cur := ModelsDirFromConfig(cfg); cur != "" {
			opts = append(opts, SuggestedModelsDirOption{Label: "Keep current config", Path: cur})
		}
	}
	return opts, nil
}

// DefaultUpstreamModelsDir returns Ollama's default models directory (~/.ollama/models).
func DefaultUpstreamModelsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ollama", "models"), nil
}

// AitDataModelsDir returns the recommended AiT data subdirectory for Ollama models.
func AitDataModelsDir() (string, error) {
	data, err := platform.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(data, "ollama", "models"), nil
}

// ModelsDirFromConfig returns the configured models_dir for Ollama, or empty when unset.
func ModelsDirFromConfig(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	pc, ok := cfg.Provider(config.ProviderOllama)
	if !ok {
		return ""
	}
	return pc.ModelsDir
}

// ResolveModelsDir picks the directory for OLLAMA_MODELS: flag override, then config, then AiT data dir.
func ResolveModelsDir(cfg *config.Config, flagOverride string) (string, error) {
	if flagOverride != "" {
		return expandModelsPath(flagOverride)
	}
	if dir := ModelsDirFromConfig(cfg); dir != "" {
		return dir, nil
	}
	return AitDataModelsDir()
}

func expandModelsPath(raw string) (string, error) {
	abs, err := platform.ExpandUserAbsPath(raw)
	if err != nil {
		if errors.Is(err, platform.ErrEmptyPath) {
			return "", fmt.Errorf("empty models directory")
		}
		return "", err
	}
	return abs, nil
}

// SetProviderModelsDir writes models_dir on ai.providers.ollama.
func SetProviderModelsDir(cfg *config.Config, dir string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	abs, err := expandModelsPath(dir)
	if err != nil {
		return err
	}
	if cfg.AI.Providers == nil {
		cfg.AI.Providers = make(map[string]config.ProviderConfig)
	}
	p, ok := cfg.AI.Providers[config.ProviderOllama]
	if !ok {
		p = config.DefaultProviderConfig(config.ProviderOllama)
	}
	p.ModelsDir = abs
	cfg.AI.Providers[config.ProviderOllama] = p
	return nil
}

// ApplyModelsDirEnv sets OLLAMA_MODELS on cmd when dir is non-empty.
func ApplyModelsDirEnv(cmd *exec.Cmd, dir string) {
	if cmd == nil || dir == "" {
		return
	}
	cmd.Env = append(os.Environ(), envModelsDir+"="+dir)
}
