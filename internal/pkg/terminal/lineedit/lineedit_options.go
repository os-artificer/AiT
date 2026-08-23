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

package lineedit

import (
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/terminal/history"
)

// Options configures ReadLine behavior in raw TTY mode.
// Prompt is redrawn after Ctrl+L and mid-line edits. History enables Up/Down recall when non-nil.
// RecordHistory appends submitted non-empty lines to History when true (see shouldRecordLine filters).
// Completer, when non-nil, handles Tab instead of inserting a tab character (e.g. ait REPL).
type Options struct {
	Prompt        string
	History       *history.LineHistory
	RecordHistory bool
	// Completer handles Tab when non-nil. line is the full buffer; cursor is a rune index.
	// display is LF-separated text to print above the redrawn prompt (empty = no list).
	// display must use only '\n' line endings (no '\r'); applyCompleter converts LF to CRLF.
	Completer func(line string, cursor int) (newLine string, newCursor int, display string)
}

func shouldRecordLine(line string) bool {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	if strings.HasPrefix(line, "export AIT_") ||
		strings.HasPrefix(line, "PS1=") ||
		strings.HasPrefix(line, "PROMPT=") ||
		strings.HasPrefix(line, "ZDOTDIR=") {
		return false
	}
	return true
}
