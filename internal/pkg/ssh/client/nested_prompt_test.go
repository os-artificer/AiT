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
	user, host, ok := parseSSHCommandLine("ssh -p 60528 root@ylgeeker.com", "root")
	if !ok {
		t.Fatal("expected ok")
	}
	if user != "root" || host != "ylgeeker.com" {
		t.Fatalf("got %s@%s", user, host)
	}

	user, host, ok = parseSSHCommandLine("ssh -p60528 root@ylgeeker.com", "root")
	if !ok || user != "root" || host != "ylgeeker.com" {
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

// TestForwardStdin_SSHEnablesChainMode (W14) arms nested await on ssh line submit.
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
}

// TestRunInteractive_SingleHopChainOff (W14) new client defaults chain off.
func TestRunInteractive_SingleHopChainOff(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	if c.ChainMode() {
		t.Fatal("new client chain mode should be off")
	}
}

// TestForwardStdin_SSHTriggersNestedPrompt (W11) injects PS1 after MOTD and quiet period.
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
	c.nestedMu.Lock()
	c.nestedMOTDSeen = true
	c.nestedMu.Unlock()

	c.maybeInjectNestedPrompt([]byte("Last login: Thu\n"))

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
