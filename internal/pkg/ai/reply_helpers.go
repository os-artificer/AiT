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

package ai

import (
	"regexp"
	"strings"
)

// Build regex patterns at init time to avoid issues with tag-like content
// in source code being misinterpreted.
var (
	thinkingRe     = regexp.MustCompile(`(?s)` + thinkOpen + `(.*?)` + thinkClose)
	xmlToolCallRe  = regexp.MustCompile(`(?s)<minimax:tool_call>.*?</minimax:tool_call>`)
	xmlInvokeRe    = regexp.MustCompile(`(?s)<invoke\b[^>]*>.*?</invoke>`)
	xmlParameterRe = regexp.MustCompile(`(?s)<parameter\b[^>]*>.*?</parameter>`)
)

const (
	// Reasoning tag markers used by MiniMax M2.x models.
	thinkOpen  = "\x3Cthink\x3E"
	thinkClose = "\x3C/think\x3E"
)

// ExtractReply splits raw LLM response content into a clean reply and optional
// thinking text. It strips:
//   - Reasoning tags from MiniMax M2.x and similar providers, captured into
//     the thinking return value (first match only; subsequent blocks are
//     stripped from reply but not appended to thinking).
//   - XML-style tool call blocks that providers embed in content instead of
//     the proper tool_calls field.
//
// DeepSeek reasoner uses a separate reasoning_content field (handled by the
// caller), so this function only deals with in-content tags.
//
// Parameters:
//   - raw: the raw content string from the LLM response message.
//
// Returns:
//   - reply: user-visible text with all non-user-facing tags removed.
//   - thinking: captured reasoning content from the first reasoning tag, or ""
//     when no tags are present.
//
// Returns ("", "") for empty input.
func ExtractReply(raw string) (reply, thinking string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}

	// Extract thinking content from reasoning tags.
	matches := thinkingRe.FindStringSubmatch(raw)
	if len(matches) >= 2 {
		thinking = strings.TrimSpace(matches[1])
	}

	// Strip all non-user-facing tag blocks from content.
	cleaned := thinkingRe.ReplaceAllString(raw, "")
	cleaned = xmlToolCallRe.ReplaceAllString(cleaned, "")
	cleaned = xmlInvokeRe.ReplaceAllString(cleaned, "")
	cleaned = xmlParameterRe.ReplaceAllString(cleaned, "")
	reply = strings.TrimSpace(cleaned)

	return reply, thinking
}
