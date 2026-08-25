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
	"os/exec"
	"runtime"
	"strings"
)

type defaultLocalRunner struct{}

func (defaultLocalRunner) Run(ctx context.Context, cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", "/c", cmd).Run()
	}
	return exec.CommandContext(ctx, "sh", "-c", cmd).Run()
}
