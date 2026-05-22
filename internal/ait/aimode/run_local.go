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
	"fmt"
	"os/exec"
	"runtime"
)

type localRunner struct{}

// DefaultLocalRunner returns a LocalRunner that runs commands via sh -c or cmd /c.
func DefaultLocalRunner() LocalRunner {
	return localRunner{}
}

// Run executes cmd on the local machine using the platform shell wrapper.
func (localRunner) Run(ctx context.Context, cmd string) error {
	if cmd == "" {
		return fmt.Errorf("empty command")
	}

	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", cmd).Run()
	}
	return exec.CommandContext(ctx, "sh", "-c", cmd).Run()
}
