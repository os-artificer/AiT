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
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ai"
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
	_, _ = c.scroll.Write([]byte("u@h [~] ➜ \r\n"))
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
	// Native prompt mode (UseRemoteSystemPrompt): meta must not poke the remote
	// shell with a prompt redraw — the native prompt is re-emitted from scrollback.
	if stdinHasPromptRedraw(stdin.String()) {
		t.Errorf("native prompt mode should not inject prompt redraw, stdin = %q", stdin.String())
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

// TestHandleMetaLineUnknownRefreshesPrompt: native prompt mode re-emits from scrollback
// without injecting a remote redraw when a prompt line is already buffered.
func TestHandleMetaLineUnknownRefreshesPrompt(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	_, _ = c.scroll.Write([]byte("u@h:~$ \r\n"))
	var stdin bytes.Buffer
	err := c.handleMetaLine(context.Background(), &stdin, "!ait nosuch", InteractiveOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if stdin.Len() != 0 {
		t.Errorf("native prompt mode should not inject remote redraw, stdin = %q", stdin.String())
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
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	_, _ = c.scroll.Write([]byte("u@h [~] ➜ \r\n"))
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
	// Native prompt mode: the buffered native prompt is re-emitted from scrollback.
	if !strings.Contains(got, "u@h [~] ➜") {
		t.Fatalf("stdout = %q, want native prompt re-emit", got)
	}
	if stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("native prompt mode should not inject prompt redraw, stdin = %q", stdin.String())
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
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	_, _ = c.scroll.Write([]byte("u@h [~] ➜ \r\n"))
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
	// Native prompt mode: the buffered native prompt is re-emitted from scrollback.
	if !strings.Contains(got, "u@h [~] ➜") {
		t.Fatalf("stdout = %q, want native prompt re-emit", got)
	}
	if stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("native prompt mode should not inject prompt redraw, stdin = %q", stdin.String())
	}
}

// TestFinishMetaInteraction_ResyncsWhenPromptShown: native prompt mode re-emits the
// native prompt from scrollback after meta, without injecting a remote redraw.
func TestFinishMetaInteraction_ResyncsWhenPromptShown(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	c.pauseRemoteOut.Store(true)

	nativeLine := "root@remote [~] ➜ "
	_, _ = c.scroll.Write([]byte(nativeLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	if stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("native prompt mode should not inject prompt redraw, stdin = %q", stdin.String())
	}
	if !strings.Contains(buf.String(), "root@remote [~] ➜") {
		t.Fatalf("stdout should re-emit native prompt, got %q", buf.String())
	}
}

// TestFinishMetaInteraction_SyncsRemotePS1WhenSetupSucceeded: native prompt mode re-emits
// the buffered native prompt and never injects PS1 redraw into the remote shell.
func TestFinishMetaInteraction_SyncsRemotePS1WhenSetupSucceeded(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	c.pauseRemoteOut.Store(true)

	nativeLine := "root@remote [~] ➜ "
	_, _ = c.scroll.Write([]byte(nativeLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	if stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("native prompt mode should not inject prompt redraw, stdin = %q", stdin.String())
	}
	if !strings.Contains(buf.String(), "root@remote [~] ➜") {
		t.Fatalf("stdout should re-emit native prompt, got %q", buf.String())
	}
}

// TestFinishMetaInteraction_RepaintsPromptWhenDuplicateFilterActive: native prompt mode
// re-emits the native prompt from scrollback after meta output, without a trailing newline.
func TestFinishMetaInteraction_RepaintsPromptWhenDuplicateFilterActive(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)
	nativeLine := "root@remote [~] ➜ "
	_, _ = c.scroll.Write([]byte(nativeLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	got := buf.String()
	if !strings.Contains(got, "root@remote [~] ➜") {
		t.Fatalf("stdout should re-emit native prompt, got %q", got)
	}
	if strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\r\n") {
		t.Fatalf("re-emitted prompt must not end with newline, got %q", got)
	}
}

// TestFinishMetaInteraction_PaintsAITWhenManaged: native prompt mode re-emits the native
// oh-my-zsh prompt after meta and never paints an AiT-managed prompt.
func TestFinishMetaInteraction_PaintsAITWhenManaged(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	c.pauseRemoteOut.Store(true)

	nativeLine := "root@remote [~] ➜"
	_, _ = c.scroll.Write([]byte(nativeLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	got := buf.String()
	if !strings.Contains(got, nativeLine) {
		t.Fatalf("expected native prompt re-emit, got %q", got)
	}
	if strings.Contains(got, "AiT ") {
		t.Fatalf("should not paint AiT prompt in native mode, got %q", got)
	}
}

// TestFinishMetaInteraction_NativePromptDialHostMismatch: when the remote PS1 hostname
// differs from the SSH dial target (e.g. root@devhost:~# via testPrimaryHostFQDN), meta must still
// re-emit the buffered native prompt without waiting for Enter.
func TestFinishMetaInteraction_NativePromptDialHostMismatch(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	nativeLine := "root@devhost:~# "
	_, _ = c.scroll.Write([]byte(nativeLine))

	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&stdin)

	got := buf.String()
	if !strings.Contains(got, "root@devhost:~#") {
		t.Fatalf("stdout should re-emit native prompt despite dial host mismatch, got %q", got)
	}
	if stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("native prompt mode should not inject prompt redraw, stdin = %q", stdin.String())
	}
}

// TestFinishNativePromptAfterMeta_SyncFallback: when scrollback has no prompt, meta falls
// back to a silent remote redraw and re-emits the prompt captured while stdout is paused.
func TestFinishNativePromptAfterMeta_SyncFallback(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	var stdin bytes.Buffer
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	go func() {
		time.Sleep(30 * time.Millisecond)
		_, _ = c.scroll.Write([]byte("root@tencentos:~# "))
		c.touchRemoteOut()
	}()

	if !c.finishNativePromptAfterMeta(&stdin) {
		t.Fatal("expected prompt after sync fallback")
	}
	got := buf.String()
	if !strings.Contains(got, "root@tencentos:~#") {
		t.Fatalf("stdout = %q, want remote prompt after sync", got)
	}
}

// TestFinishMetaInteraction_ShortBashPrompt: relaxed heuristics match short bash PS1 lines.
func TestFinishMetaInteraction_ShortBashPrompt(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	_, _ = c.scroll.Write([]byte("u@h:# "))

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	c.finishMetaInteraction(&bytes.Buffer{})

	if !strings.Contains(buf.String(), "u@h:#") {
		t.Fatalf("stdout = %q, want short bash prompt re-emit", buf.String())
	}
}

// RA3: recoverable AI input errors print hint and return nil so SSH forwardStdin continues.
func TestHandleAIMetaLine_ErrHintNoDisconnect(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	var stdin bytes.Buffer
	err := c.handleAIMetaLine(context.Background(), &stdin, "ask", InteractiveOpts{
		AI: &ai.Client{},
		OnRemoteAI: func(_ RemoteAIParams) error {
			return ai.ChatMessageRequiredErr()
		},
	})
	if err != nil {
		t.Fatalf("expected nil (SSH continues), got %v", err)
	}
}

// RA3: non-recoverable AI errors still propagate to caller.
func TestHandleAIMetaLine_APIErrorPropagates(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	var stdin bytes.Buffer
	want := errors.New("network down")
	err := c.handleAIMetaLine(context.Background(), &stdin, "hello", InteractiveOpts{
		AI: &ai.Client{},
		OnRemoteAI: func(_ RemoteAIParams) error {
			return want
		},
	})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
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
