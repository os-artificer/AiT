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
	"os/exec"
	"runtime"
	"strings"
)

// CapturingRunner runs a command and returns combined output.
type CapturingRunner interface {
	RunCapture(ctx context.Context, cmd string) (output string, err error)
}

type localCaptureRunner struct{}

// DefaultLocalCaptureRunner returns a CapturingRunner for local shell commands.
func DefaultLocalCaptureRunner() CapturingRunner {
	return localCaptureRunner{}
}

// RunCapture implements CapturingRunner using sh -c or cmd /c.
func (localCaptureRunner) RunCapture(ctx context.Context, cmd string) (string, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "", nil
	}

	var c *exec.Cmd
	if runtime.GOOS == "windows" {
		c = exec.CommandContext(ctx, "cmd", "/c", cmd)
	} else {
		c = exec.CommandContext(ctx, "sh", "-c", cmd)
	}
	out, err := c.CombinedOutput()
	s := string(out)
	if len(s) > agentMaxOutputBytes {
		s = s[len(s)-agentMaxOutputBytes:]
	}
	return s, err
}

type fnCaptureRunner struct {
	fn func(ctx context.Context, cmd string) (string, error)
}

// NewCaptureRunner wraps fn as a CapturingRunner.
func NewCaptureRunner(fn func(ctx context.Context, cmd string) (string, error)) CapturingRunner {
	return fnCaptureRunner{fn: fn}
}

func (r fnCaptureRunner) RunCapture(ctx context.Context, cmd string) (string, error) {
	return r.fn(ctx, cmd)
}
