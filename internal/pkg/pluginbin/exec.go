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

// Package pluginbin runs already-resolved plugin executables for the ait REPL.
package pluginbin

import (
	"context"
	"io"
	"sync"

	"github.com/os-artificer/AiT/internal/pkg/shellext"
)

// Runner executes path with argv (argv[0] is the display name; argv[1:] are arguments).
type Runner func(ctx context.Context, path string, argv []string) error

var (
	runnerMu sync.Mutex
	runner   Runner
)

// SetRunner registers an optional Runner used by ExecPath (TTY Restore wrapper).
// Pass nil to restore the default shellext.Run path.
// Returns the previously registered Runner (may be nil).
func SetRunner(fn Runner) Runner {
	runnerMu.Lock()
	defer runnerMu.Unlock()
	prev := runner
	runner = fn
	return prev
}

// ExecPath runs the executable at path with argv via the injected Runner or shellext.Run.
// path must already be resolved by discovery; this function never calls LookPath.
// When a Runner is set it receives path/argv only (TTY wrappers typically wire os.Stdin/Stdout);
// in/out/errOut are used by the default shellext.Run path.
func ExecPath(ctx context.Context, path string, argv []string, in io.Reader, out, errOut io.Writer) error {
	runnerMu.Lock()
	fn := runner
	runnerMu.Unlock()

	if fn != nil {
		return fn(ctx, path, argv)
	}
	return shellext.Run(ctx, path, argv, in, out, errOut)
}
