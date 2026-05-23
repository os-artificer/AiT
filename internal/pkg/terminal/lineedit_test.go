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
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReadLineRawASCII(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("help\r")))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out)
	if err != nil {
		t.Fatal(err)
	}
	if line != "help" {
		t.Errorf("line = %q, want help", line)
	}
}

func TestReadLineRawCRLF(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("list\r\n")))
	line, err := ReadLineRaw(br, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if line != "list" {
		t.Errorf("line = %q, want list", line)
	}
}

func TestReadLineRawUTF8Chinese(t *testing.T) {
	input := append([]byte("帮助"), '\r')
	br := bufio.NewReader(bytes.NewReader(input))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out)
	if err != nil {
		t.Fatal(err)
	}
	if line != "帮助" {
		t.Errorf("line = %q, want 帮助", line)
	}
	if !strings.Contains(out.String(), "帮助") {
		t.Errorf("echo missing UTF-8, got %q", out.String())
	}
}

func TestReadLineRawBackspaceUTF8(t *testing.T) {
	var input []byte
	input = append(input, []byte("a你")...)
	input = append(input, 127) // DEL
	input = append(input, '\r')
	br := bufio.NewReader(bytes.NewReader(input))
	line, err := ReadLineRaw(br, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if line != "a" {
		t.Errorf("line = %q, want a", line)
	}
}

func TestReadLineRawIgnoresEscape(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("hi\x1b[A\r")))
	line, err := ReadLineRaw(br, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if line != "hi" {
		t.Errorf("line = %q, want hi", line)
	}
}

func TestReadLineRawCtrlCEmpty(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{3}))
	var out bytes.Buffer
	line, err := ReadLineRaw(br, &out)
	if err != nil {
		t.Fatal(err)
	}
	if line != "" {
		t.Errorf("line = %q, want empty", line)
	}
	if out.String() != "^C\r\n" {
		t.Errorf("output = %q, want ^C\\r\\n", out.String())
	}
}

func TestReadLineRawCtrlDEmpty(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{4}))
	_, err := ReadLineRaw(br, io.Discard)
	if err != io.EOF {
		t.Fatalf("err = %v, want EOF", err)
	}
}

func TestReadLineRawEmptyCRIgnored(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte{'\r', 'x', '\r'}))
	line, err := ReadLineRaw(br, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if line != "x" {
		t.Errorf("line = %q, want x", line)
	}
}
