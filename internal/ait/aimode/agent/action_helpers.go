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

package agent

import (
	"encoding/json"
	"strings"
)

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
