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
	"strconv"
)

var es6UnicodeEscapeRe = regexp.MustCompile(`\\u\{([0-9A-Fa-f]{1,6})\}`)

// NormalizeES6UnicodeJSON replaces ES6-style \u{XXXX} escapes with UTF-8 runes so
// encoding/json can parse provider payloads that use JavaScript unicode syntax.
func NormalizeES6UnicodeJSON(s string) string {
	return es6UnicodeEscapeRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := es6UnicodeEscapeRe.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}

		cp, err := strconv.ParseInt(sub[1], 16, 32)
		if err != nil {
			return m
		}

		return string(rune(cp))
	})
}
