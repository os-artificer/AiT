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

package runner

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
)

// MaxOutputBytes caps the captured output passed back to the model to avoid
// blowing the prompt budget; only the trailing window is retained.
const MaxOutputBytes = 8192

// CapturingRunner runs a command and returns its combined stdout+stderr.
// Implementations should honour ctx cancellation and bound output size.
type CapturingRunner interface {
	RunCapture(ctx context.Context, cmd string) (output string, err error)
}

type localCaptureRunner struct{}

type fnCaptureRunner struct {
	fn func(ctx context.Context, cmd string) (string, error)
}

// DefaultLocalCaptureRunner returns a CapturingRunner that runs cmd via sh -c
// (POSIX) or cmd /c (Windows) and trims output to MaxOutputBytes.
func DefaultLocalCaptureRunner() CapturingRunner {
	return localCaptureRunner{}
}

// RunCapture executes cmd locally and returns the (possibly truncated) combined output.
// Returns the underlying exec.Cmd error when the command exits non-zero or fails to launch.
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
	if len(s) > MaxOutputBytes {
		s = s[len(s)-MaxOutputBytes:]
	}
	return s, err
}

// NewCaptureRunner wraps fn as a CapturingRunner. Useful in tests and remote
// scopes where capture happens by injecting commands into a live PTY.
func NewCaptureRunner(fn func(ctx context.Context, cmd string) (string, error)) CapturingRunner {
	return fnCaptureRunner{fn: fn}
}

// RunCapture forwards to the wrapped function.
func (r fnCaptureRunner) RunCapture(ctx context.Context, cmd string) (string, error) {
	return r.fn(ctx, cmd)
}
