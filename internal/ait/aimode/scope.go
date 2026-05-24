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

package aimode

import (
	"context"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// ExecPolicy controls how often run/ait commands require user confirmation.
type ExecPolicy int

const (
	// ExecPolicyPerCmd asks before each run (default).
	ExecPolicyPerCmd ExecPolicy = iota
	// ExecPolicySessionOnce asks once per AI REPL session, then trusts until exit.
	ExecPolicySessionOnce
)

// LocalRunner runs a shell command on the local machine.
type LocalRunner interface {
	Run(ctx context.Context, cmd string) error
}

// RemoteRunner sends a command to the active SSH session (PTY stdin).
type RemoteRunner interface {
	Run(cmd string) error
}

// Options configures an AI mode REPL. CLI and client inject runners; aimode does not import cli or client.
type Options struct {
	Scope       ai.Scope
	AI          *ai.Client
	ExecPolicy  ExecPolicy
	HostLabel   string
	RemoteCwd   string
	Scrollback  func() string
	Local       LocalRunner
	Remote      RemoteRunner
	Capture     CapturingRunner
	CLI         func([]string, io.Writer, io.Writer) error
	In          io.Reader
	Out         io.Writer
	ErrOut      io.Writer
	ConfirmRead func(prompt string) (ConfirmResult, error)
	// UseTTYLineEdit enables runAITTY when In is not os.Stdin (SSH remote AI ReplIn).
	UseTTYLineEdit bool
}
