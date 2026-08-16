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

package logsrc

import (
	"regexp"
	"strings"
)

// journalGrepPattern builds a PCRE for journalctl -g that AND-matches every
// keyword. A single keyword is QuoteMeta only; multiple use lookaheads so order
// in the message does not matter.
func journalGrepPattern(q Query) string {
	kws := nonEmptyKeywords(q.Keywords)
	if len(kws) == 0 {
		return ""
	}
	if len(kws) == 1 {
		return regexp.QuoteMeta(kws[0])
	}

	var b strings.Builder
	for _, kw := range kws {
		b.WriteString("(?=.*")
		b.WriteString(regexp.QuoteMeta(kw))
		b.WriteString(")")
	}
	b.WriteString(".*")
	return b.String()
}

// appendJournalGrep adds -g when q has a regex or keywords, so journalctl
// returns a smaller stream before Go applies the authoritative Filter.
// --expr is skipped: journalctl -g cannot express AND/OR/NOT safely.
func appendJournalGrep(args []string, q Query) []string {
	if q.Expr != "" {
		return args
	}
	pat := q.Regex
	if pat == "" {
		pat = journalGrepPattern(q)
	}
	if pat == "" {
		return args
	}
	args = append(args, "-g", pat)
	if !q.IgnoreCase {
		args = append(args, "--case-sensitive")
	}
	return args
}

func nonEmptyKeywords(kws []string) []string {
	out := make([]string, 0, len(kws))
	for _, kw := range kws {
		if kw != "" {
			out = append(out, kw)
		}
	}
	return out
}
