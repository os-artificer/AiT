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

package crlf

import (
	"bytes"
	"strings"
	"testing"
)

func TestCRLFWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
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
	w := NewWriter(&buf)
	_, _ = w.Write([]byte("line\n"))
	if !strings.HasSuffix(buf.String(), "\r\n") {
		t.Fatalf("suffix: %q", buf.String())
	}
}

// TestCRLFWriter_IdempotentOnCRLF verifies that input which already contains
// CRLF sequences passes through unchanged, so a CRLF writer can be composed
// over a pre-CRLF stream without producing \r\r\n.
func TestCRLFWriter_IdempotentOnCRLF(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	n, err := w.Write([]byte("a\r\nb\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 6 {
		t.Errorf("Write returned %d, want 6", n)
	}
	if got := buf.String(); got != "a\r\nb\r\n" {
		t.Errorf("got %q want %q", got, "a\r\nb\r\n")
	}
	if strings.Contains(buf.String(), "\r\r") {
		t.Errorf("unexpected \\r\\r in output: %q", buf.String())
	}
}

// TestCRLFWriter_KeepsLoneCR verifies that a lone \r (e.g. cursor return
// without a newline) is not turned into \r\n or otherwise mutated.
func TestCRLFWriter_KeepsLoneCR(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	n, err := w.Write([]byte("a\rb"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("Write returned %d, want 3", n)
	}
	if got := buf.String(); got != "a\rb" {
		t.Errorf("got %q want %q", got, "a\rb")
	}
}

// TestCRLFWriter_IdempotentAcrossWrites verifies that the \r-preceding-\n
// detection works when CR and LF are split across two Write calls.
func TestCRLFWriter_IdempotentAcrossWrites(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if _, err := w.Write([]byte("a\r")); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("\nb\n")); err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "a\r\nb\r\n" {
		t.Errorf("got %q want %q", got, "a\r\nb\r\n")
	}
}
