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
	xmlInvokeBlockRe   = regexp.MustCompile(`(?s)<invoke\b[^>]*name="([^"]+)"[^>]*>(.*?)</invoke>`)
	xmlOpenInvokeRe    = regexp.MustCompile(`(?s)<invoke\b[^>]*name="([^"]+)"[^>]*>(.*)`)
	xmlParameterRe     = regexp.MustCompile(`(?s)<parameter\b[^>]*name="([^"]+)"[^>]*>([^<]*)</parameter>`)
	xmlOpenParameterRe = regexp.MustCompile(`(?s)<parameter\b[^>]*name="([^"]+)"[^>]*>(.*?)(?:</parameter>|</invoke>|</minimax:tool_call>|\z)`)
	hybridRunRe        = regexp.MustCompile(`(?s)"` + JSONFieldAction + `"\s*:\s*"(?:` + ToolRunShell + `|` + XMLInvokeRun + `)"\s*,\s*"` + JSONFieldCommand + `"\s*>\s*([^<]+)\s*</parameter>`)
	hybridCLIRe        = regexp.MustCompile(`(?s)"` + JSONFieldAction + `"\s*:\s*"(?:` + ToolAISshCLI + `|` + ToolAitCLI + `|` + XMLInvokeAISsh + `|` + XMLInvokeAit + `)"\s*,\s*"` + JSONFieldArgs + `"\s*>\s*([^<]+)\s*</parameter>`)
	hybridAnswerRe     = regexp.MustCompile(`(?s)"` + JSONFieldAction + `"\s*:\s*"` + ToolAnswer + `"\s*,\s*"` + JSONFieldText + `"\s*>\s*([^<]+)\s*</parameter>`)
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

	if m := xmlOpenInvokeRe.FindStringSubmatch(content); len(m) >= 3 {
		if jsonStr, ok := openInvokeToJSON(m[1], m[2]); ok {
			return jsonStr, true
		}
	}

	if jsonStr, ok := parseHybridFragment(content); ok {
		return jsonStr, true
	}

	if cmd := extractParameterValue(content, JSONFieldCommand); cmd != "" {
		if strings.Contains(content, `name="`+XMLInvokeRun+`"`) || strings.Contains(content, `name="`+ToolRunShell+`"`) {
			return marshalAction(ToolRunShell, JSONFieldCommand, cmd)
		}
	}
	if args := extractParameterValue(content, JSONFieldArgs); args != "" {
		if strings.Contains(content, `name="`+XMLInvokeAISsh+`"`) ||
			strings.Contains(content, `name="`+XMLInvokeAit+`"`) ||
			strings.Contains(content, `name="`+ToolAISshCLI+`"`) ||
			strings.Contains(content, `name="`+ToolAitCLI+`"`) {
			return marshalCLIAction(args)
		}
	}
	if text := extractParameterValue(content, JSONFieldText); text != "" {
		if strings.Contains(content, `name="`+ToolAnswer+`"`) {
			return marshalAction(ToolAnswer, JSONFieldText, text)
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
		hybridCLIRe.MatchString(content) ||
		hybridAnswerRe.MatchString(content)
}

func openInvokeToJSON(invokeName, inner string) (string, bool) {
	params := map[string]string{}
	for _, pm := range xmlOpenParameterRe.FindAllStringSubmatch(inner, -1) {
		params[strings.ToLower(pm[1])] = strings.TrimSpace(pm[2])
	}
	return paramsToJSON(strings.ToLower(strings.TrimSpace(invokeName)), params)
}

func invokeToJSON(invokeName, inner string) (string, bool) {
	invokeName = strings.ToLower(strings.TrimSpace(invokeName))
	params := map[string]string{}
	for _, pm := range xmlParameterRe.FindAllStringSubmatch(inner, -1) {
		params[strings.ToLower(pm[1])] = strings.TrimSpace(pm[2])
	}
	return paramsToJSON(invokeName, params)
}

func paramsToJSON(invokeName string, params map[string]string) (string, bool) {
	switch invokeName {
	case XMLInvokeRun, ToolRunShell:
		cmd := params[JSONFieldCommand]
		if cmd == "" {
			return "", false
		}
		return marshalAction(ToolRunShell, JSONFieldCommand, cmd)
	case XMLInvokeAISsh, XMLInvokeAit, ToolAISshCLI, ToolAitCLI:
		args := params[JSONFieldArgs]
		if args == "" {
			return "", false
		}
		return marshalCLIAction(args)
	case ToolAnswer:
		text := params[JSONFieldText]
		if text == "" {
			return "", false
		}
		return marshalAction(ToolAnswer, JSONFieldText, text)
	default:
		return "", false
	}
}

func parseHybridFragment(content string) (string, bool) {
	if m := hybridRunRe.FindStringSubmatch(content); len(m) >= 2 {
		if cmd := strings.TrimSpace(m[1]); cmd != "" {
			return marshalAction(ToolRunShell, JSONFieldCommand, cmd)
		}
	}
	if m := hybridCLIRe.FindStringSubmatch(content); len(m) >= 2 {
		if args := strings.TrimSpace(m[1]); args != "" {
			return marshalCLIAction(args)
		}
	}
	if m := hybridAnswerRe.FindStringSubmatch(content); len(m) >= 2 {
		if text := strings.TrimSpace(m[1]); text != "" {
			return marshalAction(ToolAnswer, JSONFieldText, text)
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
	for _, pm := range xmlOpenParameterRe.FindAllStringSubmatch(content, -1) {
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

func marshalCLIAction(args string) (string, bool) {
	return marshalAction(ToolAISshCLI, JSONFieldArgs, args)
}

func marshalAction(action, field, value string) (string, bool) {
	b, err := json.Marshal(map[string]string{JSONFieldAction: action, field: value})
	if err != nil {
		return "", false
	}
	return string(b), true
}
