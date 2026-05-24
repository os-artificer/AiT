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
	"strings"
)

// RiskLevel classifies a command before execution.
type RiskLevel int

const (
	// RiskLow is a normal shell command.
	RiskLow RiskLevel = iota
	// RiskHigh is destructive or privileged; always confirm per command.
	RiskHigh
	// RiskInteractive is a TTY UI; never auto-run in Agent mode.
	RiskInteractive
)

// ClassifyRisk returns the risk of running cmd as a shell command.
func ClassifyRisk(cmd string) RiskLevel {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return RiskLow
	}
	if ClassifyInteractive(cmd) {
		return RiskInteractive
	}
	lower := strings.ToLower(cmd)
	for _, frag := range highRiskFragments {
		if strings.Contains(lower, frag) {
			return RiskHigh
		}
	}
	return RiskLow
}

// ClassifyInteractive reports whether cmd starts an interactive TUI program.
func ClassifyInteractive(cmd string) bool {
	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) == 0 {
		return false
	}
	name := strings.ToLower(fields[0])
	if name == "sudo" && len(fields) > 1 {
		name = strings.ToLower(fields[1])
	}
	_, ok := interactiveCommands[name]
	return ok
}

// ClassifyAitRisk returns risk for an ait CLI invocation line (without the "ait" prefix).
func ClassifyAitRisk(argsLine string) RiskLevel {
	lower := strings.ToLower(strings.TrimSpace(argsLine))
	for _, frag := range highRiskAitFragments {
		if strings.Contains(lower, frag) {
			return RiskHigh
		}
	}
	return RiskLow
}
