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

package meta

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestPrintMetaHelp(t *testing.T) {
	var buf bytes.Buffer
	PrintMetaHelp(&buf)
	out := buf.String()
	for _, want := range []string{
		"Usage:",
		"Examples:",
		"help for help",
		"help",
		"ai [question]",
		"remote AI mode",
		"user@hostB",
		"session --switch",
		"local",
		"suspend",
		toolname.MetaLine("list"),
		"auto-save",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "\r\n") {
		t.Error("help should use CRLF")
	}
}

func TestExecuteHelp(t *testing.T) {
	SetExecutorWriters(func([]string, io.Writer, io.Writer) error { return nil })
	if err := Execute(nil); err != nil {
		t.Fatalf("Execute(nil): %v", err)
	}
	if err := Execute([]string{"help"}); err != nil {
		t.Fatalf("Execute(help): %v", err)
	}
}

func TestExecuteBlockedAI(t *testing.T) {
	err := Execute([]string{"ai"})
	if err == nil {
		t.Fatal("expected error for ai subcommand in session")
	}
	if !strings.Contains(err.Error(), toolname.MetaLine("ai")) {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Errorf("expected Try block: %v", err)
	}
}

func TestExecute_LoginUnknownCommand(t *testing.T) {
	SetExecutorWriters(func([]string, io.Writer, io.Writer) error {
		return errors.New(fmt.Sprintf(`unknown command "login" for "%s"`, toolname.AISsh))
	})

	err := Execute([]string{"login", "--host", "h"})
	if err == nil {
		t.Fatal("expected error for login in session")
	}
	if !helpdoc.IsUnknownCommand(err) {
		t.Fatalf("got %v", err)
	}
}

// TestMetaCLIArgs_ListUsesSessionFlag (C25) keeps !ait list mapped to session --list.
func TestMetaCLIArgs_ListUsesSessionFlag(t *testing.T) {
	got := metaCLIArgs([]string{"list"})
	want := []string{"session", "--list"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("metaCLIArgs=%v want %v", got, want)
	}
}

func TestExecuteSubcommandUsesCRLFStdout(t *testing.T) {
	SetExecutorWriters(func(args []string, stdout, _ io.Writer) error {
		want := []string{"session", "--list"}
		if len(args) != len(want) || args[0] != want[0] || args[1] != want[1] {
			t.Fatalf("args=%v want %v", args, want)
		}
		_, err := io.WriteString(stdout, "ID\tTARGET\nrow1\tu@h:22\nrow2\tu@h:22\n")
		return err
	})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	err = Execute([]string{"list"})
	w.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	got := buf.String()
	if !strings.Contains(got, "\r\n") {
		t.Fatalf("stdout missing CRLF: %q", got)
	}
	if strings.Count(got, "\r\n") < 3 {
		t.Fatalf("want at least 3 CRLF lines, got %q", got)
	}
}
