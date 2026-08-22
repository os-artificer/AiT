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

package shell

import (
	"bytes"
	"strings"
	"testing"
)

func TestNormalizeShellLine(t *testing.T) {
	if got := normalizeShellLine("  list;  "); got != "list" {
		t.Errorf("got %q want list", got)
	}
	if got := normalizeShellLine("help"); got != "help" {
		t.Errorf("got %q want help", got)
	}
}

func TestShellPrintErrMultiline(t *testing.T) {
	var buf bytes.Buffer
	msg := "session not found\n\nThe list ID column is for display only.\n\nTry:\n  aissh session --load"
	shellPrintErr(&buf, msg)
	out := buf.String()
	if strings.Count(out, "AiT:") != 1 {
		t.Fatalf("duplicate prefix: %q", out)
	}
	if !strings.Contains(out, "Try:") {
		t.Fatalf("missing Try: %q", out)
	}
	if strings.Contains(out, "display only AiT:") {
		t.Fatalf("collapsed: %q", out)
	}
}

func TestShellPrintErr(t *testing.T) {
	var buf bytes.Buffer
	shellPrintErr(&buf, "unknown command \";\"")
	out := buf.String()
	if !strings.HasPrefix(out, "AiT: ") {
		t.Errorf("prefix missing: %q", out)
	}
	if strings.Count(out, "AiT:") != 1 {
		t.Errorf("duplicate prefix: %q", out)
	}
	if !strings.HasSuffix(out, "\r\n") {
		t.Errorf("want CRLF ending: %q", out)
	}
	if strings.Contains(out, "\n\n") {
		t.Errorf("extra newlines: %q", out)
	}
}
