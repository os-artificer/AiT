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
	"bytes"
	"strings"
	"testing"
)

func TestNormalizeLine(t *testing.T) {
	if got := NormalizeLine("  list;  "); got != "list" {
		t.Errorf("got %q want list", got)
	}
	if got := NormalizeLine("help"); got != "help" {
		t.Errorf("got %q want help", got)
	}
}

func TestCollapseTTYMsg(t *testing.T) {
	msg := collapseTTYMsg("line one\nline two")
	if strings.Contains(msg, "\n") {
		t.Errorf("want single line, got %q", msg)
	}
	if !strings.Contains(msg, "line one") || !strings.Contains(msg, "line two") {
		t.Errorf("got %q", msg)
	}
}

func TestWriteTTY_LFToCRLF(t *testing.T) {
	var buf bytes.Buffer
	WriteTTY(&buf, "a\nb")
	if buf.String() != "a\r\nb" {
		t.Fatalf("got %q", buf.String())
	}
}
