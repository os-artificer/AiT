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
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestStripSetupLeakLines_WrappedFragments (N19) drops wrapped PS1/probe echo lines from recorded data.
func TestStripSetupLeakLines_WrappedFragments(t *testing.T) {
	user, host := testSSHUser, testPrimaryHostFQDN
	data := strings.Join([]string{
		"Welcome to TencentOS Server 3 x86_64",
		`"${PWD:-/}" 2>/dev/null); printf '%s' "${d:-/}")$(b=$( 2>/dev/null) && [ -n "$b" ] && printf ":%s" "$b")> \[\]'; [ -n`,
		"hello world",
		"AITPWDEND__.';  printf '';  stty echo 2>/dev/null",
		"AiT root@remote.example.com:/tmp> ",
	}, "\r\n")

	got := StripSetupLeakLines(data, user, host)

	for _, leak := range []string{`"${PWD:-/}"`, "AITPWDEND", "stty echo"} {
		if strings.Contains(got, leak) {
			t.Errorf("leak %q should be stripped, got %q", leak, got)
		}
	}
	for _, keep := range []string{"Welcome to TencentOS", "hello world", "AiT root@remote.example.com:/tmp>"} {
		if !strings.Contains(got, keep) {
			t.Errorf("normal line %q should be kept, got %q", keep, got)
		}
	}

	if StripSetupLeakLines("", user, host) != "" {
		t.Error("empty data should stay empty")
	}
}

// TestHandleSuspendMeta_ScrollLeakFiltered (N20) filters leaked setup echo from scroll before save.
func TestHandleSuspendMeta_ScrollLeakFiltered(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	_, _ = c.scroll.Write([]byte("ls output\r\nstty -echo 2>/dev/null; set +o history 2>/dev/null\r\nAiT u@h:/tmp> "))

	origPwd := probeShellPwdHook
	defer func() { probeShellPwdHook = origPwd }()
	probeShellPwdHook = func(_ *Client, _ io.Writer) (string, error) {
		return "/tmp", nil
	}

	var saved string
	opts := InteractiveOpts{
		OnSuspend: func(scroll *terminal.Buffer, _ string, _, _ int) error {
			saved = scroll.String()
			return nil
		},
	}

	stdin := &bytes.Buffer{}
	if err := c.handleSuspendMeta(stdin, opts); err != nil {
		t.Fatalf("handleSuspendMeta: %v", err)
	}

	if strings.Contains(saved, "stty -echo") || strings.Contains(saved, "set +o history") {
		t.Fatalf("saved scrollback should not contain setup leak, got %q", saved)
	}
	if !strings.Contains(saved, "ls output") {
		t.Fatalf("normal output should be kept, got %q", saved)
	}
}

// TestFilterRemoteOutputChunk_EchoMarkerKeepsSuppress (N21) keeps suppression on for
// echoed setup text containing the prompt marker; only a real prompt line ends it.
func TestFilterRemoteOutputChunk_EchoMarkerKeepsSuppress(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.remoteSetupSuppress.Store(true)
	c.setupInjectDone.Store(true)

	// Non-leak text containing the marker substring mid-line is not a prompt.
	_ = c.filterRemoteOutputChunk([]byte("echo of AiT u@h: marker text\n"), "u", "h")
	if !c.remoteSetupSuppress.Load() {
		t.Fatal("marker substring inside echo must not end suppression")
	}

	// Leaked PS1 export (contains marker) is filtered and must not end suppression.
	leak := "export PS1='\\[\\]AiT u@h:${PWD/#$HOME/\\~}> \\[\\]'\n"
	_ = c.filterRemoteOutputChunk([]byte(leak), "u", "h")
	if !c.remoteSetupSuppress.Load() {
		t.Fatal("echoed PS1 export must not end suppression")
	}

	// A real prompt line (no trailing newline) ends suppression.
	got := c.filterRemoteOutputChunk([]byte("AiT u@h:/tmp> "), "u", "h")
	if c.remoteSetupSuppress.Load() {
		t.Fatal("real prompt line should end suppression")
	}
	if !strings.Contains(string(got), "AiT u@h:/tmp>") {
		t.Fatalf("prompt should pass through, got %q", got)
	}
}

