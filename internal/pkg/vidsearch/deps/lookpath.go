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

package deps

import (
	"context"
	"os/exec"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

// LookPathFunc locates an executable on PATH.
type LookPathFunc func(name string) (string, error)

// DefaultLookPath uses platform.LookPath.
func DefaultLookPath(name string) (string, error) {
	return platform.LookPath(name)
}

// CommandRunner runs external install commands (injectable for tests).
type CommandRunner interface {
	LookPath(name string) (string, error)
	Run(ctx context.Context, name string, args []string) error
}

type osRunner struct{}

func (osRunner) LookPath(name string) (string, error) {
	return platform.LookPath(name)
}

func (osRunner) Run(ctx context.Context, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// DefaultRunner returns a production CommandRunner.
func DefaultRunner() CommandRunner {
	return osRunner{}
}
