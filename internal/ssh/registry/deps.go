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

package registry

import (
	"context"
	"io"

	"github.com/os-artificer/AiT/internal/pkg/session"
)

// RunWithMgr opens a session.Manager and runs fn under a signal-cancellable context.
type RunWithMgr func(fn func(ctx context.Context, mgr *session.Manager) error) error

// ExecuteWriters runs the Cobra root with custom stdout/stderr (e.g. aissh shell REPL).
type ExecuteWriters func(args []string, out, errOut io.Writer) error

// Deps groups dependencies injected when attaching registered top-level commands.
type Deps struct {
	RunWithMgr       RunWithMgr
	ExecuteWriters   ExecuteWriters
	ExecuteCobraOnly ExecuteWriters
}
