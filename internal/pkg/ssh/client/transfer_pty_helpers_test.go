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

func TestBuildPutChunkCmd_Heredoc(t *testing.T) {
	tmp := "/tmp/ait-put.abc123"
	script := buildPutChunkScript(tmp, "aGVsbG8=")
	if !strings.Contains(script, "base64 -d >>") {
		t.Errorf("buildPutChunkScript missing base64 append: %q", script)
	}
	if !strings.Contains(script, ptyPutChunkHeredocDelim) {
		t.Errorf("buildPutChunkScript missing heredoc delimiter: %q", script)
	}
	if strings.Contains(script, "printf") {
		t.Errorf("buildPutChunkScript should not use printf: %q", script)
	}
	if strings.Contains(script, "./test") {
		t.Errorf("buildPutChunkScript must not reference final remote path in chunk: %q", script)
	}
	if !strings.Contains(script, shellQuoteArg(tmp)) {
		t.Errorf("buildPutChunkScript missing tmp path: %q", script)
	}
}

func TestBuildPutMvScript(t *testing.T) {
	cmd := buildPutMvScript("/tmp/staging", "/remote/target")
	if !strings.Contains(cmd, "mv ") {
		t.Fatalf("buildPutMvScript missing mv: %q", cmd)
	}
	if !strings.Contains(cmd, shellQuoteArg("/tmp/staging")) {
		t.Fatalf("buildPutMvScript missing tmp: %q", cmd)
	}
	if !strings.Contains(cmd, shellQuoteArg("/remote/target")) {
		t.Fatalf("buildPutMvScript missing target: %q", cmd)
	}
}

func TestRemotePutContentHasInjectLeak(t *testing.T) {
	if !remotePutContentHasInjectLeak([]byte("noise stty -echo more")) {
		t.Fatal("expected stty -echo leak")
	}
	if !remotePutContentHasInjectLeak([]byte("AIT_CAPTURE_B_f3e7a1c9")) {
		t.Fatal("expected capture marker leak")
	}
	if remotePutContentHasInjectLeak([]byte("hello world")) {
		t.Fatal("expected clean payload")
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
