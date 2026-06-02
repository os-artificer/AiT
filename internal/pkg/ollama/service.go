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
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// ErrAlreadyRunning is returned when StartServe finds the API already reachable.
var ErrAlreadyRunning = fmt.Errorf("ollama API already reachable")

// StartServeDetached runs ollama serve in the background with OLLAMA_MODELS when modelsDir is set.
func StartServeDetached(modelsDir string) error {
	cmd := exec.Command(binaryName, "serve")
	cmd.Stdout = nil
	cmd.Stderr = nil
	ApplyModelsDirEnv(cmd, modelsDir)
	return cmd.Start()
}

// StartServe starts ollama serve when the API is not yet reachable.
// When wait is true, polls until /api/tags responds or probeTotalTimeout elapses.
// Returns ErrAlreadyRunning when the API is already up (caller may treat as success).
// modelsDir is passed to OLLAMA_MODELS when starting serve (empty uses process default).
func StartServe(ctx context.Context, client *http.Client, baseURL string, wait bool, modelsDir string) error {
	if client == nil {
		client = &http.Client{Timeout: httpClientTimeout}
	}
	avail := CheckAvailable(ctx, client, baseURL)
	if avail.APIOK {
		return ErrAlreadyRunning
	}
	if !avail.BinaryOK {
		return fmt.Errorf("ollama binary not in PATH (run: ait ai ollama setup)")
	}
	if err := StartServeDetached(modelsDir); err != nil {
		return fmt.Errorf("start ollama serve: %w", err)
	}
	if !wait {
		return nil
	}
	return waitAPIOnly(ctx, client, baseURL)
}

func waitAPIOnly(ctx context.Context, client *http.Client, baseURL string) error {
	deadline := time.Now().Add(probeTotalTimeout)
	for time.Now().Before(deadline) {
		avail := CheckAvailable(ctx, client, baseURL)
		if avail.APIOK {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(probeInterval):
		}
	}
	return fmt.Errorf("ollama API not reachable at %s after start", TagsBaseURL(baseURL))
}

// StopServe attempts to stop local ollama serve processes.
func StopServe(ctx context.Context) error {
	_ = ctx
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("taskkill", "/IM", "ollama.exe", "/F")
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
				return nil
			}
			return fmt.Errorf("stop ollama: %w (quit the Ollama desktop app if it is still running)", err)
		}
		return nil
	default:
		cmd := exec.Command("pkill", "-f", "ollama serve")
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				return nil
			}
			return fmt.Errorf("stop ollama: %w", err)
		}
		return nil
	}
}

// DisableProvider sets ai.providers.ollama.enabled to false. When stop is true, runs StopServe first.
func DisableProvider(ctx context.Context, cfg *config.Config, stop bool) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if stop {
		if err := StopServe(ctx); err != nil {
			return err
		}
	}
	return cfg.SetProviderEnabled(config.ProviderOllama, false)
}

// BaseURLFromConfig returns the configured Ollama OpenAI base URL or the default.
func BaseURLFromConfig(cfg *config.Config) string {
	if cfg == nil {
		return config.DefaultOllamaBase
	}
	pc, ok := cfg.Provider(config.ProviderOllama)
	if !ok || pc.BaseURL == "" {
		return config.DefaultOllamaBase
	}
	return pc.BaseURL
}

// ModelFromConfig returns the configured Ollama model or the default.
func ModelFromConfig(cfg *config.Config) string {
	if cfg == nil {
		return config.DefaultOllamaModel
	}
	pc, ok := cfg.Provider(config.ProviderOllama)
	if !ok || pc.Model == "" {
		return config.DefaultOllamaModel
	}
	return pc.Model
}

// BinaryInPath reports whether the ollama binary is on PATH.
func BinaryInPath() bool {
	_, err := platform.LookPath(binaryName)
	return err == nil
}
