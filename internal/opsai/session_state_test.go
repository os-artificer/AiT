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

package opsai

import (
	"bytes"
	"context"
	"io"
	"testing"

	acrunner "github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
)

func TestSessionRuntime_OnMetaConnectWiresDynamic(t *testing.T) {
	rt := newSessionRuntime(nil, func([]string, io.Writer, io.Writer) error { return nil })
	var stdin bytes.Buffer
	rt.attachRemote(nil, sshclient.AgentSessionHandles{
		HostLabel: "u@host.example.com",
		Stdin:     &stdin,
		RunCapture: func(cmd string) (string, error) {
			return "cap:" + cmd, nil
		},
		Scrollback: func() string { return "scroll" },
	})

	snap := rt.dyn.Snapshot()
	if snap.Scope != pkgai.ScopeRemote || snap.HostLabel != "u@host.example.com" {
		t.Fatalf("snap scope=%v host=%q", snap.Scope, snap.HostLabel)
	}
	if snap.Capture == nil || snap.Remote == nil {
		t.Fatal("connect must wire Capture and Remote")
	}
	out, err := snap.Capture.RunCapture(context.Background(), "hostname")
	if err != nil || out != "cap:hostname" {
		t.Fatalf("capture out=%q err=%v", out, err)
	}
	if err := snap.Remote.Run("echo hi"); err != nil {
		t.Fatal(err)
	}
	if stdin.String() != "echo hi\n" {
		t.Fatalf("stdin=%q", stdin.String())
	}
}

// TestSessionRuntime_DisconnectClearsDynamic (OA7) closes remote dynamic state.
func TestSessionRuntime_DisconnectClearsDynamic(t *testing.T) {
	rt := newSessionRuntime(nil, nil)
	rt.attachRemote(nil, sshclient.AgentSessionHandles{
		HostLabel: "u@h",
		RunCapture: func(string) (string, error) {
			return "ok", nil
		},
	})
	if snap := rt.dyn.Snapshot(); snap.Scope != pkgai.ScopeRemote {
		t.Fatal("setup remote")
	}
	if err := rt.disconnect(); err != nil {
		t.Fatal(err)
	}
	snap := rt.dyn.Snapshot()
	if snap.Scope != pkgai.ScopeLocal || snap.HostLabel != "" || snap.Capture != nil || snap.Remote != nil {
		t.Fatalf("after disconnect snap=%+v", snap)
	}
	if rt.client != nil || rt.pty != nil {
		t.Fatal("disconnect must clear client and pty handles")
	}
}

func TestSessionRuntime_OnMetaDisconnectHandled(t *testing.T) {
	rt := newSessionRuntime(nil, nil)
	rt.dyn.SetRemote("u@h", acrunner.NewCaptureRunner(func(context.Context, string) (string, error) {
		return "", nil
	}), nil, nil)
	handled, err := rt.onMeta("disconnect")
	if err != nil || !handled {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	if snap := rt.dyn.Snapshot(); snap.Scope != pkgai.ScopeLocal {
		t.Fatal("disconnect meta should clear dynamic remote")
	}
}
