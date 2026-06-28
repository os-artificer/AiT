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
	"regexp"
	"strings"
)

var answerPrefixRe = regexp.MustCompile(
	`(?s)^\s*\{\s*"` + JSONFieldAction + `"\s*:\s*"` + ToolAnswer + `"\s*,\s*"` + JSONFieldText + `"\s*:\s*"`,
)

func finishParse(env AgentEnvelope, source ParseSource) ParseResult {
	if err := ValidateEnvelope(env); err != nil {
		return ParseResult{Envelope: env, Source: source, Errs: []error{err}}
	}
	return ParseResult{Envelope: env, Source: source}
}

func parseEnvelopeStrict(content string) (AgentEnvelope, bool) {
	if obj := extractJSONObject(content); obj != "" {
		return parseEnvelopeJSON(obj)
	}
	return AgentEnvelope{}, false
}

func parseEnvelopeJSON(s string) (AgentEnvelope, bool) {
	s = NormalizeES6UnicodeJSON(s)

	var raw struct {
		Action  string `json:"action"`
		Command string `json:"command"`
		Args    string `json:"args"`
		Text    string `json:"text"`
	}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return AgentEnvelope{}, false
	}

	raw.Action = strings.ToLower(strings.TrimSpace(raw.Action))
	if raw.Action == "" {
		return AgentEnvelope{}, false
	}
	return AgentEnvelope{
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

func extractAnswerPrefixHeuristic(content string) (AgentEnvelope, bool) {
	loc := answerPrefixRe.FindStringIndex(content)
	if loc == nil {
		return AgentEnvelope{}, false
	}

	text := content[loc[1]:]
	text = strings.TrimSuffix(text, "\n}")
	text = strings.TrimSuffix(text, "}")
	text = strings.TrimSuffix(text, `"`)
	if text == "" {
		return AgentEnvelope{}, false
	}
	return AgentEnvelope{Action: ToolAnswer, Text: text}, true
}

func looksLikeFailedEnvelope(content string) bool {
	if strings.HasPrefix(strings.TrimSpace(content), "{") {
		return true
	}
	if strings.Contains(content, `"`+JSONFieldAction+`"`) {
		return true
	}
	return hasXMLToolSignals(content)
}
