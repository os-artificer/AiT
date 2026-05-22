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
	"strings"
)

type crlfWriter struct {
	w io.Writer
}

// NewCRLFWriter returns a Writer that translates LF to CRLF before writing to w.
// Write returns len(p) on success so callers see the original byte count (TTY meta/shell output).
func NewCRLFWriter(w io.Writer) io.Writer {
	return &crlfWriter{w: w}
}

// Write implements io.Writer.
func (c *crlfWriter) Write(p []byte) (int, error) {
	s := strings.ReplaceAll(string(p), "\n", "\r\n")

	_, err := c.w.Write([]byte(s))
	return len(p), err
}
