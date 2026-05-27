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
	"testing"
	"unicode/utf8"
)

func TestTruncate_UTF8Safe(t *testing.T) {
	// Each rune is 3 bytes in UTF-8; a naive byte slice would split mid-rune.
	s := "你好世界🌍终端输出"
	const max = 12
	got := truncate(s, max)
	if len(got) > max {
		t.Fatalf("len=%d want <= %d", len(got), max)
	}
	if !utf8.ValidString(got) {
		t.Fatalf("invalid UTF-8: %q", got)
	}
	if got == "" {
		t.Fatal("expected non-empty tail")
	}
}

func TestTruncate_ShortStringUnchanged(t *testing.T) {
	s := "hello"
	if got := truncate(s, 10); got != s {
		t.Fatalf("got %q want %q", got, s)
	}
}

func TestTruncate_ASCII(t *testing.T) {
	s := "abcdefghijklmnopqrstuvwxyz"
	got := truncate(s, 5)
	if got != "vwxyz" {
		t.Fatalf("got %q want %q", got, "vwxyz")
	}
}
