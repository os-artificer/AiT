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

package pager

import "sync"

// ChanSource is a LineSource fed by a producer goroutine, used when lines
// should appear while the backend is still working. The producer closes Lines
// when it is done, and reports a failure with Fail before closing.
type ChanSource struct {
	Lines <-chan string

	mu  sync.Mutex
	err error
}

// NextLine returns the next produced line, or false once Lines is closed.
func (s *ChanSource) NextLine() (string, bool) {
	line, ok := <-s.Lines
	return line, ok
}

// Err returns the error reported by Fail, if any.
func (s *ChanSource) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Fail records why production stopped. The producer must call it before
// closing Lines; it is safe to call from the producing goroutine while the
// pager reads.
func (s *ChanSource) Fail(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
}
