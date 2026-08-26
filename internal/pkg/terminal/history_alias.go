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

package terminal

import (
	"io"

	"github.com/os-artificer/AiT/internal/pkg/terminal/history"
)

// LineHistory is the REPL input history (alias of history.LineHistory) preserved
// for backwards compatibility with existing importers (shell / agent).
type LineHistory = history.LineHistory

// NewLineHistory returns an empty REPL input history with the default capacity.
func NewLineHistory() *LineHistory {
	return history.NewLineHistory()
}

// NewReplLineHistory creates a history loaded from the default persistence file.
// Load errors are reported to warn when non-nil; REPL continues with empty history.
func NewReplLineHistory(warn io.Writer) *LineHistory {
	return history.NewReplLineHistory(warn)
}

// RunHistoryCommand executes bash-style history subcommands on h and writes listing to w.
func RunHistoryCommand(w io.Writer, h *LineHistory, args []string) error {
	return history.RunHistoryCommand(w, h, args)
}
