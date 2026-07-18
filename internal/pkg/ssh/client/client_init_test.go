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
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func stdinHasPromptRedraw(s string) bool {
	return strings.Contains(s, `[ -z "$ZSH_VERSION" ] && true`) ||
		strings.Contains(s, `print -Pn "$PROMPT"`)
}

func testClientWithPrompt(user, host string) *Client {
	c := &Client{
		scroll: &terminal.Buffer{},
		opts:   Options{User: user, Host: host},
	}
	c.initActivePrompt()
	return c
}

func TestInjectInitialCommandsSingleLine(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	done := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		c.touchRemoteOut()
		close(done)
	}()

	cmds := RemotePromptInitialCommands("u", "h", false)
	if err := c.injectInitialCommands(stdin, cmds); err != nil {
		t.Fatal(err)
	}
	<-done
	_, _ = c.scroll.Write([]byte("AiT u@h:/home> "))
	c.finishRemoteSetup(stdin, "u", "h", false)

	body := stdin.String()
	if !strings.HasPrefix(body, injectSpacePrefix) {
		t.Fatalf("stdin should use Inject space prefix, got: %q", body)
	}
	if !strings.Contains(body, "stty -echo") {
		t.Fatalf("stdin missing Inject echo suppression: %q", body)
	}
	if !strings.Contains(body, cmds[0]) {
		t.Fatalf("stdin missing setup body: %q", body)
	}
	if strings.Contains(body, "\r") {
		t.Fatalf("stdin should not refresh when marker already buffered: %q", body)
	}
	if c.pauseRemoteOut.Load() {
		t.Fatal("finishRemoteSetup should unpause")
	}
	if c.remoteSetupSuppress.Load() {
		t.Fatal("finishRemoteSetup should clear setup suppress")
	}
}

// TestWaitRemoteOutputQuietAfterIgnoresPreInject: pre-inject lastRemoteOut is ignored; when
// no post-inject bytes arrive, return after quiet since injectAt (not maxWait).
func TestWaitRemoteOutputQuietAfterIgnoresPreInject(t *testing.T) {
	c := &Client{}
	injectAt := time.Now()
	c.lastRemoteOut.Store(injectAt.Add(-time.Second).UnixNano())

	start := time.Now()
	c.waitRemoteOutputQuietAfter(injectAt, 50*time.Millisecond, 500*time.Millisecond)
	elapsed := time.Since(start)
	if elapsed < 45*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Fatalf("expected quiet-since-inject ~50ms, got %v", elapsed)
	}
}

func TestWaitRemoteOutputQuietAfterPostInject(t *testing.T) {
	c := &Client{}
	injectAt := time.Now().Add(-50 * time.Millisecond)
	c.touchRemoteOut()

	start := time.Now()
	c.waitRemoteOutputQuietAfter(injectAt, 80*time.Millisecond, time.Second)
	if time.Since(start) < 80*time.Millisecond {
		t.Fatalf("returned before quiet period: %v", time.Since(start))
	}
}

// TestFinishRemoteSetup_RefreshOnly (W6) does not print a static prompt; it refreshes remote PS1.
func TestFinishRemoteSetup_RefreshOnly(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, "u", "h", false)

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if strings.Contains(got, RemotePrompt("u", "h")) {
		t.Fatalf("stdout should not contain static placeholder prompt: %q", got)
	}
	if !c.remotePromptShown.Load() {
		t.Fatal("remotePromptShown should be set by client paint before unpause")
	}
	if !strings.Contains(got, "AiT u@h:") {
		t.Fatalf("stdout should contain client-painted prompt, got %q", got)
	}
	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin want prompt redraw inject, got %q", stdin.String())
	}
}

func TestFinishRemoteSetupShowsPromptOnceWhenPaused(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	promptLine := "AiT u@h:/work> "
	_, _ = c.scroll.Write([]byte(promptLine))

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, "u", "h", false)

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if stdin.Len() != 0 {
		t.Fatalf("stdin should be empty when marker already buffered, got %q", stdin.String())
	}
	if !c.remotePromptShown.Load() {
		t.Fatal("expected remotePromptShown")
	}
	if !strings.Contains(got, promptLine) {
		t.Fatalf("stdout should contain buffered prompt %q, got %q", promptLine, got)
	}
}

// TestFinishRemoteSetup_EmitsBufferedPromptStdout (W15) flushes scrollback prompt to stdout without stdin refresh.
func TestFinishRemoteSetup_EmitsBufferedPromptStdout(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	promptLine := "AiT u@h:/root> "
	_, _ = c.scroll.Write([]byte("setup noise\n" + promptLine))

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, "u", "h", false)

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if stdin.Len() != 0 {
		t.Fatalf("stdin should be empty when marker already buffered, got %q", stdin.String())
	}
	if !strings.Contains(got, promptLine) {
		t.Fatalf("stdout should contain buffered prompt %q, got %q", promptLine, got)
	}
	if strings.Contains(got, RemotePrompt("u", "h")) {
		t.Fatalf("stdout should not contain static placeholder prompt: %q", got)
	}
}

