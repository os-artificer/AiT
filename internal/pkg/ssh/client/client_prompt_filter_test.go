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

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestIsNativePromptLike_ANSI (P4) matches oh-my-zsh prompts with color codes.
func TestIsNativePromptLike_ANSI(t *testing.T) {
	line := "\x1b[1mroot@remote\x1b[0m [~] \x1b[1;33m➜\x1b[0m "
	if !isNativePromptLike(line) {
		t.Fatalf("expected ANSI native prompt match, line=%q", line)
	}
}

// TestIsNativePromptLike_RightPromptStatus keeps matching when zsh right-prompt time suffix is present.
func TestIsNativePromptLike_RightPromptStatus(t *testing.T) {
	line := "\x1b[1;33mroot@remote\x1b[0m [~] ➜ \x1b[K\x1b[180C[17:13:05]\x1b[190D"
	if !isNativePromptLike(line) {
		t.Fatalf("expected native prompt with right status to match, line=%q", line)
	}
}

// TestIsNativePromptLike_OSCTitleChunk ignores OSC title updates when identifying prompts.
func TestIsNativePromptLike_OSCTitleChunk(t *testing.T) {
	line := "\x1b]2;root@remote:~\a\x1b]1;~\a"
	if isNativePromptLike(line) {
		t.Fatalf("OSC-only chunk must not be treated as native prompt, line=%q", line)
	}
}

// TestIsLikelyShellPromptLine_ShortBash matches short bash PS1 variants.
func TestIsLikelyShellPromptLine_ShortBash(t *testing.T) {
	if !isLikelyShellPromptLine("u@h:#") {
		t.Fatal("expected short bash prompt match")
	}
	if isLikelyShellPromptLine("1  " + testPrimaryHostFQDN + "  active") {
		t.Fatal("session table row must not match shell prompt")
	}
}

// TestRemoveManagedNativePromptLines_MultiLine (P5) strips native prompt in command output chunks.
func TestRemoveManagedNativePromptLines_MultiLine(t *testing.T) {
	chunk := []byte("ls\r\n\x1b[1mroot@remote\x1b[0m [~] \x1b[1;33m➜\x1b[0m ")
	out, removed := removeManagedNativePromptLines(chunk)
	if !removed {
		t.Fatal("expected native prompt removal")
	}
	if !strings.Contains(string(out), "ls") {
		t.Fatalf("command output should remain, got %q", out)
	}
	if strings.Contains(string(out), "remote") {
		t.Fatalf("native prompt should be removed, got %q", out)
	}
}

// TestFilterRemoteOutputChunkRepaint_SuppressesNativeAfterSetup: after AiT setup, native
// prompt lines are stripped and a client repaint is requested.
func TestFilterRemoteOutputChunkRepaint_SuppressesNativeAfterSetup(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)

	native := "\x1b[1mroot@remote\x1b[0m [~] \x1b[1;33m➜\x1b[0m \r\n"
	out, repaint := c.filterRemoteOutputChunkRepaint([]byte(native), testSSHUser, testPrimaryHostFQDN)
	if out != nil && strings.Contains(string(out), "remote") {
		t.Fatalf("native chunk should be stripped after AiT setup, got %q", out)
	}
	if !repaint {
		t.Fatal("expected client repaint after stripping native prompt")
	}
}

// TestFilterRemoteOutputChunkRepaint_StripsNativeWhenPromptShownOnly: when the AiT prompt
// was shown, native prompt echoes are stripped even without the ready sentinel.
func TestFilterRemoteOutputChunkRepaint_StripsNativeWhenPromptShownOnly(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)

	native := "root@remote [~] ➜ \r\n"
	out, repaint := c.filterRemoteOutputChunkRepaint([]byte(native), testSSHUser, testPrimaryHostFQDN)
	if out != nil && strings.Contains(string(out), "remote") {
		t.Fatalf("native chunk should be stripped when prompt shown, got %q", out)
	}
	if !repaint {
		t.Fatal("expected client repaint after stripping native prompt")
	}
}

// TestFilterRemoteOutputChunkRepaint_KeepsNativeWithoutSetup (P7) does not strip native lines before setup completes.
func TestFilterRemoteOutputChunkRepaint_KeepsNativeWithoutSetup(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)

	native := "root@remote [~] ➜ \r\n"
	out, repaint := c.filterRemoteOutputChunkRepaint([]byte(native), testSSHUser, testPrimaryHostFQDN)
	if repaint {
		t.Fatal("should not repaint without managed setup")
	}
	if out == nil || !strings.Contains(string(out), "remote") {
		t.Fatalf("native prompt should pass through before setup, got %v", out)
	}
}