// TestWaitForRemoteAITPrompt_RealPromptLineOnly (N21) ignores marker text inside setup echo.
func TestWaitForRemoteAITPrompt_RealPromptLineOnly(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	_, _ = c.scroll.Write([]byte("export PS1='AiT u@h:dir> ' echoed back\n"))

	start := time.Now()
	c.waitForRemoteAITPrompt("u", "h", 80*time.Millisecond)
	if time.Since(start) < 80*time.Millisecond {
		t.Fatal("echoed marker should not satisfy the prompt wait")
	}

	_, _ = c.scroll.Write([]byte("AiT u@h:/tmp> "))
	start = time.Now()
	c.waitForRemoteAITPrompt("u", "h", 2*time.Second)
	if time.Since(start) > time.Second {
		t.Fatal("real prompt line should satisfy the wait quickly")
	}
}

// TestSetupOutputFilter_FlushesPromptPending (N22) releases a complete prompt without trailing newline.
func TestSetupOutputFilter_FlushesPromptPending(t *testing.T) {
	var f setupOutputFilter

	got := f.filter([]byte("AiT u@h:/tmp> "), "u", "h")
	if !strings.Contains(string(got), "AiT u@h:/tmp>") {
		t.Fatalf("prompt pending should flush immediately, got %q", got)
	}
	if len(f.pending) != 0 {
		t.Fatalf("pending should be empty after prompt flush, got %q", f.pending)
	}

	// Non-prompt incomplete lines stay pending.
	got = f.filter([]byte("partial line without newline"), "u", "h")
	if len(got) != 0 {
		t.Fatalf("incomplete non-prompt line should stay pending, got %q", got)
	}
}

// TestSetupGrace_StillFiltersAfterPromptShown (P1j) keeps leak filtering after remotePromptShown.
func TestSetupGrace_StillFiltersAfterPromptShown(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	leak := []byte("stty -echo 2>/dev/null\r\n")
	if got := c.applySetupGrace(leak, "u", "h"); len(got) != 0 {
		t.Fatalf("leak should be filtered during grace, got %q", got)
	}

	c.remotePromptShown.Store(true)
	if got := c.applySetupGrace([]byte("#\r\n"), "u", "h"); len(got) != 0 {
		t.Fatalf("noise should still be filtered after prompt shown, got %q", got)
	}
	out := []byte("normal output\r\n")
	if got := c.applySetupGrace(out, "u", "h"); !bytes.Equal(got, out) {
		t.Fatalf("normal output should pass through, got %q", got)
	}
	if c.setupLeakGraceUntil.Load() == 0 {
		t.Fatal("leak grace window should remain active until cap")
	}
}

// TestSetupGrace_CapEnds (N23) ends the grace filter after the cap even without a prompt.
func TestSetupGrace_CapEnds(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupLeakGraceUntil.Store(time.Now().Add(-time.Second).UnixNano())

	out := []byte("late output\r\n")
	if got := c.applySetupGrace(out, "u", "h"); !bytes.Equal(got, out) {
		t.Fatalf("expired grace should pass output through, got %q", got)
	}
	if c.setupLeakGraceUntil.Load() != 0 {
		t.Fatal("grace window should be cleared after cap")
	}
}

// TestGraceStripper_FlushesNonLeakPending (N24/N28) flushes complete native prompts immediately
// while holding partial fragments until the line completes.
func TestGraceStripper_FlushesNonLeakPending(t *testing.T) {
	var s setupStdoutStripper
	user, host := "u", "h"

	native := []byte("root@" + testPrimaryHostShort + " [~] \u2192")
	if got := s.filterChunk(native, user, host, true); !bytes.Equal(got, native) {
		t.Fatalf("native prompt should flush immediately, got %q", got)
	}

	// Partial leak is held until the line completes, then dropped.
	if got := s.filterChunk([]byte("stty -ech"), user, host, true); len(got) != 0 {
		t.Fatalf("partial leak should be held, got %q", got)
	}
	if got := s.filterChunk([]byte("o 2>/dev/null\r\n"), user, host, true); len(got) != 0 {
		t.Fatalf("completed leak line should be dropped, got %q", got)
	}

	// Partial non-prompt text (no terminator) stays held, not flushed mid-line.
	if got := s.filterChunk([]byte("partial frag"), user, host, true); len(got) != 0 {
		t.Fatalf("partial non-prompt pending should be held, got %q", got)
	}
}