func TestLastRemotePromptLine(t *testing.T) {
	scroll := "leaked export\nAiT u@h:/tmp:main> "
	if got := lastRemotePromptLine(scroll, "u", "h"); got != "AiT u@h:/tmp:main> " {
		t.Fatalf("got %q", got)
	}
	if got := lastRemotePromptLine("noise\rAiT u@h:/work> ", "u", "h"); got != "AiT u@h:/work> " {
		t.Fatalf("carriage-return prompt got %q", got)
	}
	if got := lastRemotePromptLine("no prompt here", "u", "h"); got != "" {
		t.Fatalf("want empty, got %q", got)
	}
}

// TestFinishRemoteSetup_RefreshPromptReachStdout (W16) paints prompt before unpause and drops duplicate refresh lines.
func TestFinishRemoteSetup_RefreshPromptReachStdout(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, "u", "h", false)

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if !c.remotePromptShown.Load() {
		t.Fatal("remotePromptShown should be set by client paint before unpause")
	}
	if !strings.Contains(got, "AiT u@h:") {
		t.Fatalf("stdout should contain client-painted prompt, got %q", got)
	}
	// After the AiT prompt is shown, duplicate AiT prompt chunks are dropped (W4).
	promptLine := "AiT u@h:/data> "
	if got := c.filterRemoteOutputChunk([]byte(promptLine), "u", "h"); got != nil {
		t.Fatalf("duplicate AiT prompt should be dropped after shown, got %q", got)
	}
}

// TestFinishRemoteSetup_EmitFailFallsBackRefresh (W16) falls back to refresh when marker exists without a prompt line.
func TestFinishRemoteSetup_EmitFailFallsBackRefresh(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	_, _ = c.scroll.Write([]byte("partial AiT u@h: without closing"))

	c.finishRemoteSetup(stdin, "u", "h", false)

	if !c.remotePromptShown.Load() {
		t.Fatal("remotePromptShown should be set by client paint when scroll emit fails")
	}
	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin want prompt redraw inject after emit fail, got %q", stdin.String())
	}
}

// TestCopyOutDropsDuplicatePromptAfterShown (TM11d/W4): after the AiT prompt is shown,
// duplicate prompt chunks are dropped while normal command output passes through.
func TestCopyOutDropsDuplicatePromptAfterShown(t *testing.T) {
	c := &Client{}
	c.remotePromptShown.Store(true)
	user, host := "u", "h"

	if got := c.filterRemoteOutputChunk([]byte("AiT u@h:/tmp:main> "), user, host); got != nil {
		t.Fatalf("duplicate AiT prompt should be dropped, got %q", got)
	}
	if got := c.filterRemoteOutputChunk([]byte("ls\r\n"), user, host); got == nil {
		t.Fatal("normal command output should pass through")
	}
}

func TestIsDuplicatePromptOutput(t *testing.T) {
	line := "AiT u@h:/data:main> "
	if !isDuplicatePromptOutput([]byte(line), "u", "h") {
		t.Fatal("exact prompt chunk should be duplicate")
	}
	if !isDuplicatePromptOutput([]byte("\r\n"+line+"\n"), "u", "h") {
		t.Fatal("prompt with surrounding whitespace should be duplicate")
	}
	if isDuplicatePromptOutput([]byte(line+"ls\n"), "u", "h") {
		t.Fatal("prompt plus command should not be duplicate")
	}
}

// TestFinishRemoteSetup_RequiresVisiblePromptBeforeUnpause (P1h) emits prompt before unpause.
func TestFinishRemoteSetup_RequiresVisiblePromptBeforeUnpause(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	promptLine := "AiT u@h:/work> "
	_, _ = c.scroll.Write([]byte(promptLine))

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, "u", "h", false)

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if !c.remotePromptShown.Load() {
		t.Fatal("remotePromptShown must be set before unpause")
	}
	if c.pauseRemoteOut.Load() {
		t.Fatal("pauseRemoteOut should be cleared after setup")
	}
	if !strings.Contains(got, promptLine) {
		t.Fatalf("stdout should contain prompt %q, got %q", promptLine, got)
	}
}

