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

// TestSingleHopPromptIdentityEqualsDial (W12) verifies active prompt starts as dial target.
func TestSingleHopPromptIdentityEqualsDial(t *testing.T) {
	c := &Client{opts: Options{User: "root", Host: "proxy.example"}}
	c.initActivePrompt()
	if c.promptUser() != "root" || c.promptHost() != "proxy.example" {
		t.Fatalf("active = %s@%s want root@proxy.example", c.promptUser(), c.promptHost())
	}
}

// TestDuplicatePromptOutput_SingleHopUnchanged (W12) uses dial identity for duplicate detection.
func TestDuplicatePromptOutput_SingleHopUnchanged(t *testing.T) {
	c := &Client{opts: Options{User: "u", Host: "proxy.h"}}
	c.initActivePrompt()
	if !isDuplicatePromptOutput([]byte("AiT u@proxy.h:/tmp> "), c.promptUser(), c.promptHost()) {
		t.Fatal("single-hop duplicate should use dial host")
	}
}

// TestDuplicatePromptOutput_ActiveHostAfterNested (W12) uses active host after nested update.
func TestDuplicatePromptOutput_ActiveHostAfterNested(t *testing.T) {
	c := &Client{opts: Options{User: "root", Host: "proxy.h"}}
	c.initActivePrompt()
	c.activePromptUser = "root"
	c.activePromptHost = "inner.example"
	if !isDuplicatePromptOutput([]byte("AiT root@inner.example:/tmp> "), c.promptUser(), c.promptHost()) {
		t.Fatal("nested active host should drive duplicate detection")
	}
	if isDuplicatePromptOutput([]byte("AiT root@inner.example:/tmp> "), "root", "proxy.h") {
		t.Fatal("dial host must not match inner prompt line")
	}
}

// TestParseSSHCommandLine_PortAndUserHost (W13) parses ssh -p PORT user@host.
func TestParseSSHCommandLine_PortAndUserHost(t *testing.T) {
	user, host, ok := parseSSHCommandLine("ssh -p 2222 root@"+testInnerHostFQDN, testSSHUser)
	if !ok {
		t.Fatal("expected ok")
	}
	if user != testSSHUser || host != testInnerHostFQDN {
		t.Fatalf("got %s@%s", user, host)
	}

	user, host, ok = parseSSHCommandLine("ssh -p2222 root@"+testInnerHostFQDN, testSSHUser)
	if !ok || user != testSSHUser || host != testInnerHostFQDN {
		t.Fatalf("combined -p: %s@%s ok=%v", user, host, ok)
	}
}

// TestParseSSHCommandLine_SkipComplex (W13) skips ProxyJump and other unsupported forms.
func TestParseSSHCommandLine_SkipComplex(t *testing.T) {
	if _, _, ok := parseSSHCommandLine("ssh -J jump final", "root"); ok {
		t.Fatal("expected -J to skip")
	}
	if _, _, ok := parseSSHCommandLine("ls -la", "root"); ok {
		t.Fatal("non-ssh should skip")
	}
}

// TestForwardStdin_SSHEnablesChainMode (W14) arms nested await on ssh line submit;
// chain stays off until shell-ready inject (W14b).
func TestForwardStdin_SSHEnablesChainMode(t *testing.T) {
	c := &Client{
		opts:   Options{User: "root", Host: "proxy.example"},
		scroll: &terminal.Buffer{},
	}
	c.initActivePrompt()
	c.onInteractiveLineSubmitted("ssh root@inner.example")
	c.nestedMu.Lock()
	awaiting := c.nestedState == nestedPromptAwaitingShell
	c.nestedMu.Unlock()
	if !awaiting {
		t.Fatal("expected awaiting shell after ssh line")
	}
	if c.ChainMode() {
		t.Fatal("chain should stay off until inject completes")
	}

	c.nestedMu.Lock()
	c.nestedMOTDSeen = true
	c.nestedShellReadySeen = true
	c.nestedMOTDAt = time.Now().Add(-shellReadyDelay - time.Second)
	c.nestedMu.Unlock()
	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())
	if !c.tryNestedPromptInject() {
		t.Fatal("expected inject after shell ready")
	}
	if !c.ChainMode() {
		t.Fatal("expected chain mode on after nested ready inject")
	}
	if c.promptHost() != "inner.example" {
		t.Fatalf("active host = %q, want inner.example", c.promptHost())
	}
}

