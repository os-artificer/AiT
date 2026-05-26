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

package history

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestHistoryCommand_List(t *testing.T) {
	h := NewLineHistory()
	h.Add("list")
	h.Add("help")

	var buf bytes.Buffer
	if err := RunHistoryCommand(&buf, h, nil); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "      1  list\n") || !strings.Contains(got, "      2  help\n") {
		t.Fatalf("output = %q", got)
	}
}

func TestHistoryCommand_LastN(t *testing.T) {
	h := NewLineHistory()
	h.Add("a")
	h.Add("b")
	h.Add("c")

	var buf bytes.Buffer
	if err := RunHistoryCommand(&buf, h, []string{"2"}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, "a") || !strings.Contains(got, "b") || !strings.Contains(got, "c") {
		t.Fatalf("output = %q, want last 2 only", got)
	}
}

func TestHistoryCommand_Clear(t *testing.T) {
	h := NewLineHistory()
	h.Add("list")
	if err := RunHistoryCommand(io.Discard, h, []string{"-c"}); err != nil {
		t.Fatal(err)
	}
	if h.Len() != 0 {
		t.Fatalf("len = %d, want 0", h.Len())
	}
}

func TestHistoryCommand_Delete(t *testing.T) {
	h := NewLineHistory()
	h.Add("first")
	h.Add("second")
	if err := RunHistoryCommand(io.Discard, h, []string{"-d", "1"}); err != nil {
		t.Fatal(err)
	}
	if h.Len() != 1 || h.At(0) != "second" {
		t.Fatalf("lines = [%q], want [second]", h.At(0))
	}
}

func TestHistoryCommand_SetLine(t *testing.T) {
	h := NewLineHistory()
	if err := RunHistoryCommand(io.Discard, h, []string{"-s", "seeded"}); err != nil {
		t.Fatal(err)
	}
	if h.Len() != 1 || h.At(0) != "seeded" {
		t.Fatalf("lines = [%q], want [seeded]", h.At(0))
	}
}
