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

// Write appends output and keeps the buffer bounded to maxInMemory bytes.
// The returned n always reflects len(p) so callers behave normally even
// when the underlying buffer drops older bytes.
func (b *Buffer) Write(p []byte) (int, error) {
	full := len(p)
	b.mu.Lock()
	defer b.mu.Unlock()

	// If the incoming chunk alone exceeds the cap, keep only the tail.
	if len(p) >= maxInMemory {
		b.data.Reset()
		b.data.WriteString(trimUTF8Tail(string(p), maxInMemory))
		return full, nil
	}

	// Otherwise drop oldest bytes from current content to fit (data + p).
	if b.data.Len()+len(p) > maxInMemory {
		s := b.data.String()
		keep := maxInMemory - len(p)
		if keep < len(s) {
			s = trimUTF8Tail(s, keep)
		}
		b.data.Reset()
		b.data.WriteString(s)
	}
	if _, err := b.data.Write(p); err != nil {
		return 0, err
	}
	return full, nil
}

// String returns captured output.
func (b *Buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data.String()
}

// StripSubstrings removes every occurrence of each fragment from the buffer.
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

// RecentLines returns the last n lines without splitting the whole buffer.
func (b *Buffer) RecentLines(n int) string {
	if n <= 0 {
		return ""
	}
	s := b.String()
	if s == "" {
		return ""
	}
	count := 0
	end := len(s)
	// Trim trailing newline so it doesn't count as an extra empty line.
	if end > 0 && s[end-1] == '\n' {
		end--
	}
	i := end
	for i > 0 {
		if s[i-1] == '\n' {
			count++
			if count == n {
				return s[i:end]
			}
		}
		i--
	}
	return s[:end]
}

// SaveToFile writes scrollback to path.
func (b *Buffer) SaveToFile(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(b.data.String()), 0o600)
}

// LoadFromFile reads scrollback from path (bounded to maxInMemory).
func LoadFromFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, maxInMemory))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Replay prints scrollback to stdout.
func Replay(data string) {
	if data == "" {
		return
	}
	os.Stdout.WriteString(data)
	if !strings.HasSuffix(data, "\n") {
		os.Stdout.WriteString("\n")
	}
}
