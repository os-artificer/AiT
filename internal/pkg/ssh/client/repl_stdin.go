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

package client

import "io"

// replStdinReader adapts per-byte stdin from the SSH client forward loop for aimode.ReadLine.
type replStdinReader struct {
	ch  <-chan byte
	buf []byte
}

func newReplStdinReader(ch <-chan byte) *replStdinReader {
	return &replStdinReader{ch: ch}
}

// Read implements io.Reader for one byte per forward from the SSH stdin reader.
func (r *replStdinReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}

	b, ok := <-r.ch
	if !ok {
		return 0, io.EOF
	}
	p[0] = b
	return 1, nil
}
