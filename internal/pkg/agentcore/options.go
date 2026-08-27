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

package agentcore

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/agentcore/approve"
	"github.com/os-artificer/AiT/internal/pkg/agentcore/runner"
	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// LocalRunner runs a shell command on the local machine.
type LocalRunner interface {
	Run(ctx context.Context, cmd string) error
}

// RemoteRunner sends a command on the host session channel (e.g. remote PTY stdin).
type RemoteRunner interface {
	Run(cmd string) error
}

// Config configures an agent session loop. Product packages inject runners; agentcore does not import cli or ssh.
type Config struct {
	Scope       ai.Scope
	AI          *ai.Client
	HostLabel   string
	RemoteCwd   string
	Scrollback  func() string
	Local       LocalRunner
	Remote      RemoteRunner
	Capture     runner.CapturingRunner
	CLI         func([]string, io.Writer, io.Writer) error
	Toolbox     func([]string, io.Writer, io.Writer) error
	In          io.Reader
	Out         io.Writer
	ErrOut      io.Writer
	ConfirmRead func(prompt string) (approve.Result, error)
	MaxSteps    int
	// CommandTimeout bounds one tool execution; defaultCommandTimeout when unset.
	CommandTimeout time.Duration
	// SuspendTTY releases the loop's raw terminal mode for the duration of a child
	// process and returns the resume function, or nil when the handoff failed.
	// Set by the TTY loop so Ctrl+C reaches the child instead of being read as a byte.
	SuspendTTY func() (resume func())
	Domain     string
	Emitter    Emitter
	Dynamic    *DynamicRemote
	// EnvExtra is appended to AssistEnv context on each turn (e.g. ops snapshot).
	EnvExtra string
	// StreamDelta receives incremental assistant text when providers stream (Phase 3).
	StreamDelta func(delta string)
	// Nested when true treats exit/back as loop end without process exit (ait> agent).
	Nested bool
}

// DynamicRemote holds live remote session fields for optional hot connect/disconnect.
// The shipped ai connect path sets static Config fields and does not wire Dynamic.
type DynamicRemote struct {
	mu sync.RWMutex

	Scope      ai.Scope
	HostLabel  string
	RemoteCwd  string
	Capture    runner.CapturingRunner
	Scrollback func() string
	Remote     RemoteRunner
}

// DynamicSnapshot is a consistent copy of DynamicRemote fields for one turn.
type DynamicSnapshot struct {
	Scope      ai.Scope
	HostLabel  string
	RemoteCwd  string
	Capture    runner.CapturingRunner
	Scrollback func() string
	Remote     RemoteRunner
}

// Snapshot returns a consistent copy of live remote fields.
func (d *DynamicRemote) Snapshot() DynamicSnapshot {
	if d == nil {
		return DynamicSnapshot{}
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	return DynamicSnapshot{
		Scope:      d.Scope,
		HostLabel:  d.HostLabel,
		RemoteCwd:  d.RemoteCwd,
		Capture:    d.Capture,
		Scrollback: d.Scrollback,
		Remote:     d.Remote,
	}
}

// SetLocal clears remote session fields (disconnect).
func (d *DynamicRemote) SetLocal() {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Scope = ai.ScopeLocal
	d.HostLabel = ""
	d.RemoteCwd = ""
	d.Capture = nil
	d.Scrollback = nil
	d.Remote = nil
}

// SetRemote updates live fields after a successful connect.
func (d *DynamicRemote) SetRemote(hostLabel string, capture runner.CapturingRunner, scrollback func() string, remote RemoteRunner) {
	if d == nil {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Scope = ai.ScopeRemote
	d.HostLabel = hostLabel
	d.Capture = capture
	d.Scrollback = scrollback
	d.Remote = remote
}

// MaxAgentSteps returns cfg.MaxSteps if set, otherwise defaultMaxSteps.
func (cfg Config) MaxAgentSteps() int {
	if cfg.MaxSteps > 0 {
		return cfg.MaxSteps
	}
	return defaultMaxSteps
}

// AgentCommandTimeout returns cfg.CommandTimeout if set, otherwise defaultCommandTimeout.
func (cfg Config) AgentCommandTimeout() time.Duration {
	if cfg.CommandTimeout > 0 {
		return cfg.CommandTimeout
	}
	return defaultCommandTimeout
}
