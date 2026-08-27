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
	"fmt"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/procsig"
)

type captureResult struct {
	out string
	err error
}

// runToolCapture executes cmd for one agent step with the terminal handed to the
// child process and a wall-clock bound applied.
// Returns the captured output and the runner error; the error is an observation
// for the model unless ctx itself was cancelled.
func runToolCapture(ctx context.Context, cfg Config, cmd string) (string, error) {
	var out string
	err := withTTYHandoff(cfg, func() error {
		var runErr error
		out, runErr = runBoundedCapture(ctx, cfg, cmd)
		return runErr
	})
	return out, err
}

// runBoundedCapture runs cmd through the configured CapturingRunner under a
// derived timeout, and additionally bounds the wait so a runner that ignores ctx
// (for example a remote PTY capture) cannot block the loop forever.
// Returns ctx.Err() when the caller cancelled, or a timeout error when the bound
// elapsed first.
func runBoundedCapture(ctx context.Context, cfg Config, cmd string) (string, error) {
	timeout := cfg.AgentCommandTimeout()
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Buffered so an abandoned runner goroutine can always finish its send.
	done := make(chan captureResult, 1)
	go func() {
		out, err := captureRun(runCtx, cfg, cmd)
		done <- captureResult{out: out, err: err}
	}()

	select {
	case res := <-done:
		return res.out, res.err
	case <-runCtx.Done():
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", commandTimeoutErr(timeout)
	}
}

// withTTYHandoff runs fn with the loop's raw terminal mode released (when the
// caller provided SuspendTTY) and parent Interrupt absorbed, so Ctrl+C
// terminates the child process instead of ending the agent session.
// Returns the error returned by fn.
func withTTYHandoff(cfg Config, fn func() error) error {
	if cfg.SuspendTTY != nil {
		if resume := cfg.SuspendTTY(); resume != nil {
			defer resume()
		}
	}
	return procsig.WithAbsorbedInterrupt(fn)
}

func commandTimeoutErr(timeout time.Duration) error {
	return fmt.Errorf("timed out after %s (agent command timeout)", timeout)
}
