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
	"regexp"
	"strings"
)

var runAitEnvelopeRe = regexp.MustCompile(
	`(?s)\{\s*"action"\s*:\s*"(?:` + XMLInvokeRun + `|` + ToolRunShell + `|` +
		XMLInvokeAISsh + `|` + XMLInvokeAit + `|` + ToolAISshCLI + `|` + ToolAitCLI + `)"`,
)

// HasRunOrAitEnvelopeSignal reports whether content looks like a broken run/ait envelope (C17).
// It matches agent JSON/XML envelope shapes, not bare JSON field names inside prose.
func HasRunOrAitEnvelopeSignal(content string) bool {
	if runAitEnvelopeRe.MatchString(content) {
		return true
	}

	lower := strings.ToLower(content)
	for _, name := range []string{
		XMLInvokeRun, ToolRunShell, XMLInvokeAISsh, XMLInvokeAit, ToolAISshCLI, ToolAitCLI,
	} {
		if strings.Contains(lower, `name="`+name+`"`) {
			return true
		}
	}
	return false
}

// FormatCorrectionPrompt builds the S2 retry user message including the original question (C15).
// originalUser is the user's initial Agent turn text; reply is the malformed model output; errs
// lists parse/validation failures from ResolveAgentReply.
func FormatCorrectionPrompt(originalUser, reply string, errs []error) string {
	var b strings.Builder
	b.WriteString(formatCorrectionIntro)
	b.WriteString(originalUser)
	b.WriteString(formatCorrectionReplyLabel)
	b.WriteString(truncateForPrompt(reply, promptExcerptMaxRunes))
	b.WriteString(formatCorrectionErrLabel)
	b.WriteString(FormatErrorSummary(errs))
	return b.String()
}

// FormatToolResultError builds a tool_result format_error message for step>0 recovery (C16).
// reply is the malformed assistant content that failed ResolveAgentReply.
func FormatToolResultError(reply string, result ParseResult) string {
	return ToolResultPrefix + FormatErrorPrefix + FormatErrorSummary(result.Errs) +
		"\nexcerpt:\n" + truncateForPrompt(reply, toolErrExcerptRunes)
}

// FormatErrorSummary joins parse errors for prompts and tool_result feedback.
// Returns "invalid agent response format" when errs is empty; otherwise joins non-nil err.Error() with "; ".
// When errs is non-empty but every entry is nil, returns an empty string.
func FormatErrorSummary(errs []error) string {
	if len(errs) == 0 {
		return "invalid agent response format"
	}

	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			parts = append(parts, err.Error())
		}
	}
	return strings.Join(parts, "; ")
}

func truncateForPrompt(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes]) + "…"
}
