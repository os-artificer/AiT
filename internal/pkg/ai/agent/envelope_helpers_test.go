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
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateForPrompt_UTF8(t *testing.T) {
	glyph := "报"
	long := strings.Repeat(glyph, toolErrExcerptRunes+5)

	got := truncateForPrompt(long, toolErrExcerptRunes)
	if !utf8.ValidString(got) {
		t.Fatalf("invalid UTF-8: %q", got)
	}

	runes := []rune(got)
	if len(runes) != toolErrExcerptRunes+1 {
		t.Fatalf("len runes=%d want %d", len(runes), toolErrExcerptRunes+1)
	}
	if runes[toolErrExcerptRunes] != '…' {
		t.Fatalf("last rune=%q want ellipsis", runes[toolErrExcerptRunes])
	}
}
