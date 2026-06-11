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

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestRefreshRemotePrompt(t *testing.T) {
	c := &Client{}
	var stdin bytes.Buffer
	c.refreshRemotePrompt(&stdin)
	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin = %q, want prompt redraw inject", stdin.String())
	}

	stdin.Reset()
	c.refreshRemotePrompt(nil)
	if stdin.Len() != 0 {
		t.Fatalf("nil stdin should skip refresh, got %q", stdin.String())
	}

	stdin.Reset()
	c.localShell.SetPTYForTest(os.Stdin)
	c.refreshRemotePrompt(&stdin)
	c.localShell.SetPTYForTest(nil)
	if stdin.Len() != 0 {
		t.Fatalf("local mode should skip refresh, got %q", stdin.String())
	}
}

func TestHandleMetaLineCLI(t *testing.T) {
	var got []string
	var stdin bytes.Buffer
	c := testClientWithPrompt("u", "h")
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
	if !stdinHasPromptRedraw(stdin.String()) {
		t.Errorf("stdin = %q, want prompt redraw inject", stdin.String())
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
	if !stdinHasPromptRedraw(stdin.String()) {
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
	c := testClientWithPrompt("u", "h")
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
	if !strings.Contains(got, "no trailing newline") {
		t.Fatalf("stdout = %q, want meta output", got)
	}
	if !strings.Contains(got, "AiT u@h:") {
		t.Fatalf("stdout = %q, want repainted prompt", got)
	}
	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin = %q, want prompt redraw inject", stdin.String())
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
	c := testClientWithPrompt("u", "h")
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
	if !strings.Contains(got, "row1") {
		t.Fatalf("stdout = %q, want meta output", got)
	}
	if !strings.Contains(got, "AiT u@h:") {
		t.Fatalf("stdout = %q, want repainted prompt", got)
	}
	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin = %q, want prompt redraw inject", stdin.String())
	}
}

// TestFinishMetaInteraction_ResyncsWhenPromptShown (M5) injects remote redraw whenever the
// AiT prompt was already shown, even when the setup sentinel was never recorded.
func TestFinishMetaInteraction_ResyncsWhenPromptShown(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	c.pauseRemoteOut.Store(true)

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin = %q, want prompt redraw inject when prompt shown", stdin.String())
	}
	if !strings.Contains(buf.String(), "AiT root@ylgeeker.com:") {
		t.Fatalf("stdout should repaint AiT prompt, got %q", buf.String())
	}
}

// TestFinishMetaInteraction_SyncsRemotePS1WhenSetupSucceeded (M6) wakes remote PS1 even when
// scroll already holds a stale AiT prompt line (meta must not skip precmd).
func TestFinishMetaInteraction_SyncsRemotePS1WhenSetupSucceeded(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	c.pauseRemoteOut.Store(true)

	promptLine := "AiT root@ylgeeker.com:~> "
	_, _ = c.scroll.Write([]byte(promptLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin = %q, want prompt redraw inject when setup succeeded", stdin.String())
	}
	if !strings.Contains(buf.String(), "AiT root@ylgeeker.com:") {
		t.Fatalf("stdout should repaint AiT prompt, got %q", buf.String())
	}
}

// TestFinishMetaInteraction_RepaintsPromptWhenDuplicateFilterActive (M4) repaints the
// prompt locally after meta output because remote redraw is dropped as duplicate.
func TestFinishMetaInteraction_RepaintsPromptWhenDuplicateFilterActive(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")
	c.remotePromptShown.Store(true)
	promptLine := "AiT root@ylgeeker.com:~> "
	_, _ = c.scroll.Write([]byte(promptLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	got := buf.String()
	if !strings.Contains(got, "AiT root@ylgeeker.com:") {
		t.Fatalf("stdout should repaint prompt, got %q", got)
	}
	if strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\r\n") {
		t.Fatalf("repainted prompt must not end with newline, got %q", got)
	}

	before := buf.Len()
	c.handleRemoteOutput([]byte(promptLine + "\r\n"))
	if buf.Len() != before {
		t.Fatalf("duplicate remote redraw should still be filtered, got %q", buf.String())
	}
}

// TestFinishMetaInteraction_EmitsNativePromptFromScrollback (M7) repaints the user's native
// shell prompt after meta when scrollback holds native output but no AiT line (oh-my-zsh hosts).
func TestFinishMetaInteraction_EmitsNativePromptFromScrollback(t *testing.T) {
	c := testClientWithPrompt("root", "ylgeeker.com")
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	c.pauseRemoteOut.Store(true)

	nativeLine := "root@ylgeeker [~] ➜"
	_, _ = c.scroll.Write([]byte(nativeLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	got := buf.String()
	if strings.Contains(got, "AiT root@ylgeeker.com:") {
		t.Fatalf("stdout should not client-paint AiT when native prompt is buffered, got %q", got)
	}
	if !strings.Contains(got, nativeLine) {
		t.Fatalf("stdout should emit native prompt from scrollback, got %q", got)
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

// Coverage moved to internal/pkg/ssh/client/meta/scanner_test.go (TestMetaScanner_HelpLine).
