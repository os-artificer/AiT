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
	"fmt"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

var errUnclosedQuote = errhint.WithTry(
	fmt.Errorf("unclosed quote"),
	unclosedQuoteExpl,
	cmdHelp,
)

// parseShellLine splits a REPL line into words, respecting single- and double-quoted segments.
// Quotes are stripped from tokens; spaces inside quotes stay part of the word.
func parseShellLine(line string) ([]string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	var words []string
	var cur strings.Builder
	var quote byte // 0, '\' or '"'
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case (c == '\'' || c == '"') && quote == 0:
			quote = c
		case c == quote:
			quote = 0
		case (c == ' ' || c == '\t') && quote == 0:
			if cur.Len() > 0 {
				words = append(words, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
		}
	}

	if quote != 0 {
		return nil, errUnclosedQuote
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}
	return words, nil
}