// TestRunInteractive_SingleHopChainOff (W14) new client defaults chain off.
func TestRunInteractive_SingleHopChainOff(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	if c.ChainMode() {
		t.Fatal("new client chain mode should be off")
	}
}

// TestForwardStdin_SSHTriggersNestedPrompt (W11) injects PS1 after MOTD, shell prompt, and quiet period.
func TestForwardStdin_SSHTriggersNestedPrompt(t *testing.T) {
	c := &Client{
		opts:               Options{User: "root", Host: "proxy.example"},
		scroll:             &terminal.Buffer{},
		remotePromptDirAbs: false,
	}
	c.initActivePrompt()
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin

	c.onInteractiveLineSubmitted("ssh root@inner.example")

	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())
	c.maybeInjectNestedPrompt([]byte("Last login: Thu\n"))
	if c.ChainMode() {
		t.Fatal("expected no inject on MOTD alone (W11a)")
	}
	if stdin.Len() > 0 {
		t.Fatal("expected no stdin write on MOTD alone")
	}

	c.nestedMu.Lock()
	c.nestedMOTDAt = time.Now().Add(-shellReadyDelay - time.Second)
	c.nestedMu.Unlock()
	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())

	c.maybeInjectNestedPrompt([]byte("root@inner.example [~] ->"))

	if !c.ChainMode() {
		t.Fatal("expected chain mode on after nested inject")
	}
	if c.promptHost() != "inner.example" {
		t.Fatalf("active host = %q", c.promptHost())
	}
	out := stdin.String()
	if !strings.Contains(out, "AiT root@inner.example:") {
		t.Fatalf("stdin missing nested PS1: %q", out)
	}
	if !strings.Contains(out, `basename "${PWD:-/}"`) {
		t.Fatalf("stdin missing short dir segment: %q", out)
	}
}

// TestNestedPrompt_MOTDOnlyNoInject (W11a) does not inject after MOTD without shell prompt.
func TestNestedPrompt_MOTDOnlyNoInject(t *testing.T) {
	c := &Client{
		opts:   Options{User: "root", Host: "proxy.example"},
		scroll: &terminal.Buffer{},
	}
	c.initActivePrompt()
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin

	c.onInteractiveLineSubmitted("ssh root@inner.example")
	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())
	c.maybeInjectNestedPrompt([]byte("Welcome to TencentOS\nLast login: Thu\n"))

	if c.ChainMode() {
		t.Fatal("MOTD alone must not trigger inject")
	}
	if stdin.Len() > 0 {
		t.Fatal("MOTD alone must not write setup commands")
	}
}

// TestNestedPrompt_NoMOTDFallbackInject (W11c) injects after max wait without MOTD.
func TestNestedPrompt_NoMOTDFallbackInject(t *testing.T) {
	c := &Client{
		opts:               Options{User: "root", Host: "proxy.example"},
		scroll:             &terminal.Buffer{},
		remotePromptDirAbs: false,
	}
	c.initActivePrompt()
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin

	c.onInteractiveLineSubmitted("ssh root@inner.example")
	c.nestedMu.Lock()
	c.nestedAwaitingSince = time.Now().Add(-nestedPromptMaxWait - time.Second)
	c.nestedMu.Unlock()
	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())

	c.maybeInjectNestedPrompt([]byte("connected\n"))

	if !c.ChainMode() {
		t.Fatal("expected inject after max wait without MOTD")
	}
	if !strings.Contains(stdin.String(), "AiT root@inner.example:") {
		t.Fatalf("stdin missing nested PS1: %q", stdin.String())
	}
}

// TestParseSSHConnectionClosedHost (N6) parses OpenSSH client disconnect banners.
func TestParseSSHConnectionClosedHost(t *testing.T) {
	got := parseSSHConnectionClosedHost("Connection to " + testInnerHostFQDN + " closed.\r\n")
	if got != testInnerHostFQDN {
		t.Fatalf("host = %q", got)
	}
	if parseSSHConnectionClosedHost("logout\r\n") != "" {
		t.Fatal("unrelated chunk should not match")
	}
}