// TestSeedReplayScrollback_FiltersDevNullFragments (P2h) seeds sanitized scroll without stdout.
func TestSeedReplayScrollback_FiltersDevNullFragments(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	data := strings.Join([]string{
		"Welcome to TencentOS",
		"#",
		"S",
		"a",
		"dev/null",
		"hello world",
	}, "\n")
	c.seedReplayScrollback(data, testSSHUser, testPrimaryHostFQDN)

	if buf.Len() != 0 {
		t.Fatalf("seed replay must not write stdout, got %q", buf.String())
	}
	scroll := c.scroll.String()
	for _, leak := range []string{"#", "dev/null", "\nS\n", "\na\n"} {
		if strings.Contains(scroll, leak) {
			t.Fatalf("seed scroll should not contain %q, got %q", leak, scroll)
		}
	}
	if !strings.Contains(scroll, "hello world") {
		t.Fatalf("seed scroll should keep content, got %q", scroll)
	}
}

// TestReplayFilteredScrollback_CollapsesBlankRuns (P2i) folds excessive blank lines on replay.
func TestReplayFilteredScrollback_CollapsesBlankRuns(t *testing.T) {
	got := collapseBlankRuns("a\n\n\n\nb")
	if strings.Count(got, "\n\n") != 1 {
		t.Fatalf("collapseBlankRuns = %q", got)
	}
}

// TestSeedReplayScrollback_NoStdout (P2k) seeds scroll only; stdout stays empty.
func TestSeedReplayScrollback_NoStdout(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)

	data := "Welcome\r\nhello world"
	c.seedReplayScrollback(data, "u", "h")

	if buf.Len() != 0 {
		t.Fatalf("seed replay must not write stdout, got %q", buf.String())
	}
	if strings.Contains(c.scroll.String(), "stty -echo") {
		t.Fatalf("seed scroll should not contain setup leak, got %q", c.scroll.String())
	}
	if !strings.Contains(c.scroll.String(), "hello world") {
		t.Fatalf("scroll should seed replay data, got %q", c.scroll.String())
	}
}

// TestFinishRemoteSetup_NoScrollPrompt_InjectRedraw (P1) injects redraw when scrollback has no prompt.
func TestFinishRemoteSetup_NoScrollPrompt_InjectRedraw(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	start := time.Now()
	c.finishRemoteSetup(stdin, "u", "h", false)
	if time.Since(start) > 4*time.Second {
		t.Fatalf("finishRemoteSetup should not block for full prompt max wait: %v", time.Since(start))
	}
	if !stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("stdin want prompt redraw inject, got %q", stdin.String())
	}
}

func TestWaitForRemoteAITPrompt(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	_, _ = c.scroll.Write([]byte("noise\nAiT u@h:/home> "))

	start := time.Now()
	c.waitForRemoteAITPrompt("u", "h", time.Second)
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("should return quickly when marker present: %v", time.Since(start))
	}
}

// TestWaitForShellReadyBeforeInject_NativePrompt (P1e) returns quickly when scroll has a native prompt and output is quiet.
func TestWaitForShellReadyBeforeInject_NativePrompt(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	attachAt := time.Now().Add(-500 * time.Millisecond)
	_, _ = c.scroll.Write([]byte("motd\nroot@remote [~] ➜"))
	c.touchRemoteOut()
	time.Sleep(remoteOutputQuietPeriod + 30*time.Millisecond)

	start := time.Now()
	c.waitForShellReadyBeforeInject(attachAt, testSSHUser, testPrimaryHostFQDN, shellReadyMaxWait)
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("should return quickly when native prompt present and quiet: %v", time.Since(start))
	}
}

// TestTriggerRemotePrecmd_WritesSpaceNewline (P1f) sends a history-ignored blank line to stdin.
func TestTriggerRemotePrecmd_WritesSpaceNewline(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	stdin := &bytes.Buffer{}
	c.triggerRemotePrecmd(stdin)
	if stdin.String() != " \n" {
		t.Fatalf("stdin = %q, want %q", stdin.String(), " \n")
	}
}

// TestEmitBufferedRemotePromptIfReady_NoPrecmd (P1g) does not trigger precmd from emit helper.
func TestEmitBufferedRemotePromptIfReady_NoPrecmd(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.pauseRemoteOut.Store(true)
	stdin := &bytes.Buffer{}

	c.emitBufferedRemotePromptIfReady(stdin, testSSHUser, testPrimaryHostFQDN)

	if strings.HasPrefix(stdin.String(), " \n") {
		t.Fatalf("emit helper should not trigger precmd, got %q", stdin.String())
	}
}

// TestWaitRemoteOutputQuietSince_BeforeShellInject (P1d) verifies quiet-since-attach fallback timing.
func TestWaitRemoteOutputQuietSince_BeforeShellInject(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	attachAt := time.Now()
	c.touchRemoteOut()

	start := time.Now()
	c.waitRemoteOutputQuietSince(attachAt, remoteOutputQuietPeriod, shellQuietMaxWait)
	elapsed := time.Since(start)
	if elapsed < remoteOutputQuietPeriod {
		t.Fatalf("should wait at least quiet period after MOTD, got %v", elapsed)
	}
	if elapsed > shellQuietMaxWait+500*time.Millisecond {
		t.Fatalf("should not exceed shell quiet max wait, got %v", elapsed)
	}
}

