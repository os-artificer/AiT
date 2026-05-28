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

import (
	"strings"
	"testing"
)

func TestChunkString(t *testing.T) {
	got := chunkString("abcdef", 3)
	if len(got) != 2 {
		t.Fatalf("chunkString = %v, want 2 chunks", got)
	}
	if got[0] != "abc" {
		t.Errorf("chunk[0] = %q, want %q", got[0], "abc")
	}
	if got[1] != "def" {
		t.Errorf("chunk[1] = %q, want %q", got[1], "def")
	}
}

func TestChunkString_ShortString(t *testing.T) {
	got := chunkString("ab", 10)
	if len(got) != 1 {
		t.Fatalf("chunkString = %v, want 1 chunk", got)
	}
	if got[0] != "ab" {
		t.Errorf("chunk[0] = %q, want %q", got[0], "ab")
	}
}

func TestChunkString_ExactMultiple(t *testing.T) {
	got := chunkString("abcd", 2)
	if len(got) != 2 {
		t.Fatalf("chunkString = %v, want 2 chunks", got)
	}
}

func TestBuildPutChunkCmd(t *testing.T) {
	cmd := buildPutChunkCmd("/tmp/file.txt", "aGVsbG8=")
	if !strings.Contains(cmd, "base64 -d") {
		t.Errorf("buildPutChunkCmd missing base64 -d: %q", cmd)
	}
	if !strings.Contains(cmd, ">> ") {
		t.Errorf("buildPutChunkCmd missing append: %q", cmd)
	}
	if !strings.Contains(cmd, "printf") {
		t.Errorf("buildPutChunkCmd missing printf: %q", cmd)
	}
}

func TestShellQuoteArg(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "'hello'"},
		{"it's", "'it'\"'\"'s'"},
		{"", "''"},
	}

	for _, tt := range tests {
		got := shellQuoteArg(tt.input)
		if got != tt.want {
			t.Errorf("shellQuoteArg(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLocalFileHash(t *testing.T) {
	// sha256 of empty string is well-known.
	got := localFileHash([]byte{})
	want := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if got != want {
		t.Errorf("localFileHash(empty) = %q, want %q", got, want)
	}
}
