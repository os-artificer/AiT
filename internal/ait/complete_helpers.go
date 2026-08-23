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

package ait

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// completeFirstToken completes the first command token against names.
// names are sorted and deduplicated before matching. cursor is a rune index into line.
func completeFirstToken(names []string, line string, cursor int) (newLine string, newCursor int, display string) {
	names = uniqueSortedNames(names)
	runes := []rune(line)
	cursor = clampRuneCursor(cursor, len(runes))

	tok, ok := findFirstToken(runes, cursor)
	if !ok {
		return line, cursor, ""
	}

	prefix := string(runes[tok.start:min(cursor, tok.end)])
	matches := matchPrefix(names, prefix)
	switch len(matches) {
	case 0:
		return line, cursor, ""
	case 1:
		return replaceFirstTokenUnique(runes, tok, matches[0], cursor)
	default:
		return replaceFirstTokenLCPOrList(runes, tok, prefix, matches, cursor)
	}
}

type firstToken struct {
	start int // rune index of first non-space of token
	end   int // rune index after last rune of token
}

// findFirstToken locates the first unquoted token. Returns ok=false for no-op cases:
// cursor after first token, cursor in leading space with a token present, or quotes in token.
func findFirstToken(runes []rune, cursor int) (firstToken, bool) {
	i := 0
	for i < len(runes) && isTokenSpace(runes[i]) {
		i++
	}
	if i >= len(runes) {
		// Empty or whitespace-only: empty prefix at cursor (list all).
		return firstToken{start: i, end: i}, true
	}
	if cursor < i {
		return firstToken{}, false
	}

	start := i
	for i < len(runes) {
		r := runes[i]
		if r == '\'' || r == '"' {
			return firstToken{}, false
		}
		if isTokenSpace(r) {
			break
		}
		i++
	}
	end := i
	if cursor > end {
		return firstToken{}, false
	}
	return firstToken{start: start, end: end}, true
}

func replaceFirstTokenUnique(runes []rune, tok firstToken, match string, cursor int) (string, int, string) {
	hasArgs := tokenHasTrailingArgs(runes, tok.end)
	var out string
	var newCursor int
	if hasArgs {
		out = replaceTokenSpan(runes, tok, match)
		newCursor = tok.start + utf8.RuneCountInString(match)
	} else {
		out = string(runes[:tok.start]) + match + " "
		newCursor = utf8.RuneCountInString(out)
	}

	line := string(runes)
	if out == line && newCursor == cursor {
		return line, cursor, ""
	}
	return out, newCursor, ""
}

func replaceFirstTokenLCPOrList(runes []rune, tok firstToken, prefix string, matches []string, cursor int) (string, int, string) {
	lcp := longestCommonPrefix(matches)
	if len(lcp) > len(prefix) {
		out := replaceTokenSpan(runes, tok, lcp)
		newCursor := tok.start + utf8.RuneCountInString(lcp)
		return out, newCursor, ""
	}
	return string(runes), cursor, strings.Join(matches, "\n")
}

func replaceTokenSpan(runes []rune, tok firstToken, replacement string) string {
	var b strings.Builder
	b.WriteString(string(runes[:tok.start]))
	b.WriteString(replacement)
	b.WriteString(string(runes[tok.end:]))
	return b.String()
}

func tokenHasTrailingArgs(runes []rune, tokenEnd int) bool {
	for i := tokenEnd; i < len(runes); i++ {
		if !isTokenSpace(runes[i]) {
			return true
		}
	}
	return false
}

func matchPrefix(names []string, prefix string) []string {
	var out []string
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			out = append(out, name)
		}
	}
	return out
}

func longestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	prefix := strs[0]
	for _, s := range strs[1:] {
		for !strings.HasPrefix(s, prefix) {
			if prefix == "" {
				return ""
			}
			prefix = prefix[:len(prefix)-1]
		}
	}
	return prefix
}

func uniqueSortedNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	cp := append([]string(nil), names...)
	sort.Strings(cp)
	out := cp[:0]
	var prev string
	for i, name := range cp {
		if i > 0 && name == prev {
			continue
		}
		out = append(out, name)
		prev = name
	}
	return out
}

func isTokenSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func clampRuneCursor(cursor, n int) int {
	if cursor < 0 {
		return 0
	}
	if cursor > n {
		return n
	}
	return cursor
}
