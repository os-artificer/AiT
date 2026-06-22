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

package ai

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/os-artificer/AiT/internal/ai-ssh/ai/mode"
	"github.com/os-artificer/AiT/internal/ai-ssh/ai/mode/agent"
	"github.com/os-artificer/AiT/internal/ai-ssh/ai/runner"
	"github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
)

// ptyRemoteRunner executes commands by writing them to the remote PTY stdin.
type ptyRemoteRunner struct {
	stdin io.Writer
}

// Run writes cmd followed by a newline to the remote PTY stdin.
// Returns an error when the remote stdin is not available.
func (r *ptyRemoteRunner) Run(cmd string) error {
	if r.stdin == nil {
		return fmt.Errorf("remote stdin not available")
	}
	_, err := io.WriteString(r.stdin, cmd+"\n")
	return err
}

func localAIClient(mgr *session.Manager, modelOverride string) (*ai.Client, error) {
	return runner.LocalAIClient(mgr, modelOverride)
}

func localAIModeOptions(aiClient *ai.Client, maxSteps int, exec CLIExecutor) mode.Options {
	return runner.LocalAIModeOptions(aiClient, maxSteps, exec)
}

func runRemoteAIMode(p sshclient.RemoteAIParams, exec CLIExecutor) error {
	if p.AI == nil {
		return ai.NotConfiguredErr()
	}

	var in io.Reader = os.Stdin
	useTTY := false
	if p.ReplIn != nil {
		in = p.ReplIn
		useTTY = true
	}

	opts := mode.Options{
		Scope:          ai.ScopeRemote,
		AI:             p.AI,
		HostLabel:      p.HostLabel,
		RemoteCwd:      p.RemoteCwd,
		Scrollback:     p.Scrollback,
		Remote:         &ptyRemoteRunner{stdin: p.Stdin},
		CLI:            func(argv []string, out, errOut io.Writer) error { return exec(argv, out, errOut) },
		In:             in,
		Out:            os.Stdout,
		ErrOut:         os.Stderr,
		UseTTYLineEdit: useTTY,
		MaxSteps:       p.MaxSteps,
	}
	if p.RunCapture != nil {
		opts.Capture = agent.NewCaptureRunner(func(_ context.Context, cmd string) (string, error) {
			return p.RunCapture(cmd)
		})
	}
	return mode.RunREPL(p.Ctx, opts, p.InitialQuestion)
}
