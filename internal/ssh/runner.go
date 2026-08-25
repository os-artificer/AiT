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

package aissh

import (
	"context"

	"github.com/os-artificer/AiT/internal/pkg/session"
)

// Runner opens a Manager, runs fn with a signal-cancellable context, and closes the store.
type Runner struct {
	MetaCLI func([]string) error
}

// Run opens a Manager, executes fn with a signal-cancellable context, and closes the store.
// Returns any error from NewManager or fn.
func (r Runner) Run(fn func(ctx context.Context, mgr *session.Manager) error) error {
	mgr, err := NewManager(r.MetaCLI)
	if err != nil {
		return err
	}
	defer mgr.Store.Close()

	ctx, cancel := CommandContext()
	defer cancel()
	return fn(ctx, mgr)
}
