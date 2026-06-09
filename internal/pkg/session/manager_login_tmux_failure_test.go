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
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/store"
)

// TestCleanupOnTmuxFailure_NewRowDeletes (TM14d) removes a just-created row on tmux.Ensure failure.
func TestCleanupOnTmuxFailure_NewRowDeletes(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Tmux: true, Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	if len(list) != 1 {
		t.Fatalf("sessions=%d want 1", len(list))
	}
	id := list[0].ID
	_ = m.Store.SetMeta(store.MetaCurrentSession, id)

	m.cleanupOnTmuxFailure(true, true, id)

	list, _ = m.Store.ListSessions()
	if len(list) != 0 {
		t.Fatalf("new row should be deleted, sessions=%d", len(list))
	}
	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur != "" {
		t.Fatalf("current_session_id=%q want empty", cur)
	}
}

// TestCleanupOnTmuxFailure_ExistingRowSuspends (TM14d) marks an existing row suspended on failure.
func TestCleanupOnTmuxFailure_ExistingRowSuspends(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Tmux: true, Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	id := list[0].ID

	m.cleanupOnTmuxFailure(true, false, id)

	sess, err := m.Store.GetSession(id)
	if err != nil {
		t.Fatal(err)
	}
	if sess.Status != store.StatusSuspended {
		t.Fatalf("status=%q want suspended", sess.Status)
	}
}

// TestDeleteLoginSession_RemovesRow (TM14c) deletes session row and clears current meta.
func TestDeleteLoginSession_RemovesRow(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Tmux: true, Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	id := list[0].ID
	_ = m.Store.SetMeta(store.MetaCurrentSession, id)

	m.deleteLoginSession(id)

	if _, err := m.Store.GetSession(id); err == nil {
		t.Fatal("expected session deleted")
	}
	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur != "" {
		t.Fatalf("current_session_id=%q want empty", cur)
	}
}

// TestLoginInteractiveFailure_NewRowDeletes (TM14c) simulates post-runInteractive cleanup for new --tmux login.
func TestLoginInteractiveFailure_NewRowDeletes(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Tmux: true, Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	id := list[0].ID
	_ = m.Store.SetMeta(store.MetaCurrentSession, id)

	m.deleteLoginSession(id)

	list, _ = m.Store.ListSessions()
	if len(list) != 0 {
		t.Fatalf("phantom row should be gone, sessions=%d", len(list))
	}
}
