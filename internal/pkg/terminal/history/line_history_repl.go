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

package history

import (
	"fmt"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// NewReplLineHistory creates a history loaded from the default persistence file.
// Load errors are reported to warn when non-nil; REPL continues with empty history.
func NewReplLineHistory(warn io.Writer) *LineHistory {
	h := NewLineHistory()
	path, err := platform.HistoryPath()
	if err != nil {
		writeHistoryWarn(warn, err)
		return h
	}
	h.SetFilePath(path)
	if err := h.LoadFromFile(path); err != nil {
		writeHistoryWarn(warn, err)
	}
	return h
}

func writeHistoryWarn(w io.Writer, err error) {
	if w == nil || err == nil {
		return
	}
	_, _ = fmt.Fprintf(w, "AiT: warning: history: %v\r\n", err)
}