// TestOnNestedSSHSessionClosed_RestoresDial (N6) resets active prompt after nested exit.
func TestOnNestedSSHSessionClosed_RestoresDial(t *testing.T) {
	c := &Client{opts: Options{User: testSSHUser, Host: testDevHostFQDN}}
	c.initActivePrompt()
	c.activePromptHost = testInnerHostFQDN
	c.nestedRemoteHost = testInnerHostFQDN
	c.SetChainMode(true)

	if !c.onNestedSSHSessionClosed(testInnerHostFQDN) {
		t.Fatal("expected nested close handling")
	}
	if c.promptHost() != testDevHostFQDN {
		t.Fatalf("active host = %q", c.promptHost())
	}
	if c.ChainMode() {
		t.Fatal("chain mode should be off after nested exit")
	}
	if c.nestedRemoteHost != "" {
		t.Fatal("nested remote host should be cleared")
	}
}

// TestHandleRemoteOutput_NestedCloseNoRemoteInject (N6b) native prompt mode re-emits
// from scrollback without injecting stty/print -Pn after nested ssh exit.
func TestHandleRemoteOutput_NestedCloseNoRemoteInject(t *testing.T) {
	enableRemoteSystemPromptForTest(t)
	c := testClientWithPrompt(testSSHUser, testDevHostFQDN)
	c.activePromptHost = testInnerHostFQDN
	c.nestedRemoteHost = testInnerHostFQDN
	c.SetChainMode(true)
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin
	_, _ = c.scroll.Write([]byte("root@" + testDevHostShort + ":~# \r\n"))

	c.handleRemoteOutput([]byte("Connection to " + testInnerHostFQDN + " closed.\r\nroot@" + testDevHostShort + ":~# "))

	if stdinHasPromptRedraw(stdin.String()) {
		t.Fatalf("nested close must not inject prompt redraw, stdin = %q", stdin.String())
	}
	if c.promptHost() != testDevHostFQDN {
		t.Fatalf("active host = %q", c.promptHost())
	}
}

// TestRepaintAfterNestedSSHClose_NoSyncWhenScrollbackEmpty (N6c) does not inject a
// remote redraw when scrollback lacks an outer prompt yet.
func TestRepaintAfterNestedSSHClose_NoSyncWhenScrollbackEmpty(t *testing.T) {
	enableRemoteSystemPromptForTest(t)
	c := testClientWithPrompt(testSSHUser, testDevHostFQDN)
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin
	_, _ = c.scroll.Write([]byte("Connection to " + testInnerHostFQDN + " closed.\r\n"))

	c.repaintAfterNestedSSHClose(testSSHUser, testDevHostFQDN)

	if stdin.Len() != 0 {
		t.Fatalf("must not inject remote sync when scrollback has no prompt, stdin = %q", stdin.String())
	}
}

// TestScrollChunkLooksLikeShellPrompt detects common inner-shell prompt lines.
func TestScrollChunkLooksLikeShellPrompt(t *testing.T) {
	if !scrollChunkLooksLikeShellPrompt([]byte("root@" + testInnerHostFQDN + " [~] ->\n")) {
		t.Fatal("expected TencentOS-style prompt")
	}
	if !scrollChunkLooksLikeShellPrompt([]byte("root@" + testInnerHostShort + " [~] ➜\n")) {
		t.Fatal("expected oh-my-zsh arrow prompt")
	}
	if scrollChunkLooksLikeShellPrompt([]byte("Last login: Thu\n")) {
		t.Fatal("MOTD must not match shell prompt")
	}
	if scrollChunkLooksLikeShellPrompt([]byte("Welcome to TencentOS\n")) {
		t.Fatal("welcome banner must not match shell prompt")
	}
}

