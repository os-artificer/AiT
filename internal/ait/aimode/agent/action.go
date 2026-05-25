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

// Package agent represents the AI agent loop's parsed actions and the local
// capture runner used to feed shell output back into the model.
package agent

import (
	"encoding/json"
	"strings"
)

// Action verbs returned by the model in Agent mode.
const (
	// ActionRun is the canonical verb for "run a shell command".
	ActionRun = "run"
	// ActionRunAlt is the OpenAI function-calling style alias for ActionRun.
	ActionRunAlt = "run_shell"
	// ActionAit is the canonical verb for "invoke an ait CLI subcommand".
	ActionAit = "ait"
	// ActionAitAlt is the OpenAI function-calling style alias for ActionAit.
	ActionAitAlt = "ait_cli"
	// ActionAnswer terminates the agent loop with a free-text reply.
	ActionAnswer = "answer"
)

// Action is one structured step returned by the model in Agent mode.
// Action is the verb; Command/Args/Text carry verb-specific payload.
type Action struct {
	Action  string
	Command string
	Args    string
	Text    string
}

// IsAnswer reports whether a is the terminal answer action.
func (a Action) IsAnswer() bool { return a.Action == ActionAnswer }

// IsRun reports whether a is a run-shell-command action (canonical or alt verb).
func (a Action) IsRun() bool {
	return a.Action == ActionRun || a.Action == ActionRunAlt
}

// IsAit reports whether a is an ait-CLI invocation action (canonical or alt verb).
func (a Action) IsAit() bool {
	return a.Action == ActionAit || a.Action == ActionAitAlt
}

// ParseAction extracts one Action from a model reply (JSON object embedded in
// free text). Returns (zero, false) when the content holds no parseable JSON
// or when the action verb is empty.
func ParseAction(content string) (Action, bool) {
	content = strings.TrimSpace(content)
	if content == "" {
		return Action{}, false
	}
	if obj := extractJSONObject(content); obj != "" {
		return parseJSON(obj)
	}
	return Action{}, false
}

func parseJSON(s string) (Action, bool) {
	var raw struct {
		Action  string `json:"action"`
		Command string `json:"command"`
		Args    string `json:"args"`
		Text    string `json:"text"`
	}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return Action{}, false
	}
	raw.Action = strings.ToLower(strings.TrimSpace(raw.Action))
	if raw.Action == "" {
		return Action{}, false
	}
	return Action{
		Action:  raw.Action,
		Command: strings.TrimSpace(raw.Command),
		Args:    strings.TrimSpace(raw.Args),
		Text:    strings.TrimSpace(raw.Text),
	}, true
}

func extractJSONObject(s string) string {
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return ""
}
