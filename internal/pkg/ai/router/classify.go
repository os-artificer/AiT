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

package router

import (
	"strings"
	"unicode/utf8"

	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
	"github.com/os-artificer/AiT/internal/pkg/config"
)

// ClassifyPrompt reports whether tiered routing should try the local Ollama
// provider before any cloud provider. It inspects the configured limits
// (history length, prompt length, complex keywords, code fences) and the
// session override.
func ClassifyPrompt(cfg *config.Config, user string, historyLen int, sessionOverride provider.ID, skipTiered bool) bool {
	if cfg == nil || !cfg.TieredEnabled() || skipTiered {
		return false
	}
	if !cfg.ProviderEnabled(cfg.LocalProviderID()) {
		return false
	}
	if sessionOverride != "" {
		return false
	}
	if historyLen > cfg.AI.Tiered.MaxHistoryTurns {
		return false
	}
	user = strings.TrimSpace(user)
	if utf8.RuneCountInString(user) > cfg.AI.Tiered.MaxPromptChars {
		return false
	}
	if strings.Contains(user, "```") {
		return false
	}
	lower := strings.ToLower(user)
	for _, kw := range cfg.AI.Tiered.ComplexKeywords {
		if kw != "" && strings.Contains(lower, strings.ToLower(kw)) {
			return false
		}
	}
	return true
}
