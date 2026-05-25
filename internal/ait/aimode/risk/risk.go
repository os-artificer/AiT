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

// Package risk classifies AI-mode commands by execution risk so callers can
// gate confirmation, denylist TUI programs, and apply pattern/session trust.
package risk

import "strings"

// Level classifies a command before execution.
type Level int

const (
	// Low is a normal shell command.
	Low Level = iota
	// High is destructive or privileged; always confirm per command.
	High
	// Interactive is a TTY UI; never auto-run in Agent mode.
	Interactive
)

// Classify returns the risk Level of running cmd as a shell command.
// Returns Low for empty input, Interactive when cmd launches a TTY UI,
// High when cmd matches any denylist fragment, otherwise Low.
func Classify(cmd string) Level {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return Low
	}
	if IsInteractive(cmd) {
		return Interactive
	}
	lower := strings.ToLower(cmd)
	for _, frag := range highRiskFragments {
		if strings.Contains(lower, frag) {
			return High
		}
	}
	return Low
}

// IsInteractive reports whether cmd starts an interactive TUI program.
// Looks at the first token (or the second when prefixed with sudo).
func IsInteractive(cmd string) bool {
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

// ClassifyAit returns the risk Level for an ait CLI invocation line
// (without the "ait" prefix). Returns High when the line touches
// destructive subcommands, otherwise Low.
func ClassifyAit(argsLine string) Level {
	lower := strings.ToLower(strings.TrimSpace(argsLine))
	for _, frag := range highRiskAitFragments {
		if strings.Contains(lower, frag) {
			return High
		}
	}
	return Low
}