// TestGraceStripper_HoldsSingleCharPending (N25) never flushes lone characters during grace.
func TestGraceStripper_HoldsSingleCharPending(t *testing.T) {
	user, host := "u", "h"

	for _, ch := range []string{"#", "S", "a", "H"} {
		var s setupStdoutStripper
		if got := s.filterChunk([]byte(ch), user, host, true); len(got) != 0 {
			t.Errorf("single char %q should be held, got %q", ch, got)
		}
		// Completing the line turns it into a noise line, which is dropped.
		if got := s.filterChunk([]byte("\r\n"), user, host, true); len(got) != 0 {
			t.Errorf("noise line %q should be dropped, got %q", ch, got)
		}
	}
}

// TestHandleRemoteOutput_GraceFiltersScrollAndStdout (N26) keeps scroll and stdout
// symmetric during the grace window so leaked setup echo never reaches the saved scrollback.
func TestHandleRemoteOutput_GraceFiltersScrollAndStdout(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.pauseRemoteOut.Store(false)
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	c.handleRemoteOutput([]byte("stty -echo 2>/dev/null\r\n"))
	c.handleRemoteOutput([]byte("#\r\n"))
	c.handleRemoteOutput([]byte("normal output line\r\n"))

	scroll := c.scroll.String()
	stdout := buf.String()
	for _, leak := range []string{"stty -echo", "#"} {
		if strings.Contains(scroll, leak) {
			t.Errorf("scroll should not contain %q, got %q", leak, scroll)
		}
		if strings.Contains(stdout, leak) {
			t.Errorf("stdout should not contain %q, got %q", leak, stdout)
		}
	}
	if !strings.Contains(scroll, "normal output line") {
		t.Errorf("scroll should keep normal output, got %q", scroll)
	}
	if !strings.Contains(stdout, "normal output line") {
		t.Errorf("stdout should keep normal output, got %q", stdout)
	}
}

// TestApplySetupGrace_SuppressesLoneNewlineAfterPrompt (P1o) drops stray PTY newlines after client paint.
func TestApplySetupGrace_SuppressesLoneNewlineAfterPrompt(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.remotePromptShown.Store(true)
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	if got := c.applySetupGrace([]byte("\r\n"), "u", "h"); got != nil {
		t.Fatalf("lone newline should be suppressed, got %q", got)
	}
	if got := c.applySetupGrace([]byte("ls\r\n"), "u", "h"); got == nil {
		t.Fatal("normal output should not be suppressed")
	}
}

// TestProcessForwardStdin_FirstKeyEndsKeystrokeGrace (N29) clears keystroke grace
// but leaves leak filtering active when the user's first keystroke is forwarded.
func TestProcessForwardStdin_FirstKeyEndsKeystrokeGrace(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())
	c.setupKeystrokeGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	var stdin bytes.Buffer
	res := c.processForwardStdinByte(
		context.Background(),
		&stdin,
		stdinByte{B: 'l'},
		newMetaScanner(""),
		InteractiveOpts{},
	)
	if res.action == fwdStdinReturn && res.err != nil {
		t.Fatalf("unexpected return: %v", res.err)
	}
	if stdin.Len() != 1 || stdin.Bytes()[0] != 'l' {
		t.Fatalf("expected keystroke forwarded, stdin=%q", stdin.Bytes())
	}
	if c.setupKeystrokeGraceUntil.Load() != 0 {
		t.Fatal("first forwarded keystroke should end keystroke grace")
	}
	if c.setupLeakGraceUntil.Load() == 0 {
		t.Fatal("leak filter should remain active after first keystroke")
	}
}

// TestHandleRemoteOutput_EchoVisibleAfterKeystroke (N30) passes single-character
// keystroke echo straight to stdout and scroll once the grace window has ended.
func TestHandleRemoteOutput_EchoVisibleAfterKeystroke(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())
	c.setupKeystrokeGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	var stdin bytes.Buffer
	_ = c.processForwardStdinByte(
		context.Background(),
		&stdin,
		stdinByte{B: 'l'},
		newMetaScanner(""),
		InteractiveOpts{},
	)

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.pauseRemoteOut.Store(false)

	c.handleRemoteOutput([]byte("l"))
	c.handleRemoteOutput([]byte("s"))

	if got := buf.String(); got != "ls" {
		t.Fatalf("keystroke echo should be visible, stdout=%q", got)
	}
	if !strings.Contains(c.scroll.String(), "ls") {
		t.Fatalf("keystroke echo should reach scroll, got %q", c.scroll.String())
	}
}

