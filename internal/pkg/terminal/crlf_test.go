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

func TestCRLFWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewCRLFWriter(&buf)
	n, err := w.Write([]byte("a\nb\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Errorf("Write returned %d, want 4", n)
	}
	if got := buf.String(); got != "a\r\nb\r\n" {
		t.Errorf("got %q want %q", got, "a\r\nb\r\n")
	}
}

func TestCRLFWriterPreservesReturnCount(t *testing.T) {
	var buf bytes.Buffer
	w := NewCRLFWriter(&buf)
	_, _ = w.Write([]byte("line\n"))
	if !strings.HasSuffix(buf.String(), "\r\n") {
		t.Fatalf("suffix: %q", buf.String())
	}
}
