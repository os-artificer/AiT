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

package shellext

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// Run executes path with argv[1:] as arguments (argv[0] is the command name for display).
// Stdin, stdout, and stderr are wired to in, out, and errOut. AIT_HOME is set when AppRoot resolves.
func Run(ctx context.Context, path string, argv []string, in io.Reader, out, errOut io.Writer) error {
	args := argv[1:]
	if len(args) == 0 {
		args = nil
	}

	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = errOut
	cmd.Env = appendEnvAITHome(os.Environ())
	return cmd.Run()
}

func appendEnvAITHome(env []string) []string {
	root, err := platform.AppRoot()
	if err != nil || root == "" {
		return env
	}
	return append(env, "AIT_HOME="+root)
}
