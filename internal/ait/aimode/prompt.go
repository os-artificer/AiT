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

package aimode

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// Prompt returns the AI mode prompt for scope, host label, and interaction mode.
func Prompt(scope ai.Scope, hostLabel string, mode InteractionMode) string {
	modeLabel := interactionPromptLabel(mode)
	if scope == ai.ScopeRemote {
		if hostLabel != "" {
			return fmt.Sprintf("AiT AI(%s) %s> ", modeLabel, hostLabel)
		}
		return fmt.Sprintf("AiT AI(%s) remote> ", modeLabel)
	}
	return fmt.Sprintf("AiT AI(%s)> ", modeLabel)
}

func interactionPromptLabel(m InteractionMode) string {
	if m == ModeAgent {
		return "Agent"
	}
	return "Ask"
}
