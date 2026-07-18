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
	"os"
	"sync"

	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

// ErrAIExperimentalDisabled is the sentinel when AI experimental features are off.
var ErrAIExperimentalDisabled = errhint.NewSentinel("AI experimental features are disabled")

var (
	aiExperimentalCLIMU sync.RWMutex
	aiExperimentalCLI   bool
)

const (
	// EnvAIExperimental enables AI experimental features when set to "1".
	EnvAIExperimental = "AIT_AI_EXPERIMENTAL"
	// FlagAIExperimental is the process-entry CLI flag (stripped before route/parse).
	FlagAIExperimental = "--ai-experimental"

	hintAIExperimentalExplain = "AI features in aitssh are experimental and disabled by default."
)

// SetAIExperimentalCLI records whether the process-entry CLI flag was seen.
// Only the top-level ExecuteWriters path should set this to true.
func SetAIExperimentalCLI(enabled bool) {
	aiExperimentalCLIMU.Lock()
	aiExperimentalCLI = enabled
	aiExperimentalCLIMU.Unlock()
}

// AIExperimentalCLI reports the process-entry CLI flag state.
func AIExperimentalCLI() bool {
	aiExperimentalCLIMU.RLock()
	defer aiExperimentalCLIMU.RUnlock()
	return aiExperimentalCLI
}

// AIExperimentalEnabled reports whether AI experimental features are on via CLI, env, or config.
// A nil cfg skips the YAML bit (CLI and env still apply).
func AIExperimentalEnabled(cfg *Config) bool {
	if AIExperimentalCLI() {
		return true
	}
	if os.Getenv(EnvAIExperimental) == "1" {
		return true
	}
	return cfg != nil && cfg.AI.Experimental.Enabled
}

// AIExperimentalDisabledErr builds the multi-line Try error for disabled experimental AI.
// It does not inspect CLI, env, or config — callers that need a gate check use RequireAIExperimental.
// Always returns an error satisfying errors.Is(err, ErrAIExperimentalDisabled).
// Callers: RequireAIExperimental (disabled branch) and the SSH disabled OnRemoteAI stub.
func AIExperimentalDisabledErr() error {
	return errhint.WithTry(ErrAIExperimentalDisabled,
		hintAIExperimentalExplain,
		clihints.TryAIExperimentalCLI,
		clihints.TryAIExperimentalConfig,
		clihints.TryAIExperimentalEnv,
	)
}

// RequireAIExperimental returns ErrAIExperimentalDisabled with Try lines when features are off.
// When cfg is nil, Load is attempted for the YAML bit; Load errors are ignored (treat as off).
func RequireAIExperimental(cfg *Config) error {
	if cfg == nil {
		loaded, err := Load()
		if err == nil {
			cfg = loaded
		}
	}
	if AIExperimentalEnabled(cfg) {
		return nil
	}
	return AIExperimentalDisabledErr()
}

// StripAIExperimentalFlag removes all --ai-experimental tokens from args.
// saw is true when at least one flag was present.
func StripAIExperimentalFlag(args []string) (rest []string, saw bool) {
	if len(args) == 0 {
		return args, false
	}
	out := make([]string, 0, len(args))
	for _, a := range args {
		if a == FlagAIExperimental {
			saw = true
			continue
		}
		out = append(out, a)
	}
	return out, saw
}