// TestStripUnwantedPromptLines_KeepsTableDropsPrompt (P9) keeps session list rows when a
// multi-line chunk also contains a native prompt echo line.
func TestStripUnwantedPromptLines_KeepsTableDropsPrompt(t *testing.T) {
	table := "#  ID        HOST           USER  STATE\r\n1  abc12345  " + testPrimaryHostFQDN + "   root  active\r\n"
	native := "\x1b[1mroot@remote\x1b[0m [~] \x1b[1;33m➜\x1b[0m \r\n"
	chunk := []byte(table + native)

	out, nativeRemoved := stripUnwantedPromptLines(chunk, testSSHUser, testPrimaryHostFQDN, true, true)
	if !nativeRemoved {
		t.Fatal("expected prompt line removal")
	}
	if !strings.Contains(string(out), "abc12345") {
		t.Fatalf("session table should remain, got %q", out)
	}
	if strings.Contains(string(out), "remote") && !strings.Contains(string(out), testPrimaryHostFQDN) {
		t.Fatalf("native prompt host should be removed, got %q", out)
	}
}

// TestStripUnwantedPromptLines_KeepsClearSequenceWithPrompt (P12) preserves clear-screen
// output even when prompt de-duplication is active.
func TestStripUnwantedPromptLines_KeepsClearSequenceWithPrompt(t *testing.T) {
	line := "\x1b[2J\x1b[HAiT root@testDevHostFQDN:~> "
	out, promptRemoved := stripUnwantedPromptLines([]byte(line), testSSHUser, testDevHostFQDN, true, true)
	if promptRemoved {
		t.Fatal("clear-sequence line should not be stripped as duplicate prompt")
	}
	if string(out) != line {
		t.Fatalf("clear-sequence line should remain intact, got %q", out)
	}
}

// TestFilterRemoteOutputChunkRepaint_RepaintsAfterAITDuplicateStrip: command output is kept
// while a trailing native prompt is stripped and a client repaint is requested.
func TestFilterRemoteOutputChunkRepaint_RepaintsAfterAITDuplicateStrip(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)

	native := "root@remote [~] ➜ "
	chunk := []byte("go  ips  snap\r\n" + native)
	out, repaint := c.filterRemoteOutputChunkRepaint(chunk, testSSHUser, testPrimaryHostFQDN)
	if out == nil || !strings.Contains(string(out), "go") {
		t.Fatalf("command output should remain, got %q", out)
	}
	if strings.Contains(string(out), "root@remote [~] ➜") {
		t.Fatalf("native prompt should be stripped, got %q", out)
	}
	if !repaint {
		t.Fatal("expected client repaint after stripping native prompt")
	}
}

// TestIsDuplicatePromptLine_AIT (P8) still recognizes AiT prompt lines with ANSI stripped.
func TestIsDuplicatePromptLine_AIT(t *testing.T) {
	line := prompt.RemoteFormat(testSSHUser, testPrimaryHostFQDN, "~", "")
	if !isDuplicatePromptLine(line, testSSHUser, testPrimaryHostFQDN) {
		t.Fatalf("AiT prompt line should be duplicate: %q", line)
	}
}

// TestLastNativePromptLineForHost_SkipsMismatchedHost (N7a) ignores inner native prompts
// when the active host is the outer dial target.
func TestLastNativePromptLineForHost_SkipsMismatchedHost(t *testing.T) {
	scroll := "motd\nroot@remote [~] ->\n"
	if got := lastNativePromptLineForHost(scroll, testDevHostFQDN); got != "" {
		t.Fatalf("inner native prompt should not match dial host, got %q", got)
	}
	if got := lastNativePromptLineForHost(scroll, testPrimaryHostFQDN); got != "root@remote [~] ->" {
		t.Fatalf("inner native prompt should match inner host, got %q", got)
	}
}

// TestEmitMetaPromptStdout_SkipsInnerNativeAfterNestedClose (N7b) does not re-emit a stale
// inner native prompt whose host does not match the active dial target; AiT mode paints
// the managed prompt for the dial host instead.
func TestEmitMetaPromptStdout_SkipsInnerNativeAfterNestedClose(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testDevHostFQDN)
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	_, _ = c.scroll.Write([]byte("Connection to " + testInnerHostFQDN + " closed.\r\nroot@remote [~] ->\r\n"))

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	if !c.emitMetaPromptStdout(testSSHUser, testDevHostFQDN) {
		t.Fatal("expected AiT prompt paint for dial host")
	}
	got := buf.String()
	if strings.Contains(got, "root@remote") {
		t.Fatalf("should not repaint inner native prompt, got %q", got)
	}
	if !strings.Contains(got, "AiT ") {
		t.Fatalf("expected AiT-managed dial-host prompt, got %q", got)
	}
}

// TestEmitMetaPromptStdout_NoNativeFallbackWhenPromptShown (P11): when the AiT prompt is
// managed, emit paints AiT instead of falling back to a native scrollback line alone.
func TestEmitMetaPromptStdout_NoNativeFallbackWhenPromptShown(t *testing.T) {
	c := testClientWithPrompt(testSSHUser, testPrimaryHostFQDN)
	c.remotePromptShown.Store(true)
	c.remotePromptReadySeen.Store(true)
	_, _ = c.scroll.Write([]byte("root@remote [~] ➜ \r\n"))

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	if !c.emitMetaPromptStdout(testSSHUser, testPrimaryHostFQDN) {
		t.Fatal("expected AiT prompt emit")
	}
	got := buf.String()
	if !strings.Contains(got, "AiT ") {
		t.Fatalf("expected AiT-managed prompt paint, got %q", got)
	}
}
