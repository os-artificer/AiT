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
	"path/filepath"
	"strings"
)

// BuildLogShowPredicate builds an NSPredicate for `log show`/`log stream`
// that narrows eventMessage to q's text filters. An empty string means no
// --predicate should be passed (scan the full window).
//
// When q.Expr is set, the boolean AST is compiled to CONTAINS/AND/OR/NOT.
// Otherwise keywords AND together and Regex uses MATCHES. Filtering remains
// authoritative in Go; this only reduces native volume.
func BuildLogShowPredicate(q Query) string {
	contains := "CONTAINS"
	if q.IgnoreCase {
		contains = "CONTAINS[c]"
	}

	if q.Expr != "" {
		e, err := ParseExpr(q.Expr)
		if err != nil {
			return ""
		}
		return exprToPredicate(e, contains)
	}

	var parts []string
	if q.Regex != "" {
		pat := q.Regex
		if q.IgnoreCase && !strings.HasPrefix(pat, "(?") {
			pat = "(?i)" + pat
		}
		parts = append(parts, "eventMessage MATCHES "+quoteNSPredicate(pat))
	}
	for _, kw := range q.Keywords {
		if kw == "" {
			continue
		}
		parts = append(parts, "eventMessage "+contains+" "+quoteNSPredicate(kw))
	}
	return strings.Join(parts, " AND ")
}

// hasTextFilter reports whether q asks for keyword, regex, or --expr matching.
func hasTextFilter(q Query) bool {
	if q.Expr != "" || q.Regex != "" {
		return true
	}
	for _, kw := range q.Keywords {
		if kw != "" {
			return true
		}
	}
	return false
}

// quoteNSPredicate wraps s as a double-quoted NSPredicate string literal.
func quoteNSPredicate(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' || c == '"' {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	b.WriteByte('"')
	return b.String()
}

// appendLogShowPredicate adds --predicate when q carries text filters.
func appendLogShowPredicate(args []string, q Query) []string {
	pred := BuildLogShowPredicate(q)
	if pred == "" {
		return args
	}
	return append(args, "--predicate", pred)
}

func processBaseName(imagePath string) string {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return ""
	}
	return filepath.Base(imagePath)
}

func applySyslogProcess(remainder string, rec *Record) {
	tag, msg, ok := strings.Cut(remainder, " ")
	if !ok {
		tag = remainder
	}
	unit, pid := parseSyslogTag(tag)
	if unit != "" {
		rec.Unit = unit
		rec.Process = unit
	}
	if pid > 0 {
		rec.PID = pid
	}
	if ok && msg != "" {
		rec.Message = msg
	}
}
