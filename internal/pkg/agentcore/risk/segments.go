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

// segmentOperators are the unquoted separators that end a command segment.
// Two-byte operators must precede their single-byte prefixes.
var segmentOperators = []string{"&&", "||", ";;", ";", "|", "\n"}

// unsafeReadMeta are metacharacters that disqualify a segment from the
// read-only fast path: they can redirect output into a file, run a second
// command, or background it, none of which the first-token read check sees.
var unsafeReadMeta = []string{">", "<", "$(", "`", "&"}

// benignRedirects cannot modify user data, so they neither disqualify the
// read-only fast path nor contribute to a high-risk match.
var benignRedirects = []string{"2>&1", "1>&2", ">&2", "> /dev/null", ">/dev/null"}

const (
	quoteSingle = '\''
	quoteDouble = '"'
	escapeByte  = '\\'

	// maxSegments bounds pathological command lines.
	maxSegments = 64
)

// SplitSegments splits cmd into individually classifiable command segments at
// unquoted shell operators (`;`, `;;`, `&&`, `||`, `|`, newline).
// Text inside single or double quotes is never split, and backslash escapes are
// preserved outside single quotes.
// Command substitutions are deliberately not parsed; segments containing
// substitution or redirection metacharacters are reported by hasUnsafeReadMeta
// and never take the read-only fast path.
// Returns nil for blank input, and at most maxSegments segments.
func SplitSegments(cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	var out []string
	var buf strings.Builder
	quote := byte(0)

	for i := 0; i < len(cmd); {
		width, isOperator := scanSegmentByte(cmd, i, &quote, &buf)
		i += width
		if !isOperator {
			continue
		}

		out = appendSegment(out, buf.String())
		buf.Reset()
		if len(out) >= maxSegments {
			return out
		}
	}
	return appendSegment(out, buf.String())
}

// scanSegmentByte consumes one quote-aware unit of s at index i, appending it to
// buf unless it is a segment operator. quote tracks the open quote byte, if any.
// Returns the number of bytes consumed and whether a segment operator was found.
func scanSegmentByte(s string, i int, quote *byte, buf *strings.Builder) (width int, isOperator bool) {
	c := s[i]
	if c == escapeByte && *quote != quoteSingle && i+1 < len(s) {
		buf.WriteByte(c)
		buf.WriteByte(s[i+1])
		return 2, false
	}
	if *quote != 0 {
		if c == *quote {
			*quote = 0
		}
		buf.WriteByte(c)
		return 1, false
	}
	if c == quoteSingle || c == quoteDouble {
		*quote = c
		buf.WriteByte(c)
		return 1, false
	}

	if op := matchSegmentOperator(s, i); op != "" {
		return len(op), true
	}
	buf.WriteByte(c)
	return 1, false
}

func matchSegmentOperator(s string, i int) string {
	for _, op := range segmentOperators {
		if strings.HasPrefix(s[i:], op) {
			return op
		}
	}
	return ""
}

func appendSegment(out []string, seg string) []string {
	seg = strings.TrimSpace(seg)
	if seg == "" {
		return out
	}
	return append(out, seg)
}

// hasUnsafeReadMeta reports whether seg contains unquoted metacharacters that
// can write a file or run another command. Quoted text and benign redirections
// (fd duplication, /dev/null) are ignored.
func hasUnsafeReadMeta(seg string) bool {
	text := stripBenignRedirects(stripQuoted(seg))
	for _, meta := range unsafeReadMeta {
		if strings.Contains(text, meta) {
			return true
		}
	}
	return false
}

// stripQuoted removes the contents of quoted spans so metacharacter scans only
// see unquoted text. The quote bytes themselves are dropped as well.
func stripQuoted(s string) string {
	var b strings.Builder
	quote := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == escapeByte && quote != quoteSingle:
			i++
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == quoteSingle || c == quoteDouble:
			quote = c
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func stripBenignRedirects(s string) string {
	for _, r := range benignRedirects {
		s = strings.ReplaceAll(s, r, " ")
	}
	return s
}
