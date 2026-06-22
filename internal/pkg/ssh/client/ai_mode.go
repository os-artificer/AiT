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
	"context"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/ai"
)

// RemoteAIParams is input for entering remote AI mode from !ai-ssh ai / !!.
type RemoteAIParams struct {
	Ctx             context.Context
	Stdin           io.Writer
	InitialQuestion string
	AI              *ai.Client
	HostLabel       string
	RemoteCwd       string
	Scrollback      func() string
	// ReplIn receives stdin bytes while remote AI REPL runs (exclusive with SSH forward).
	ReplIn io.Reader
	// RunCapture runs a command on the remote PTY and returns new scrollback output.
	RunCapture func(cmd string) (string, error)
	// MaxSteps overrides the agent turn limit; 0 means use default.
	MaxSteps int
}

// RemoteAIRunner runs remote-scoped AI mode without client importing aimode.
type RemoteAIRunner func(p RemoteAIParams) error
