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

// Package scrollback contains receiver-free helpers that compute new output
// captured from the remote PTY scrollback buffer. The PTY capture entry-point
// (Client.RunPTYCapture) stays in the parent ssh/client package because it
// touches Client internals; this package owns the pure delta arithmetic and
// the capture-related timing constants.
package scrollback

import (
	"strings"
	"time"
)

// CaptureQuiet is the idle period after writing a command before reading the delta.
const CaptureQuiet = 120 * time.Millisecond

// CaptureMaxWait is the maximum time RunPTYCapture waits for the remote shell to fall idle.
const CaptureMaxWait = 8 * time.Second

// Delta returns the suffix of after not present in before, trimmed of leading/trailing whitespace.
// When after equals before it returns "".
func Delta(before, after string) string {
	if after == before {
		return ""
	}
	if strings.HasPrefix(after, before) {
		return strings.TrimSpace(after[len(before):])
	}
	return strings.TrimSpace(after)
}