// TestSetupGraceStripper_EndGraceDiscardsPending (P2) drops incomplete pending on grace end.
func TestSetupGraceStripper_EndGraceDiscardsPending(t *testing.T) {
	var s setupStdoutStripper
	user, host := "u", "h"

	_ = s.filterChunk([]byte("partial frag"), user, host, true)
	if got := s.endGrace(user, host); len(got) != 0 {
		t.Fatalf("endGrace should discard pending, got %q", got)
	}
	if len(s.pending) != 0 {
		t.Fatal("endGrace should clear pending buffer")
	}
}

// TestStripSetupLeakLines_DropsIncompletePrompt (P2) removes partial AiT prompt prefix lines.
func TestStripSetupLeakLines_DropsIncompletePrompt(t *testing.T) {
	user, host := testSSHUser, testPrimaryHostFQDN
	data := strings.Join([]string{
		"Welcome to TencentOS Server 3 x86_64",
		"partial AiT root@remote.example.com: without closing",
		"hello world",
		"AiT root@remote.example.com:/tmp> ",
	}, "\r\n")

	got := StripSetupLeakLines(data, user, host)

	if strings.Contains(got, "without closing") {
		t.Fatalf("incomplete prompt fragment should be stripped, got %q", got)
	}
	if !strings.Contains(got, "AiT root@remote.example.com:/tmp>") {
		t.Fatalf("real prompt line should be kept, got %q", got)
	}
}

// TestHandleRemoteOutput_FirstKeyStillFiltersLeak (P1m) strips late setup noise after first keystroke.
func TestHandleRemoteOutput_FirstKeyStillFiltersLeak(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())
	c.setupKeystrokeGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())
	c.remotePromptShown.Store(true)

	var stdin bytes.Buffer
	_ = c.processForwardStdinByte(
		context.Background(),
		&stdin,
		stdinByte{B: '\r'},
		newMetaScanner(""),
		InteractiveOpts{},
	)

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	c.pauseRemoteOut.Store(false)

	c.handleRemoteOutput([]byte("#\r\n"))
	if strings.Contains(buf.String(), "#") {
		t.Fatalf("late noise should be filtered after first key, stdout=%q", buf.String())
	}
}

// TestApplySetupGrace_FirstKeyPreservesClearTokens keeps Ctrl+L clear output visible after first keystroke.
func TestApplySetupGrace_FirstKeyPreservesClearTokens(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())
	c.setupKeystrokeGraceUntil.Store(0)

	clearChunk := []byte("\x1b[2J\x1b[HAiT u@h:/tmp> ")
	if got := c.applySetupGrace(clearChunk, "u", "h"); !bytes.Equal(got, clearChunk) {
		t.Fatalf("clear-token chunk should pass through, got %q", got)
	}
}

// TestSanitizeReplayScrollback_StripsPromptLines (P2j) removes AiT and native prompts from seed data.
func TestSanitizeReplayScrollback_StripsPromptLines(t *testing.T) {
	user, host := testSSHUser, testPrimaryHostFQDN
	data := strings.Join([]string{
		"output line",
		"AiT root@remote.example.com:/tmp> ",
		"root@remote [~] ->",
		"more output",
	}, "\n")
	got := SanitizeReplayScrollback(data, user, host)
	if strings.Contains(got, "AiT root@remote.example.com:") {
		t.Fatalf("sanitized seed should drop AiT prompt, got %q", got)
	}
	if strings.Contains(got, "root@remote [~]") {
		t.Fatalf("sanitized seed should drop native prompt, got %q", got)
	}
	if !strings.Contains(got, "output line") || !strings.Contains(got, "more output") {
		t.Fatalf("sanitized seed should keep command output, got %q", got)
	}
}

