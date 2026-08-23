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

// Package scrollback implements a bounded ring buffer that captures terminal
// output for replay (`aissh session --load`) and AI assist context.
package scrollback

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const maxInMemory = 512 * 1024

// Buffer captures terminal output for replay.
type Buffer struct {
	mu   sync.Mutex
	data strings.Builder
}

// Write appends p to the scrollback buffer, dropping oldest data to stay within maxInMemory.
// Returns (len(p), err); n is always len(p) on success so io.Writer callers see normal behavior.
// Errors: strings.Builder write failures (unexpected).
func (b *Buffer) Write(p []byte) (int, error) {
	full := len(p)
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(p) >= maxInMemory {
		b.data.Reset()
		b.data.WriteString(trimUTF8Tail(string(p), maxInMemory))
		return full, nil
	}

	if b.data.Len()+len(p) > maxInMemory {
		s := trimBufferForAppend(b.data.String(), len(p), maxInMemory)
		b.data.Reset()
		b.data.WriteString(s)
	}

	if _, err := b.data.Write(p); err != nil {
		return 0, err
	}
	return full, nil
}

// String returns a copy of all captured terminal output under the buffer lock.
// b must be non-nil.
func (b *Buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data.String()
}

// ReplaceWith replaces buffer contents with fn(s) under the buffer lock.
func (b *Buffer) ReplaceWith(fn func(string) string) {
	if fn == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	s := fn(b.data.String())
	b.data.Reset()
	b.data.WriteString(s)
}

// StripSubstrings removes every occurrence of each non-empty fragment from the buffer.
// Empty fragments in fragments are ignored.
func (b *Buffer) StripSubstrings(fragments ...string) {
	if len(fragments) == 0 {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.data.String()
	for _, frag := range fragments {
		if frag == "" {
			continue
		}
		s = strings.ReplaceAll(s, frag, "")
	}
	b.data.Reset()
	b.data.WriteString(s)
}

// RecentLines returns the last n logical lines of scrollback (trailing newline ignored).
// Returns "" when n <= 0 or the buffer is empty.
func (b *Buffer) RecentLines(n int) string {
	if n <= 0 {
		return ""
	}

	s := b.String()
	if s == "" {
		return ""
	}

	end := len(s)
	if end > 0 && s[end-1] == '\n' {
		end--
	}
	return lastNLinesFrom(s, end, n)
}

// SaveToFile writes the buffer contents to path, creating parent directories (0700) and
// the file with mode 0600. Errors: MkdirAll or WriteFile failures.
func (b *Buffer) SaveToFile(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.data.String()), 0o600)
}
