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

package ask

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteHelp(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteHelp(&buf); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Usage:") {
		t.Fatalf("missing Usage: %q", out)
	}
	if !strings.Contains(out, "Examples:") {
		t.Fatalf("missing Examples: %q", out)
	}
	if !strings.Contains(out, "help for ask") {
		t.Fatalf("missing help footer: %q", out)
	}
	if strings.Contains(out, "execute shell commands") {
		t.Fatalf("ask help must not mention agent execute intro: %q", out)
	}
}
