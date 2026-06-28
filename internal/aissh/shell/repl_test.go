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
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestExecuteLineExit(t *testing.T) {
	exit, err := ExecuteLine(context.Background(), nil, "exit", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !exit {
		t.Error("want exitREPL")
	}
}

func TestExecuteLineQuit(t *testing.T) {
	exit, err := ExecuteLine(context.Background(), nil, "quit", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !exit {
		t.Error("want exitREPL for quit")
	}
}

func TestReadTTYLineEndsOnCRAlone(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("help\r")))
	line, err := ReadTTYLine(br, io.Discard, "AiT> ", nil)
	if err != nil {
		t.Fatal(err)
	}
	if line != "help" {
		t.Errorf("line = %q, want help", line)
	}
	if br.Buffered() > 0 {
		t.Error("unexpected buffered bytes after lone CR submit")
	}
}

func TestExecuteLineHelp(t *testing.T) {
	var buf bytes.Buffer
	exit, err := ExecuteLine(context.Background(), nil, "help", nil)
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	PrintHelp(&buf, false)
	out := buf.String()
	for _, want := range []string{"Usage:", "Examples:", "help for help", "[flags] destination", "-p port"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestShellPrompt(t *testing.T) {
	if shellPrompt() != client.BasePrompt {
		t.Errorf("shellPrompt = %q", shellPrompt())
	}
}

func TestReadLine_NoHistoryOnPick(t *testing.T) {
	hist := terminal.NewLineHistory()
	hist.Add("list")

	br := bufio.NewReader(bytes.NewReader([]byte("2\r\x1b[A\r")))
	_, err := ReadTTYLineNoHistory(br, io.Discard, "Select: ")
	if err != nil {
		t.Fatal(err)
	}

	line, err := ReadTTYLine(br, io.Discard, "AiT> ", hist)
	if err != nil {
		t.Fatal(err)
	}
	if line != "list" {
		t.Fatalf("line = %q, want list (picker input must not pollute history)", line)
	}
}
