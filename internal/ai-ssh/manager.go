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

package aish

import (
	"context"

	"github.com/os-artificer/AiT/internal/ai-ssh/ai"
	"github.com/os-artificer/AiT/internal/ai-ssh/meta"
	pkgai "github.com/os-artificer/AiT/internal/pkg/ai"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/session"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

// NewManager opens the store, loads config, optionally constructs an AI client, and wires metaCLI.
// metaCLI runs ai-ssh subcommands from in-session !ai-ssh lines (provided by internal/ai-ssh/meta).
func NewManager(metaCLI func([]string) error, onRemoteAI sshclient.RemoteAIRunner) (*session.Manager, error) {
	st, err := store.Open()
	if err != nil {
		return nil, err
	}

	cfg, err := config.Load()
	if err != nil {
		st.Close()
		return nil, err
	}

	var aiClient *pkgai.Client
	if c, err := pkgai.NewClient(cfg); err == nil {
		aiClient = c
	}

	mgr := &session.Manager{Store: st, Config: cfg, AI: aiClient, OnRemoteAI: onRemoteAI}
	mgr.MetaCLI = metaCLI
	return mgr, nil
}

func newManager() (*session.Manager, error) {
	return NewManager(meta.Execute, ai.RemoteRunner(ExecuteWriters))
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
