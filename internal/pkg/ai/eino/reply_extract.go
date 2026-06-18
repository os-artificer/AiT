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

package eino

import (
	"regexp"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/ai/agent"
)

var (
	thinkingRe     = regexp.MustCompile(`(?s)` + thinkOpen + `(.*?)` + thinkClose)
	xmlToolCallRe  = regexp.MustCompile(`(?s)<minimax:tool_call>.*?</minimax:tool_call>`)
	xmlInvokeRe    = regexp.MustCompile(`(?s)<invoke\b[^>]*>.*?</invoke>`)
	xmlParameterRe = regexp.MustCompile(`(?s)<parameter\b[^>]*>.*?</parameter>`)
	orphanCloseRe  = regexp.MustCompile(`(?s)</(?:minimax:tool_call|invoke|parameter)>`)
	hybridParamRe  = regexp.MustCompile(`(?s)"(?:` + agent.JSONFieldCommand + `|` + agent.JSONFieldArgs + `|` + agent.JSONFieldText + `)"\s*>\s*[^<]*</parameter>`)
	partialJSONRe  = regexp.MustCompile(`(?s)\{"` + agent.JSONFieldAction + `":"(?:` + agent.ActionNamePattern + `)"[^}]*\z`)
)

const (
	thinkOpen  = "\x3Cthink\x3E"
	thinkClose = "\x3C/think\x3E"
)

// ExtractReply splits raw LLM response content into a clean reply and optional thinking text.
// Semantics match internal/pkg/ai/client/reply_helpers.go (Phase 1 will delegate client to here).
func ExtractReply(raw string) (reply, thinking string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}

	matches := thinkingRe.FindStringSubmatch(raw)
	if len(matches) >= 2 {
		thinking = strings.TrimSpace(matches[1])
	}

	cleaned := thinkingRe.ReplaceAllString(raw, "")
	cleaned = xmlToolCallRe.ReplaceAllString(cleaned, "")
	cleaned = xmlInvokeRe.ReplaceAllString(cleaned, "")
	cleaned = xmlParameterRe.ReplaceAllString(cleaned, "")
	cleaned = orphanCloseRe.ReplaceAllString(cleaned, "")
	cleaned = hybridParamRe.ReplaceAllString(cleaned, "")
	cleaned = partialJSONRe.ReplaceAllString(cleaned, "")
	reply = strings.TrimSpace(cleaned)

	return reply, thinking
}