// TestNestedPrompt_QuietIdleInjectsWithoutExtraChunk (NP1/W14b): after a prompt chunk and
// wall quiet, tryNestedPromptInject enables chain/host without another PTY chunk.
func TestNestedPrompt_QuietIdleInjectsWithoutExtraChunk(t *testing.T) {
	c := &Client{
		opts:               Options{User: "root", Host: "proxy.example"},
		scroll:             &terminal.Buffer{},
		remotePromptDirAbs: false,
	}
	c.initActivePrompt()
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin

	c.onInteractiveLineSubmitted("ssh root@inner.example")
	_, _ = c.scroll.Write([]byte("Last login: Thu\nroot@inner.example [~] ➜ \n"))
	c.nestedMu.Lock()
	c.nestedMOTDSeen = true
	c.nestedShellReadySeen = true
	c.nestedMOTDAt = time.Now().Add(-shellReadyDelay - time.Second)
	c.nestedMu.Unlock()
	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())

	if !c.tryNestedPromptInject() {
		t.Fatal("expected inject on quiet idle without extra chunk")
	}
	if !c.ChainMode() {
		t.Fatal("expected chain on")
	}
	if c.promptHost() != "inner.example" {
		t.Fatalf("host = %q", c.promptHost())
	}
}

// TestNestedPrompt_ANSIZshPromptReady (NP2/W11b): CSI-wrapped zsh prompts count as ready.
func TestNestedPrompt_ANSIZshPromptReady(t *testing.T) {
	chunk := []byte("\x1b[01;33mroot@" + testInnerHostShort + "\x1b[00m [\x1b[01;36m~\x1b[00m] \x1b[33m➜ \x1b[00m")
	if !scrollChunkLooksLikeShellPrompt(chunk) {
		t.Fatal("ANSI zsh prompt should match shell-ready heuristics")
	}
}

// TestNestedPrompt_MOTDTimeoutFallback (NP3): MOTD without shell-ready still injects after fallback.
func TestNestedPrompt_MOTDTimeoutFallback(t *testing.T) {
	c := &Client{
		opts:               Options{User: "root", Host: "proxy.example"},
		scroll:             &terminal.Buffer{},
		remotePromptDirAbs: false,
	}
	c.initActivePrompt()
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin

	c.onInteractiveLineSubmitted("ssh root@inner.example")
	c.maybeInjectNestedPrompt([]byte("Last login: Thu\nWelcome to TencentOS\n"))
	if c.ChainMode() {
		t.Fatal("must not inject immediately on MOTD alone")
	}

	c.nestedMu.Lock()
	c.nestedMOTDAt = time.Now().Add(-nestedMOTDReadyFallback - time.Second)
	c.nestedMu.Unlock()
	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())

	if !c.tryNestedPromptInject() {
		t.Fatal("expected MOTD timeout fallback inject")
	}
	if !c.ChainMode() {
		t.Fatal("expected chain on after MOTD fallback")
	}
}

// TestNestedPrompt_NativeModeChainWithoutPS1 (NP4/W14c): native prompt mode still enables
// chain/host without writing AiT PS1 setup to stdin.
func TestNestedPrompt_NativeModeChainWithoutPS1(t *testing.T) {
	enableRemoteSystemPromptForTest(t)
	c := &Client{
		opts:               Options{User: "root", Host: "proxy.example"},
		scroll:             &terminal.Buffer{},
		remotePromptDirAbs: false,
	}
	c.initActivePrompt()
	var stdin bytes.Buffer
	c.interactiveStdin = &stdin

	c.onInteractiveLineSubmitted("ssh root@inner.example")
	c.nestedMu.Lock()
	c.nestedMOTDSeen = true
	c.nestedShellReadySeen = true
	c.nestedMOTDAt = time.Now().Add(-shellReadyDelay - time.Second)
	c.nestedMu.Unlock()
	c.lastRemoteOut.Store(time.Now().Add(-5 * time.Second).UnixNano())

	if !c.tryNestedPromptInject() {
		t.Fatal("expected identity/chain update in native mode")
	}
	if !c.ChainMode() {
		t.Fatal("expected chain on in native mode")
	}
	if c.promptHost() != "inner.example" {
		t.Fatalf("host = %q", c.promptHost())
	}
	out := stdin.String()
	if strings.Contains(out, "export PROMPT=") || strings.Contains(out, "AiT root@inner") {
		t.Fatalf("native mode must not inject AiT PS1, stdin = %q", out)
	}
}
