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

var (
	xmlInvokeBlockRe = regexp.MustCompile(`(?s)<invoke\b[^>]*name="([^"]+)"[^>]*>(.*?)</invoke>`)
	xmlParameterRe   = regexp.MustCompile(`(?s)<parameter\b[^>]*name="([^"]+)"[^>]*>([^<]*)</parameter>`)
	hybridRunRe      = regexp.MustCompile(`(?s)"action"\s*:\s*"(?:run_shell|run)"\s*,\s*"command"\s*>\s*([^<]+)\s*</parameter>`)
	hybridAitRe      = regexp.MustCompile(`(?s)"action"\s*:\s*"(?:ait_cli|ait)"\s*,\s*"args"\s*>\s*([^<]+)\s*</parameter>`)
	hybridAnswerRe   = regexp.MustCompile(`(?s)"action"\s*:\s*"answer"\s*,\s*"text"\s*>\s*([^<]+)\s*</parameter>`)
)

// XMLToolCallToJSON extracts the first MiniMax XML-style tool call embedded in
// content and converts it to the same Agent-mode JSON envelope as ToolCallToJSON.
// Returns ("", false) when no XML tool signals are present or parsing fails.
func XMLToolCallToJSON(content string) (string, bool) {
	content = strings.TrimSpace(content)
	if content == "" || !hasXMLToolSignals(content) {
		return "", false
	}

	if m := xmlInvokeBlockRe.FindStringSubmatch(content); len(m) >= 3 {
		return invokeToJSON(m[1], m[2])
	}

	if jsonStr, ok := parseHybridFragment(content); ok {
		return jsonStr, true
	}

	if cmd := extractParameterValue(content, "command"); cmd != "" {
		if strings.Contains(content, `name="run"`) || strings.Contains(content, `name="run_shell"`) {
			return marshalAction(ToolRunShell, "command", cmd)
		}
	}
	if args := extractParameterValue(content, "args"); args != "" {
		if strings.Contains(content, `name="ait"`) || strings.Contains(content, `name="ait_cli"`) {
			return marshalAction(ToolAitCLI, "args", args)
		}
	}
	if text := extractParameterValue(content, "text"); text != "" {
		if strings.Contains(content, `name="answer"`) {
			return marshalAction(ToolAnswer, "text", text)
		}
	}

	return "", false
}

func hasXMLToolSignals(content string) bool {
	return strings.Contains(content, "<invoke") ||
		strings.Contains(content, "<parameter") ||
		strings.Contains(content, "</minimax:tool_call>") ||
		strings.Contains(content, "</invoke>") ||
		strings.Contains(content, "</parameter>") ||
		hybridRunRe.MatchString(content) ||
		hybridAitRe.MatchString(content) ||
		hybridAnswerRe.MatchString(content)
}

func invokeToJSON(invokeName, inner string) (string, bool) {
	invokeName = strings.ToLower(strings.TrimSpace(invokeName))
	params := map[string]string{}
	for _, pm := range xmlParameterRe.FindAllStringSubmatch(inner, -1) {
		params[strings.ToLower(pm[1])] = strings.TrimSpace(pm[2])
	}

	switch invokeName {
	case "run", "run_shell":
		cmd := params["command"]
		if cmd == "" {
			return "", false
		}
		return marshalAction(ToolRunShell, "command", cmd)
	case "ait", "ait_cli":
		args := params["args"]
		if args == "" {
			return "", false
		}
		return marshalAction(ToolAitCLI, "args", args)
	case "answer":
		text := params["text"]
		if text == "" {
			return "", false
		}
		return marshalAction(ToolAnswer, "text", text)
	default:
		return "", false
	}
}

func parseHybridFragment(content string) (string, bool) {
	if m := hybridRunRe.FindStringSubmatch(content); len(m) >= 2 {
		if cmd := strings.TrimSpace(m[1]); cmd != "" {
			return marshalAction(ToolRunShell, "command", cmd)
		}
	}
	if m := hybridAitRe.FindStringSubmatch(content); len(m) >= 2 {
		if args := strings.TrimSpace(m[1]); args != "" {
			return marshalAction(ToolAitCLI, "args", args)
		}
	}
	if m := hybridAnswerRe.FindStringSubmatch(content); len(m) >= 2 {
		if text := strings.TrimSpace(m[1]); text != "" {
			return marshalAction(ToolAnswer, "text", text)
		}
	}
	return "", false
}

func extractParameterValue(content, paramName string) string {
	for _, pm := range xmlParameterRe.FindAllStringSubmatch(content, -1) {
		if strings.EqualFold(pm[1], paramName) {
			return strings.TrimSpace(pm[2])
		}
	}
	hybridRe := regexp.MustCompile(`(?s)"` + regexp.QuoteMeta(paramName) + `"\s*>\s*([^<]+)\s*</parameter>`)
	if m := hybridRe.FindStringSubmatch(content); len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func marshalAction(action, field, value string) (string, bool) {
	b, err := json.Marshal(map[string]string{"action": action, field: value})
	if err != nil {
		return "", false
	}
	return string(b), true
}
