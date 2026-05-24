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
	"encoding/json"
	"strings"
)

const (
	agentActionRun    = "run"
	agentActionRunAlt = "run_shell"
	agentActionAit    = "ait"
	agentActionAitAlt = "ait_cli"
	agentActionAnswer = "answer"
)

// AgentAction is one structured step from the model in Agent mode.
type AgentAction struct {
	Action  string
	Command string
	Args    string
	Text    string
}

// parseAgentAction extracts one action from model output (JSON object).
func parseAgentAction(content string) (AgentAction, bool) {
	content = strings.TrimSpace(content)
	if content == "" {
		return AgentAction{}, false
	}
	if json := extractJSONObject(content); json != "" {
		return parseAgentJSON(json)
	}
	return AgentAction{}, false
}

func parseAgentJSON(s string) (AgentAction, bool) {
	var raw struct {
		Action  string `json:"action"`
		Command string `json:"command"`
		Args    string `json:"args"`
		Text    string `json:"text"`
	}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return AgentAction{}, false
	}
	raw.Action = strings.ToLower(strings.TrimSpace(raw.Action))
	if raw.Action == "" {
		return AgentAction{}, false
	}
	return AgentAction{
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

func actionIsAnswer(a AgentAction) bool {
	return a.Action == agentActionAnswer
}

func actionIsRun(a AgentAction) bool {
	return a.Action == agentActionRun || a.Action == agentActionRunAlt
}

func actionIsAit(a AgentAction) bool {
	return a.Action == agentActionAit || a.Action == agentActionAitAlt
}
