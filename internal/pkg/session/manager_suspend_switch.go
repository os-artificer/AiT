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
	"fmt"
	"path/filepath"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func (m *Manager) suspend(id string, scroll *terminal.Buffer, cwd string, cols, rows int) error {
	dir, err := platform.ScrollbackDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, id+".log")
	if err := scroll.SaveToFile(path); err != nil {
		return err
	}

	st := &store.SessionState{
		SessionID: id, Cols: cols, Rows: rows,
		RemoteCwd: cwd, ScrollbackPath: terminal.ScrollbackFileName(id), SuspendedAt: time.Now().Unix(),
	}
	if err := m.Store.SaveSessionState(st); err != nil {
		return err
	}

	return m.Store.UpdateStatus(id, store.StatusSuspended)
}

// SuspendCurrent prints guidance when suspend is requested from the local AiT shell (not inside SSH).
// Returns fmt.Errorf when no current session is recorded in store meta.
func (m *Manager) SuspendCurrent() error {
	id, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if id == "" {
		return errhint.WithTryMsg("no current session",
			"Load or switch to a session first, or suspend from inside SSH.",
			tryLoadFromLocal, trySuspendInSSH, trySessionHelp,
		)
	}
	fmt.Println("Suspend from the local shell is not supported.")
	fmt.Println("Reconnect to the session and run !ait suspend inside SSH.")
	return nil
}

// Switch moves the active session to target (id or alias), optionally forcing past an still-active current session.
// ctx, useTmux, and password behave like Load. force skips the active-session guard and marks the current row suspended.
// Returns store lookup errors, fmt.Errorf when the current session is active and force is false, or Load errors.
func (m *Manager) Switch(ctx context.Context, target string, useTmux bool, password string, force bool) error {
	tgt, err := m.Store.GetSession(target)
	if err != nil {
		return err
	}

	current, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if current == tgt.ID {
		fmt.Println("already on session", tgt.AliasOrID())
		return nil
	}

	if err := m.reconcileCurrentOnSwitch(current, force); err != nil {
		return err
	}

	return m.Load(ctx, tgt.ID, useTmux || tgt.TmuxEnabled, password)
}
