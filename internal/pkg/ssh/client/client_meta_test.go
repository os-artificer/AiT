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

package client

import (
	"bytes"
	"context"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestRefreshRemotePrompt(t *testing.T) {
	c := &Client{}
	var stdin bytes.Buffer
	c.refreshRemotePrompt(&stdin)
	if stdin.String() != "\r" {
		t.Fatalf("stdin = %q, want \\r", stdin.String())
	}

	stdin.Reset()
	c.refreshRemotePrompt(nil)
	if stdin.Len() != 0 {
		t.Fatalf("nil stdin should skip refresh, got %q", stdin.String())
	}

	stdin.Reset()
	c.localShell.mu.Lock()
	c.localShell.pty = os.Stdin
	c.localShell.mu.Unlock()
	c.refreshRemotePrompt(&stdin)
	c.localShell.mu.Lock()
	c.localShell.pty = nil
	c.localShell.mu.Unlock()
	if stdin.Len() != 0 {
		t.Fatalf("local mode should skip refresh, got %q", stdin.String())
	}
}

func TestHandleMetaLineCLI(t *testing.T) {
	var got []string
	var stdin bytes.Buffer
	c := &Client{}
	err := c.handleMetaLine(context.Background(), &stdin, "!ait list", InteractiveOpts{
		OnMetaCLI: func(args []string) error {
			got = args
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"list"}) {
		t.Errorf("OnMetaCLI args = %v, want [list]", got)
	}
	if stdin.String() != "\r" {
		t.Errorf("stdin = %q, want remote prompt refresh \\r", stdin.String())
	}
}

func TestHandleMetaLineHelp(t *testing.T) {
	var got []string
	c := &Client{}
	for _, line := range []string{"!ait help", "!ait"} {
		got = nil
		err := c.handleMetaLine(context.Background(), &bytes.Buffer{}, line, InteractiveOpts{
			OnMetaCLI: func(args []string) error {
				got = args
				return nil
			},
		})
		if err != nil {
			t.Fatalf("%q: %v", line, err)
		}
		if got != nil {
			t.Errorf("%q: want nil args, got %v", line, got)
		}
	}
}

func TestHandleMetaLineUnknownRefreshesPrompt(t *testing.T) {
	c := &Client{}
	var stdin bytes.Buffer
	err := c.handleMetaLine(context.Background(), &stdin, "not-a-meta-line", InteractiveOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if stdin.String() != "\r" {
		t.Errorf("unknown meta should refresh remote prompt, stdin = %q", stdin.String())
	}
}

func TestRunMetaCLIFinishesStdout(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	var stdin bytes.Buffer
	c := &Client{}
	err = c.runMetaCLI(&stdin, InteractiveOpts{
		OnMetaCLI: func(args []string) error {
			if len(args) != 1 || args[0] != "list" {
				t.Fatalf("args=%v", args)
			}
			_, err := io.WriteString(os.Stdout, "no trailing newline")
			return err
		},
	}, []string{"list"})
	w.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	got := buf.String()
	if !strings.HasSuffix(got, "no trailing newline"+termEOL) {
		t.Fatalf("stdout = %q, want suffix %q", got, "no trailing newline"+termEOL)
	}
	if stdin.String() != "\r" {
		t.Fatalf("stdin = %q, want remote prompt refresh \\r", stdin.String())
	}
}

func TestHandleMetaLineListFinishesStdout(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStdout := os.Stdout
	os.Stdout = w

	var stdin bytes.Buffer
	c := &Client{}
	err = c.handleMetaLine(context.Background(), &stdin, "!ait list", InteractiveOpts{
		OnMetaCLI: func(args []string) error {
			_, err := io.WriteString(os.Stdout, "row1")
			return err
		},
	})
	w.Close()
	os.Stdout = oldStdout
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()

	got := buf.String()
	if !strings.HasSuffix(got, "row1"+termEOL) {
		t.Fatalf("stdout = %q, want suffix %q", got, "row1"+termEOL)
	}
	if stdin.String() != "\r" {
		t.Fatalf("stdin = %q, want remote prompt refresh \\r", stdin.String())
	}
}

func TestHandleMetaLineTransferSkipsOnMetaCLI(t *testing.T) {
	var called bool
	c := &Client{}
	err := c.handleMetaLine(context.Background(), &bytes.Buffer{}, "!ait get /tmp/r ./l", InteractiveOpts{
		OnMetaCLI: func(args []string) error {
			called = true
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("OnMetaCLI must not run for !ait get")
	}
}

func TestMetascannerHelpLine(t *testing.T) {
	fwd, _, meta := drive("!ait help\n", true)
	if len(fwd) != 0 {
		t.Errorf("forwarded %q", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait help" {
		t.Errorf("meta=%v want [!ait help]", meta)
	}
}
