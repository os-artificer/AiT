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

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/ssh/meta"
)

// NewManager opens the store, loads config, and wires metaCLI.
// metaCLI runs aissh subcommands from in-session !ait lines (provided by internal/ssh/meta).
func NewManager(metaCLI func([]string) error) (*session.Manager, error) {
	st, err := store.Open()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		st.Close()
		return nil, err
	}

	mgr := &session.Manager{Store: st, Config: cfg}
	mgr.MetaCLI = metaCLI
	return mgr, nil
}

func newManager() (*session.Manager, error) {
	return NewManager(meta.Execute)
}

func runWithManager(fn func(ctx context.Context, mgr *session.Manager) error) error {
	mgr, err := newManager()
	if err != nil {
		return err
	}
	defer mgr.Store.Close()

	ctx, cancel := CommandContext()
	defer cancel()
	return fn(ctx, mgr)
}
