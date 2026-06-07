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
	_ "embed"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

//go:embed client.go
var clientGoSource string

// TestInjectInitialCommands_LoadCdSpacePrefixed (W20) verifies load cwd restore uses Inject space prefix.
func TestInjectInitialCommands_LoadCdSpacePrefixed(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	go func() {
		time.Sleep(20 * time.Millisecond)
		c.touchRemoteOut()
	}()

	setup := RemotePromptInitialCommands("u", "h", false)
	cmds := append(setup, "cd '/data/work'")
	if err := c.injectInitialCommands(stdin, cmds); err != nil {
		t.Fatal(err)
	}

	body := stdin.String()
	parts := strings.Split(strings.TrimSpace(body), "\n")
	if len(parts) != 2 {
		t.Fatalf("want 2 injected lines, got %d: %q", len(parts), body)
	}
	if !strings.Contains(parts[1], injectSpacePrefix+"cd '/data/work'") {
		t.Fatalf("cd line not space-prefixed via Inject: %q", parts[1])
	}
}

// TestCaptureRemoteCwd_UsesInjectCapture (W21) verifies suspend cwd probe uses probeShellPwd, not raw stdin.
func TestCaptureRemoteCwd_UsesInjectCapture(t *testing.T) {
	c := testClientWithPrompt("u", "h")
	c.scroll = &terminal.Buffer{}

	var probeCalled bool
	orig := probeShellPwdHook
	defer func() { probeShellPwdHook = orig }()
	probeShellPwdHook = func(_ *Client, _ io.Writer) (string, error) {
		probeCalled = true
		return "/data/work", nil
	}

	got := c.captureRemoteCwd(&bytes.Buffer{}, InteractiveOpts{})
	if !probeCalled {
		t.Fatal("captureRemoteCwd should delegate to probeShellPwd when tmux is off")
	}
	if got != "/data/work" {
		t.Fatalf("cwd = %q, want /data/work", got)
	}
}

// TestNoRawPTYCommandInject guards against new bare stdin command writes in client.go.
func TestNoRawPTYCommandInject(t *testing.T) {
	const allowedRefresh = `stdin.Write([]byte("\r"))`

	for _, line := range strings.Split(clientGoSource, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, "stdin.Write") {
			continue
		}
		if trimmed == allowedRefresh {
			continue
		}
		t.Fatalf("client.go must not raw-write PTY commands; use Inject: %s", trimmed)
	}
}
