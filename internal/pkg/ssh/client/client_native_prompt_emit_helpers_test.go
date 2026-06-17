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

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestExtractNativePromptForEmit_BashBracketedPaste(t *testing.T) {
	raw := "\x1b[?2004h\x1b]0;root@devhost: ~\x07root@devhost:~# "
	got := extractNativePromptForEmit(raw)
	if got != "root@devhost:~#" {
		t.Fatalf("got %q", got)
	}
}

func TestExtractNativePromptForEmit_ZshRightPrompt(t *testing.T) {
	raw := "\x1b[01;33mroot@remote\x1b[00m [\x1b[01;36m~\x1b[00m] \x1b[33m➜ \x1b[00m \x1b[K\x1b[117C[23:35:18]\x1b[127D"
	got := extractNativePromptForEmit(raw)
	if !strings.Contains(got, "root@remote") || !strings.Contains(got, "➜") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "[23:35:18]") || strings.Contains(got, "\x1b[117C") {
		t.Fatalf("right prompt tail should be stripped, got %q", got)
	}
}

func TestIsNativePromptLike_TrailingSpaceAfterArrow(t *testing.T) {
	line := "\x1b[33mroot@remote\x1b[00m [~] ➜   "
	if !isNativePromptLike(line) {
		t.Fatalf("trailing spaces after arrow should still match")
	}
}

func TestExtractNativePromptForEmit_RejectsAITPrompt(t *testing.T) {
	if got := extractNativePromptForEmit("AiT u@h:/tmp> "); got != "" {
		t.Fatalf("AiT prompt must not emit as native, got %q", got)
	}
	if got := lastNativePromptLine("AiT u@h:/tmp> "); got != "" {
		t.Fatalf("lastNativePromptLine must ignore AiT-only scroll, got %q", got)
	}
}

func TestTrimPartialCSIPrefix_PreservesAIT(t *testing.T) {
	if got := trimPartialCSIPrefix("AiT u@h:/tmp>"); got != "AiT u@h:/tmp>" {
		t.Fatalf("got %q", got)
	}
	if got := trimPartialCSIPrefix("2Jroot@devhost:~#"); got != "root@devhost:~#" {
		t.Fatalf("got %q", got)
	}
}

func TestEmitBufferedNativePrompt_SanitizedBash(t *testing.T) {
	c := testClientWithPrompt("root", "dev.testPrimaryHostFQDN")
	_, _ = c.scroll.Write([]byte("\x1b[?2004h\x1b]0;root@devhost: ~\x07root@devhost:~# "))

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	if !c.emitBufferedNativePrompt() {
		t.Fatal("expected emit")
	}
	got := buf.String()
	if strings.Contains(got, "\x1b[?2004h") || strings.Contains(got, "\x1b]0;") {
		t.Fatalf("stdout must not include bracketed paste / OSC, got %q", got)
	}
	if !strings.Contains(got, "root@devhost:~#") {
		t.Fatalf("stdout = %q", got)
	}
}

func TestSyncNativeRemoteShellAfterMeta_NoPrintPn(t *testing.T) {
	c := &Client{}
	var stdin bytes.Buffer
	c.syncNativeRemoteShellAfterMeta(&stdin)
	if strings.Contains(stdin.String(), `print -Pn "$PROMPT"`) {
		t.Fatalf("native sync must not use print -Pn, stdin = %q", stdin.String())
	}
}

func TestExtractNativePromptForEmit_ZshSGRTerminator(t *testing.T) {
	raw := "\x1b[01;33mroot@remote\x1b[00m [\x1b[01;36m~\x1b[00m] \x1b[33m➜ \x1b[00m \x1b[K\x1b[118C[0:07:03]\x1b[127D"
	got := extractNativePromptForEmit(raw)
	plain := strings.TrimSpace(stripANSIForPromptMatch(got))
	if strings.HasPrefix(plain, "mroot") || strings.HasPrefix(plain, "33m") {
		t.Fatalf("SGR tail must not prefix username, plain=%q got=%q", plain, got)
	}
	if !strings.HasPrefix(plain, "root@remote") {
		t.Fatalf("plain=%q got=%q", plain, got)
	}
	if strings.Contains(plain, "[0:07:03]") {
		t.Fatalf("zsh right prompt tail should be stripped, plain=%q", plain)
	}
}

// TestExtractNativePromptForEmit_ChainedSSHClearScreen: scrollback split CSI must not leak "2J".
func TestExtractNativePromptForEmit_ChainedSSHClearScreen(t *testing.T) {
	for _, raw := range []string{
		"\x1b[H\x1b[2Jroot@devhost:~# ",
		"2Jroot@devhost:~# ",
		"33mroot@remote:~# ",
		"\b \b\b\be \b\x1b[H\x1b[2J\x1b[0m\x1b[27m\x1b[24m\x1b[J\x1b[01;33mroot@remote\x1b[00m [\x1b[01;36m~\x1b[00m] \x1b[33m➜ \x1b[00m \x1b[K",
	} {
		got := extractNativePromptForEmit(raw)
		plain := strings.TrimSpace(stripANSIForPromptMatch(got))
		if strings.Contains(plain, "2J") || strings.HasPrefix(plain, "33m") || strings.HasPrefix(plain, "mroot") {
			t.Fatalf("raw=%q plain=%q got=%q", raw, plain, got)
		}
	}
}

func TestEmitBufferedNativePrompt_ChainedSSHNoCSIJunk(t *testing.T) {
	c := testClientWithPrompt("root", "dev.testPrimaryHostFQDN")
	_, _ = c.scroll.Write([]byte("\x1b[H\x1b[2Jroot@devhost:~# "))

	var buf bytes.Buffer
	c.interactiveOut = terminal.NewCRLFWriter(&buf)
	if !c.emitBufferedNativePrompt() {
		t.Fatal("expected emit")
	}
	got := buf.String()
	if strings.Contains(got, "2J") {
		t.Fatalf("stdout must not include split CSI tail, got %q", got)
	}
	if !strings.Contains(got, "root@devhost:~#") {
		t.Fatalf("stdout = %q", got)
	}
}
