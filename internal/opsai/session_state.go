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
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/os-artificer/AiT/internal/pkg/agentcore"
	acrunner "github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// sessionRuntime holds optional live SSH client state for connect/disconnect (OA5–OA7).
type sessionRuntime struct {
	mu   sync.Mutex
	cfg  *config.Config
	exec CLIExecutor

	client *sshclient.Client
	pty    *sshclient.AgentPTY
	dyn    *agentcore.DynamicRemote
}

func newSessionRuntime(cfg *config.Config, exec CLIExecutor) *sessionRuntime {
	return &sessionRuntime{
		cfg:  cfg,
		exec: exec,
		dyn: &agentcore.DynamicRemote{
			Scope: pkgai.ScopeLocal,
		},
	}
}

func (s *sessionRuntime) onMeta(line string) (bool, error) {
	fields := strings.Fields(strings.TrimSpace(line))
	if len(fields) == 0 {
		return false, nil
	}
	switch strings.ToLower(fields[0]) {
	case cmdConnect:
		return true, s.connectMeta(fields[1:])
	case "disconnect":
		return true, s.disconnect()
	default:
		return false, nil
	}
}

func (s *sessionRuntime) connectMeta(argv []string) error {
	if len(argv) == 0 {
		return errhint.WithTryMsg("connect requires a destination",
			"Pass an SSH destination.",
			"connect user@host",
			toolname.AICLI("connect user@host"),
		)
	}

	spec, err := openssh.ParseLogin(argv)
	if err != nil {
		return err
	}
	if spec.Tmux || len(spec.RemoteCommand) > 0 {
		return errhint.WithTryMsg("unsupported connect options",
			"ai connect does not support --tmux or a remote command.",
			"connect user@host",
		)
	}

	resolved, err := resolve.Resolve(spec)
	if err != nil {
		return err
	}

	client, err := sshclient.DialChain(resolved.ClientOptions())
	if err != nil {
		return err
	}

	pty, err := client.StartAgentPTY(resolved.User + "@" + resolved.Host)
	if err != nil {
		_ = client.Close()
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeLocked()
	s.client = client
	s.pty = pty
	s.attachHandlesLocked(pty.Handles)
	return nil
}

func (s *sessionRuntime) attachHandlesLocked(h sshclient.AgentSessionHandles) {
	capture := acrunner.NewCaptureRunner(func(_ context.Context, cmd string) (string, error) {
		return h.RunCapture(cmd)
	})
	var remote agentcore.RemoteRunner
	if h.Stdin != nil {
		remote = &ptyRemoteRunner{stdin: h.Stdin}
	}
	s.dyn.SetRemote(h.HostLabel, capture, h.Scrollback, remote)
}

func (s *sessionRuntime) disconnect() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeLocked()
	return nil
}

func (s *sessionRuntime) closeLocked() {
	if s.pty != nil {
		_ = s.pty.Close()
		s.pty = nil
	}
	if s.client != nil {
		_ = s.client.Close()
		s.client = nil
	}
	s.dyn.SetLocal()
}

func (s *sessionRuntime) close() {
	_ = s.disconnect()
}

// attachRemote wires Capture after a successful AgentSession (tests).
func (s *sessionRuntime) attachRemote(client *sshclient.Client, h sshclient.AgentSessionHandles) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeLocked()
	s.client = client
	s.attachHandlesLocked(h)
}

// ptyRemoteRunner executes commands by writing them to the remote PTY stdin.
type ptyRemoteRunner struct {
	stdin io.Writer
}

// Run writes cmd followed by a newline to the remote PTY stdin.
func (r *ptyRemoteRunner) Run(cmd string) error {
	if r.stdin == nil {
		return fmt.Errorf("remote stdin not available")
	}
	_, err := io.WriteString(r.stdin, cmd+"\n")
	return err
}
