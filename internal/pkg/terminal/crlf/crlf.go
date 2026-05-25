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

// Package crlf provides an io.Writer adapter that translates LF to CRLF so
// shell/meta output renders correctly on raw TTYs and remote PTYs.
package crlf

import (
	"io"
)

type writer struct {
	w      io.Writer
	prevCR bool
}

// NewWriter returns an io.Writer that translates a bare LF to CRLF before
// writing to w. An LF already preceded by CR (whether in the same Write or
// across consecutive Writes) is left untouched, so wrapping an already
// CRLF-correct stream is a safe no-op (idempotent under composition).
// Write returns len(p) on success so callers see the original byte count
// (TTY meta/shell output expects exact byte accounting).
func NewWriter(w io.Writer) io.Writer {
	return &writer{w: w}
}

// Write implements io.Writer.
func (c *writer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	buf := make([]byte, 0, len(p)+len(p)/8)
	for _, ch := range p {
		if ch == '\n' && !c.prevCR {
			buf = append(buf, '\r')
		}
		buf = append(buf, ch)
		c.prevCR = ch == '\r'
	}
	_, err := c.w.Write(buf)
	return len(p), err
}
