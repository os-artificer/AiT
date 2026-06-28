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
	"errors"
	"strings"
)

// ResolveAgentReply is the SSOT parse + S1 fallback chain for Agent-mode LLM replies.
// content is raw assistant text after Eino post-processing (may include partial JSON).
// Order: strict JSON, XML tool fallback, answer-prefix heuristic, run/ait guard, prose.
// Returns ParseResult with Source set when a strategy matched; Errs populated when all strategies fail.
// Prose fallback wraps plain text as action answer when no envelope signals are detected.
func ResolveAgentReply(content string) ParseResult {
	content = strings.TrimSpace(content)
	if content == "" {
		return ParseResult{Errs: []error{errors.New("empty content")}}
	}

	if env, ok := parseEnvelopeStrict(content); ok {
		return finishParse(env, SourceJSON)
	}

	if jsonStr, ok := XMLToolCallToJSON(content); ok {
		if env, ok := parseEnvelopeJSON(jsonStr); ok {
			return finishParse(env, SourceXML)
		}
	}

	if env, ok := extractAnswerPrefixHeuristic(content); ok {
		return finishParse(env, SourceHeuristic)
	}

	if HasRunOrAitEnvelopeSignal(content) {
		return ParseResult{Errs: []error{errors.New("run/ait envelope detected but parse failed")}}
	}

	if looksLikeFailedEnvelope(content) {
		return ParseResult{Errs: []error{errors.New("envelope parse failed")}}
	}

	env := AgentEnvelope{Action: ToolAnswer, Text: content}
	return finishParse(env, SourceProse)
}
