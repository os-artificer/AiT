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

// Normalize fills defaults and clamps values for agent config.
// An unparseable command_timeout is left as written so Validate can report the
// field; AgentCommandTimeout falls back to the default meanwhile.
func (a *AgentConfig) Normalize() {
	if a.MaxSteps <= 0 {
		a.MaxSteps = DefaultAgentMaxSteps
	} else if a.MaxSteps > AgentMaxStepsLimit {
		a.MaxSteps = AgentMaxStepsLimit
	}

	a.CommandTimeout = clampTimeoutString(a.CommandTimeout)
}

// clampTimeoutString returns raw clamped into the agent command timeout bounds,
// or raw unchanged when it is blank or not a valid duration.
func clampTimeoutString(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return raw
	}

	d, err := time.ParseDuration(trimmed)
	if err != nil {
		return raw
	}

	if d < MinAgentCommandTimeout {
		return MinAgentCommandTimeout.String()
	}
	if d > MaxAgentCommandTimeout {
		return MaxAgentCommandTimeout.String()
	}
	return trimmed
}
