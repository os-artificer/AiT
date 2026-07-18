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

package connect

import (
	"context"
	"errors"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"

	"golang.org/x/crypto/ssh"
)

// exitError wraps an exit code for main.
type exitError int

// Error implements error; the message is empty because ExitCode carries the code.
func (e exitError) Error() string {
	return ""
}

// Run parses argv, resolves OpenSSH config, and runs an interactive or one-shot login via mgr.
func Run(ctx context.Context, mgr *session.Manager, argv []string) error {
	spec, err := Parse(argv)
	if err != nil {
		return err
	}

	resolved, err := resolve.Resolve(spec)
	if err != nil {
		return err
	}

	return mgr.Login(ctx, OptsFromResolved(resolved))
}

// ExitCode returns the process exit code if err is from a remote command.
// ok is true when err wraps exitError or client.RemoteExitError.
func ExitCode(err error) (int, bool) {
	var ee exitError
	if errors.As(err, &ee) {
		return int(ee), true
	}

	var re *client.RemoteExitError
	if errors.As(err, &re) {
		return re.Code, true
	}

	var exitErr *ssh.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitStatus(), true
	}
	return 0, false
}
