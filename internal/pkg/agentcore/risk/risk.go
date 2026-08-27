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
	// Low is a normal shell command (may read or write).
	Low Level = iota
	// Read is an observably read-only command; auto-approved in Agent mode.
	Read
	// High is destructive or privileged; always confirm per command.
	High
	// Interactive is a TTY UI; never auto-run in Agent mode.
	Interactive
)

// Classify returns the risk Level of running cmd as a shell command.
// cmd is split into segments at unquoted shell operators and each segment is
// classified on its own, so a read-only first token cannot hide a destructive
// later segment (`ls; rm -rf /`). The strictest segment level wins, and
// fragments that pipe into a shell are matched against the whole line.
// Returns Low for empty input, Interactive when any segment launches a TTY UI,
// High when any segment is destructive, Read when every segment is observably
// read-only and free of writing metacharacters, otherwise Low.
func Classify(cmd string) Level {
	segments := SplitSegments(cmd)
	if len(segments) == 0 {
		return Low
	}

	worst := Read
	for _, seg := range segments {
		worst = moreSevere(worst, classifySegment(seg))
	}

	if matchesPipeHighRisk(cmd) {
		worst = moreSevere(worst, High)
	}
	return worst
}

// IsRead reports whether cmd is an observably read-only command.
// It checks the first token (or the second when prefixed with sudo)
// against the readCommands map. For git, it also checks the subcommand
// against readGitSubcommands.
func IsRead(cmd string) bool {
	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) == 0 {
		return false
	}
	name := strings.ToLower(fields[0])
	if name == "sudo" && len(fields) > 1 {
		name = strings.ToLower(fields[1])
		fields = fields[1:]
	}
	if _, ok := readCommands[name]; !ok {
		return false
	}
	// git is in readCommands but only certain subcommands are read-only.
	if name == "git" && len(fields) > 1 {
		sub := strings.ToLower(fields[1])
		// Handle flags between git and subcommand, e.g. "git -C /tmp status"
		for i := 1; i < len(fields); i++ {
			f := fields[i]
			if strings.HasPrefix(f, "-") {
				continue
			}
			sub = strings.ToLower(f)
			break
		}
		_, ok := readGitSubcommands[sub]
		return ok
	}
	return true
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

// ClassifyAit returns the risk Level for an aissh CLI invocation line
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
