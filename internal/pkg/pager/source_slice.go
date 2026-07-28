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

// SliceSource is a LineSource over an already-collected slice of lines.
type SliceSource struct {
	Lines []string
	i     int
}

// NextLine returns the next line, or false once the slice is exhausted.
func (s *SliceSource) NextLine() (string, bool) {
	if s.i >= len(s.Lines) {
		return "", false
	}
	line := s.Lines[s.i]
	s.i++
	return line, true
}

// Err always returns nil: an in-memory slice cannot fail.
func (s *SliceSource) Err() error { return nil }
