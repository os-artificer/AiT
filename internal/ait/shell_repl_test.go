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

package ait

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/sshclient"
)

func TestExecuteShellLineExit(t *testing.T) {
	exit, err := ExecuteShellLine(context.Background(), nil, "exit")
	if err != nil {
		t.Fatal(err)
	}
	if !exit {
		t.Error("want exitREPL")
	}
}

func TestExecuteShellLineQuit(t *testing.T) {
	exit, err := ExecuteShellLine(context.Background(), nil, "quit")
	if err != nil {
		t.Fatal(err)
	}
	if !exit {
		t.Error("want exitREPL for quit")
	}
}

// TestReadTTYLineEndsOnCRAlone ensures Enter that sends only \r does not block for a second key.
func TestReadTTYLineEndsOnCRAlone(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("help\r")))
	line, err := readTTYLine(br)
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

func TestReadTTYLineEndsOnCRLF(t *testing.T) {
	br := bufio.NewReader(bytes.NewReader([]byte("list\r\n")))
	line, err := readTTYLine(br)
	if err != nil {
		t.Fatal(err)
	}
	if line != "list" {
		t.Errorf("line = %q, want list", line)
	}
}

func TestExecuteShellLineHelp(t *testing.T) {
	var buf bytes.Buffer
	exit, err := ExecuteShellLine(context.Background(), nil, "help")
	if err != nil || exit {
		t.Fatalf("exit=%v err=%v", exit, err)
	}
	PrintShellHelp(&buf)
	if !strings.Contains(buf.String(), "login --host") {
		t.Error("help should mention login")
	}
}

func TestShellPrompt(t *testing.T) {
	if shellPrompt() != sshclient.BasePrompt {
		t.Errorf("shellPrompt = %q", shellPrompt())
	}
}
