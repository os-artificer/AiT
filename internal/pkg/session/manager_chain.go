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

package session

import (
	"context"
	"errors"

	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

func (m *Manager) runChainAction(ctx context.Context, chain *sshclient.ChainActionError) error {
	if chain == nil {
		return errors.New("internal: nil chain action")
	}

	switch chain.Op {
	case sshclient.ChainOpLogin:
		return m.runChainLogin(ctx, chain.Args)
	case sshclient.ChainOpLoad:
		return m.runChainLoad(ctx, chain.Args)
	case sshclient.ChainOpSwitch:
		return m.runChainSwitch(ctx, chain.Args)
	default:
		return errors.New("internal: unknown chain action")
	}
}

func (m *Manager) runChainLogin(ctx context.Context, argv []string) error {
	spec, err := openssh.ParseLogin(argv)
	if err != nil {
		return err
	}

	resolved, err := resolve.Resolve(spec)
	if err != nil {
		return err
	}

	return m.Login(ctx, LoginOptsFromResolved(resolved))
}

func (m *Manager) runChainLoad(ctx context.Context, argv []string) error {
	target, password, useTmux, needPick, err := ParseChainLoadArgs(argv)
	if err != nil {
		return err
	}

	if needPick {
		target, err = m.pickSessionForChain()
		if err != nil {
			return err
		}
	}

	return m.Load(ctx, target, useTmux, password)
}

func (m *Manager) runChainSwitch(ctx context.Context, argv []string) error {
	target, password, useTmux, needPick, err := ParseChainSwitchArgs(argv)
	if err != nil {
		return err
	}

	if needPick {
		target, err = m.pickSessionForChain()
		if err != nil {
			return err
		}
	}

	return m.switchFromChain(ctx, target, useTmux, password)
}

func (m *Manager) switchFromChain(ctx context.Context, target string, useTmux bool, password string) error {
	tgt, err := m.Store.GetSession(target)
	if err != nil {
		return err
	}

	current, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if current == tgt.ID {
		return m.Load(ctx, tgt.ID, useTmux || tgt.TmuxEnabled, password)
	}

	return m.Switch(ctx, target, useTmux, password, false)
}

func (m *Manager) pickSessionForChain() (string, error) {
	return m.PickSessionInteractiveTTY()
}
