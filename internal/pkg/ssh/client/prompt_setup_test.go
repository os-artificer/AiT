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
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestInjectInitialCommandsKeepsPause(t *testing.T) {
	c := &Client{}
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)

	if err := c.injectInitialCommands(stdin, []string{"true"}); err != nil {
		t.Fatal(err)
	}
	if !c.pauseRemoteOut.Load() {
		t.Fatal("injectInitialCommands must not unpause; RunInteractive unpause once after setup")
	}
}

// TestInjectInitialCommands_ArmsLeakGrace (N20) enables stdout leak filtering for
// native-prompt load/switch cwd restore (cd inject with stty wrap).
func TestInjectInitialCommands_ArmsLeakGrace(t *testing.T) {
	c := &Client{}
	stdin := &bytes.Buffer{}
	if err := c.injectInitialCommands(stdin, []string{"cd '/root'"}); err != nil {
		t.Fatal(err)
	}
	if c.setupLeakGraceUntil.Load() == 0 {
		t.Fatal("expected setup leak grace after injectInitialCommands")
	}
}

func TestApplySetupGrace_FiltersLoadCdInjectLeak(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	graceEnd := time.Now().Add(setupLeakGraceMax)
	c.setupLeakGraceUntil.Store(graceEnd.UnixNano())
	c.setupKeystrokeGraceUntil.Store(graceEnd.UnixNano())
	c.setupGraceStripper.reset()

	leak := []byte(" stty -echo 2>/dev/null; cd '/root'; stty echo 2>/dev/null\r\n")
	got := c.applySetupGrace(leak, testSSHUser, testPrimaryHostFQDN)
	if strings.Contains(string(got), "stty -echo") || strings.Contains(string(got), "cd '/root'") {
		t.Fatalf("inject echo must be filtered, got %q", got)
	}
}

func TestFlushPendingNativePrompt_MultiLinePending(t *testing.T) {
	var s setupStdoutStripper
	s.pending = []byte("Welcome to TencentOS Server 3\r\nLast login: Thu Jun 18\r\n[root@remote ~]# ")
	got := s.flushPendingNativePrompt(testSSHUser, testPrimaryHostFQDN)
	if !strings.Contains(string(got), "root@remote") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(string(got), "Welcome") {
		t.Fatal("must emit prompt line only, not MOTD")
	}
	if len(s.pending) != 0 {
		t.Fatal("pending should be cleared")
	}
}

func TestHandleRemoteOutput_EmitsNativePromptFromGracePending(t *testing.T) {
	enableRemoteSystemPromptForTest(t)
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	graceEnd := time.Now().Add(setupLeakGraceMax)
	c.setupLeakGraceUntil.Store(graceEnd.UnixNano())
	c.setupKeystrokeGraceUntil.Store(graceEnd.UnixNano())
	c.setupGraceStripper.pending = []byte("root@remote [~] ->")

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.handleRemoteOutput([]byte{})

	if !c.remotePromptShown.Load() {
		t.Fatal("expected remotePromptShown after grace pending emit")
	}
	if !strings.Contains(buf.String(), "root@remote") {
		t.Fatalf("stdout = %q", buf.String())
	}
}

func TestEndGrace_FlushesNativePrompt(t *testing.T) {
	var s setupStdoutStripper
	s.pending = []byte("root@remote [~] ->")
	got := s.endGrace(testSSHUser, testPrimaryHostFQDN)
	if !strings.Contains(string(got), "root@remote") {
		t.Fatalf("endGrace should flush native prompt, got %q", got)
	}
	if len(s.pending) != 0 {
		t.Fatal("pending should be cleared after flush")
	}
}

func TestFinishNativePromptAfterAttach_FlushesGracePending(t *testing.T) {
	enableRemoteSystemPromptForTest(t)
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	graceEnd := time.Now().Add(setupLeakGraceMax)
	c.setupLeakGraceUntil.Store(graceEnd.UnixNano())
	c.setupKeystrokeGraceUntil.Store(graceEnd.UnixNano())
	c.setupGraceStripper.pending = []byte("root@remote [~] ->")
	c.lastRemoteOut.Store(time.Now().UnixNano())

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.finishNativePromptAfterAttach()

	if !c.remotePromptShown.Load() {
		t.Fatal("expected remotePromptShown")
	}
	if !strings.Contains(buf.String(), "root@remote") {
		t.Fatalf("stdout = %q", buf.String())
	}
}

// TestFinishNativePromptAfterAttach_EmitsPrompt (N20b) load/switch cwd inject must
// still paint the native prompt without requiring Enter.
func TestFinishNativePromptAfterAttach_EmitsPrompt(t *testing.T) {
	enableRemoteSystemPromptForTest(t)
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	_, _ = c.scroll.Write([]byte("Last login: Thu\nroot@remote [~] ->\r\n"))
	c.lastRemoteOut.Store(time.Now().UnixNano())

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.finishNativePromptAfterAttach()

	if !c.remotePromptShown.Load() {
		t.Fatal("expected remotePromptShown")
	}
	if !strings.Contains(buf.String(), "root@remote") {
		t.Fatalf("stdout = %q", buf.String())
	}
}

// TestNativeLoadSwitchPath_FiltersInjectAndEmitsPrompt (N20c) end-to-end native
// load/switch tail: inject cd with leak grace, then finish paints prompt.
func TestNativeLoadSwitchPath_FiltersInjectAndEmitsPrompt(t *testing.T) {
	enableRemoteSystemPromptForTest(t)
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	_, _ = c.scroll.Write([]byte("Welcome\nroot@remote [~] ->\r\n"))
	var stdin bytes.Buffer
	c.lastRemoteOut.Store(time.Now().UnixNano())

	if err := c.injectInitialCommands(&stdin, []string{"cd '/root'"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdin.String(), "print -Pn") {
		t.Fatalf("must not inject prompt redraw, stdin = %q", stdin.String())
	}

	leak := []byte(" stty -echo 2>/dev/null; cd '/root'; stty echo 2>/dev/null\r\n")
	if got := c.applySetupGrace(leak, testSSHUser, testPrimaryHostFQDN); len(got) != 0 {
		t.Fatalf("inject leak should be filtered, got %q", got)
	}

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.finishNativePromptAfterAttach()

	if !strings.Contains(buf.String(), "root@remote") {
		t.Fatalf("prompt not emitted, stdout = %q", buf.String())
	}
}
