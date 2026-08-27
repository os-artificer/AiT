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

package risk

import "strings"

// classifySegment classifies one command segment. The high-risk denylist is
// checked before the read-only fast path so a destructive segment can never be
// auto-approved because of its first token.
func classifySegment(seg string) Level {
	if IsInteractive(seg) {
		return Interactive
	}
	if matchesHighRisk(seg) {
		return High
	}
	if IsRead(seg) && !hasUnsafeReadMeta(seg) {
		return Read
	}
	return Low
}

// matchesHighRisk reports whether text contains a denylisted destructive
// fragment. Benign redirections are stripped first so writing to /dev/null does
// not match the "> /dev/" device-write fragment.
func matchesHighRisk(text string) bool {
	lower := stripBenignRedirects(strings.ToLower(text))
	for _, frag := range highRiskFragments {
		if strings.Contains(lower, frag) {
			return true
		}
	}
	return false
}

// matchesPipeHighRisk reports whether the whole command line pipes into a shell.
// These fragments span a segment operator and so cannot be found per segment.
func matchesPipeHighRisk(cmd string) bool {
	lower := strings.ToLower(cmd)
	for _, frag := range pipeHighRiskFragments {
		if strings.Contains(lower, frag) {
			return true
		}
	}
	return false
}

// severity orders levels for worst-case aggregation; the Level constants are
// declared in parse order, not in risk order.
func severity(l Level) int {
	switch l {
	case Read:
		return 0
	case High:
		return 2
	case Interactive:
		return 3
	default:
		return 1
	}
}

// moreSevere returns whichever level requires the stricter gate.
func moreSevere(a, b Level) Level {
	if severity(b) > severity(a) {
		return b
	}
	return a
}