// TestFinishRemoteSetup_EmitsNativePromptFallback (P1c) writes native prompt only when setup sentinel is absent.
func TestFinishRemoteSetup_EmitsNativePromptFallback(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	nativeLine := "root@remote [~] ➜"
	_, _ = c.scroll.Write([]byte("motd\n" + nativeLine))
	// No RemotePromptReadySentinel — native fallback is allowed.

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, testSSHUser, testPrimaryHostFQDN, false)

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if !strings.Contains(got, nativeLine) {
		t.Fatalf("stdout should contain native prompt %q, got %q", nativeLine, got)
	}
	if !c.remotePromptShown.Load() {
		t.Fatal("remotePromptShown should be set after native emit fallback")
	}
}

func TestLastNativePromptLine(t *testing.T) {
	scroll := "motd\nroot@remote [~] ➜\n"
	if got := lastNativePromptLine(scroll); got != "root@remote [~] ➜" {
		t.Fatalf("got %q", got)
	}
	if got := lastNativePromptLine("AiT u@h:/tmp> "); got != "" {
		t.Fatalf("want empty for AiT-only scroll, got %q", got)
	}
}

// TestFinishRemoteSetup_ClientPaintWhenSentinelOnlyNative (P1i) paints AiT prompt instead of native fallback.
func TestFinishRemoteSetup_ClientPaintWhenSentinelOnlyNative(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	nativeLine := "root@remote [~] ->"
	_, _ = c.scroll.Write([]byte(nativeLine + "\n" + prompt.RemotePromptReadySentinel + "\n"))
	c.remotePromptReadySeen.Store(true)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, testSSHUser, testPrimaryHostFQDN, false)

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if strings.Contains(got, nativeLine) {
		t.Fatalf("stdout should not contain native prompt %q, got %q", nativeLine, got)
	}
	if !strings.Contains(got, "AiT root@"+testPrimaryHostFQDN+":") {
		t.Fatalf("stdout should contain client-painted AiT prompt, got %q", got)
	}
}

// TestEnsureFirstPromptVisible_NoPrecmdWhenSetupSucceeded (P1n) paints without precmd when sentinel is ready.
func TestEnsureFirstPromptVisible_NoPrecmdWhenSetupSucceeded(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.pauseRemoteOut.Store(true)
	c.remotePromptReadySeen.Store(true)
	stdin := &bytes.Buffer{}

	if !c.ensureFirstPromptVisible(stdin, testSSHUser, testPrimaryHostFQDN) {
		t.Fatal("expected prompt visible")
	}
	if strings.Contains(stdin.String(), " \n") {
		t.Fatalf("precmd should be skipped when setup succeeded, stdin=%q", stdin.String())
	}
	if !c.remotePromptShown.Load() {
		t.Fatal("remotePromptShown should be set")
	}
}

// TestFormatActivePromptStdout_NoTrailingNewline (P1m) keeps the cursor on the prompt line.
func TestFormatActivePromptStdout_NoTrailingNewline(t *testing.T) {
	line := "AiT root@" + testPrimaryHostFQDN + ":~> "
	out := formatActivePromptStdout(line)
	if strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\r\n") {
		t.Fatalf("active prompt must not end with newline, got %q", out)
	}
	if !strings.HasPrefix(out, "\r") {
		t.Fatalf("active prompt should start with carriage return, got %q", out)
	}
	if !strings.HasSuffix(out, promptClearToEOL) {
		t.Fatalf("active prompt should clear to EOL, got %q", out)
	}
	if !strings.HasSuffix(strings.TrimSuffix(out, promptClearToEOL), "> ") {
		t.Fatalf("active prompt should end at prompt suffix, got %q", out)
	}
}

// TestFinishRemoteSetup_SinglePrecmdTrigger (P1k) sends at most one precmd blank line during setup finish.
func TestFinishRemoteSetup_SinglePrecmdTrigger(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	c.finishRemoteSetup(stdin, testSSHUser, testPrimaryHostFQDN, false)

	if strings.Count(stdin.String(), " \n") != 1 {
		t.Fatalf("want exactly one precmd trigger, stdin=%q", stdin.String())
	}
}

// TestFinishRemoteSetup_HardGateUnpause (P1l) never clears pauseRemoteOut without remotePromptShown.
func TestFinishRemoteSetup_HardGateUnpause(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	c.finishRemoteSetup(stdin, "u", "h", false)

	if !c.remotePromptShown.Load() {
		t.Fatal("remotePromptShown must be true before unpause")
	}
	if c.pauseRemoteOut.Load() {
		t.Fatal("pauseRemoteOut should clear once prompt is shown")
	}
}
