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

package scrollback

import (
	"bytes"
	"strings"
	"testing"
)

func TestBuffer_BoundedAfterSmallWrites(t *testing.T) {
	var b Buffer
	chunk := bytes.Repeat([]byte("x"), 10*1024)
	for i := 0; i < 100; i++ {
		if _, err := b.Write(chunk); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if got := len(b.String()); got > maxInMemory {
		t.Fatalf("len=%d > maxInMemory=%d", got, maxInMemory)
	}
}

func TestBuffer_BoundedWhenSingleWriteExceedsCap(t *testing.T) {
	var b Buffer
	big := bytes.Repeat([]byte("y"), maxInMemory+5_000)
	n, err := b.Write(big)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(big) {
		t.Fatalf("Write returned %d, want %d", n, len(big))
	}
	if got := len(b.String()); got != maxInMemory {
		t.Fatalf("len=%d, want maxInMemory=%d", got, maxInMemory)
	}
}

func TestBuffer_RecentLines(t *testing.T) {
	var b Buffer
	lines := []string{"line1", "line2", "line3", "line4", "line5"}
	if _, err := b.Write([]byte(strings.Join(lines, "\n") + "\n")); err != nil {
		t.Fatalf("Write: %v", err)
	}
	got := b.RecentLines(2)
	want := "line4\nline5"
	if got != want {
		t.Fatalf("RecentLines(2)=%q want %q", got, want)
	}
	if b.RecentLines(0) != "" {
		t.Fatalf("RecentLines(0) should be empty")
	}
	if b.RecentLines(100) != strings.Join(lines, "\n") {
		t.Fatalf("RecentLines(100) mismatch")
	}
}

func TestBuffer_RecentLinesEmpty(t *testing.T) {
	var b Buffer
	if got := b.RecentLines(5); got != "" {
		t.Fatalf("empty buffer should yield empty, got %q", got)
	}
}