// TestSanitizeReplayScrollback_FiltersDevNull (P2h) removes split zsh history fragments on replay.
func TestSanitizeReplayScrollback_FiltersDevNull(t *testing.T) {
	data := strings.Join([]string{
		"Welcome",
		"dev/null",
		"HISTFILE=/dev/null",
		"hello",
	}, "\n")

	got := SanitizeReplayScrollback(data, "u", "h")
	if strings.Contains(got, "dev/null") {
		t.Fatalf("sanitized replay should not contain dev/null, got %q", got)
	}
	if !strings.Contains(got, "hello") {
		t.Fatalf("sanitized replay should keep hello, got %q", got)
	}
}

// TestStripSetupLeakLines_DropsSingleCharNoise (N27) removes lone-character noise
// lines (ZLE redraw fragments) from recorded scrollback data.
func TestStripSetupLeakLines_DropsSingleCharNoise(t *testing.T) {
	data := strings.Join([]string{
		"Welcome to TencentOS Server 3 x86_64",
		"#", "S", "a", "H",
		"hello world",
	}, "\r\n")

	got := StripSetupLeakLines(data, "u", "h")

	want := "Welcome to TencentOS Server 3 x86_64\r\nhello world"
	if got != want {
		t.Fatalf("StripSetupLeakLines = %q, want %q", got, want)
	}
}

// TestProcessForwardStdin_MetaEchoEndsKeystrokeGrace (MD1) clears keystroke grace on meta local echo.
func TestProcessForwardStdin_MetaEchoEndsKeystrokeGrace(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupKeystrokeGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	var stdin bytes.Buffer
	var stdout bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&stdout)

	scan := newMetaScanner("AiT u@h:/tmp> ")
	for _, b := range []byte("!ait ") {
		res := c.processForwardStdinByte(
			context.Background(),
			&stdin,
			stdinByte{B: b},
			scan,
			InteractiveOpts{},
		)
		if res.action == fwdStdinReturn && res.err != nil {
			t.Fatalf("unexpected return on %q: %v", b, res.err)
		}
	}
	if stdout.Len() == 0 {
		t.Fatal("meta echo should reach interactive stdout")
	}
	if c.setupKeystrokeGraceUntil.Load() != 0 {
		t.Fatal("meta echo should end keystroke grace")
	}
}

// TestProcessForwardStdin_LocalPassthroughEndsKeystrokeGrace (MD1b) clears grace on local PTY write.
func TestProcessForwardStdin_LocalPassthroughEndsKeystrokeGrace(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.setupKeystrokeGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer pr.Close()
	c.localShell.SetPTYForTest(pw)
	defer c.localShell.SetPTYForTest(nil)

	var stdin bytes.Buffer
	scan := newMetaScanner(prompt.Local)
	scan.SyncFrom(true, prompt.Local)

	for _, b := range []byte("!!\n") {
		res := c.processForwardStdinByte(
			context.Background(),
			&stdin,
			stdinByte{B: b},
			scan,
			InteractiveOpts{},
		)
		if res.action == fwdStdinReturn && res.err != nil {
			t.Fatalf("byte %q: %v", b, res.err)
		}
	}

	if c.setupKeystrokeGraceUntil.Load() != 0 {
		t.Fatal("local passthrough should end keystroke grace")
	}

	buf := make([]byte, 8)
	n, _ := pr.Read(buf)
	if string(buf[:n]) != "!!\n" {
		t.Fatalf("local PTY got %q want !!\\n", buf[:n])
	}
}

// TestProcessForwardStdin_MetaEchoCRLF (MD3) routes meta echo through interactiveOut with CRLF safety.
func TestProcessForwardStdin_MetaEchoCRLF(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&stdout)

	scan := newMetaScanner("AiT u@h:/tmp> ")
	for _, b := range []byte("!ait help\n") {
		_ = c.processForwardStdinByte(
			context.Background(),
			&stdin,
			stdinByte{B: b},
			scan,
			InteractiveOpts{},
		)
	}

	got := stdout.String()
	if !strings.Contains(got, "\r\n") {
		t.Fatalf("meta echo should use CRLF, stdout=%q", got)
	}
	if strings.Contains(got, "\r\r\n") {
		t.Fatalf("meta echo must not double CRLF, stdout=%q", got)
	}
}
