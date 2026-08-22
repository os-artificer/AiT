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

package aissh

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/ssh/shell"
)

func TestExecuteWritersSilencesStderr(t *testing.T) {
	capture := &bytes.Buffer{}
	err := ExecuteWriters([]string{"ll"}, os.Stdout, capture)
	if err == nil {
		t.Fatal("expected error")
	}
	if capture.Len() != 0 {
		t.Errorf("cobra should not write stderr, got %q", capture.String())
	}
}

func TestTopLevelLLUnknown(t *testing.T) {
	err := ExecuteWriters([]string{"ll"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	if !pkghelpdoc.IsUnknownCommand(err) {
		t.Fatalf("want unknown command, got %v", err)
	}
}

func TestExecuteWritersErrorFormattedForREPL(t *testing.T) {
	err := ExecuteWriters([]string{"session"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error")
	}
	var buf bytes.Buffer
	shell.PrintErr(&buf, err.Error())
	out := buf.String()
	if strings.Count(out, "AiT:") != 1 {
		t.Errorf("want single AiT prefix: %q", out)
	}
}
