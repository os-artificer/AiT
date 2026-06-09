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
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/store"

	"github.com/google/uuid"
)

func persistLoginForTest(t *testing.T, m *Manager, opts LoginOpts) {
	t.Helper()
	plan, err := m.planLoginPersistence(opts, uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}
	if err := m.persistLoginSession(nil, opts, plan); err != nil {
		t.Fatal(err)
	}
}

func TestShouldPersistLogin(t *testing.T) {
	tests := []struct {
		name string
		opts LoginOpts
		want bool
	}{
		{"none", LoginOpts{Host: "h"}, false},
		{"save only", LoginOpts{Host: "h", Save: true}, true},
		{"alias only", LoginOpts{Host: "h", Alias: "prod"}, true},
		{"both", LoginOpts{Host: "h", Save: true, Alias: "prod"}, true},
		{"tmux only", LoginOpts{Host: "h", Tmux: true}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldPersistLogin(tc.opts); got != tc.want {
				t.Fatalf("shouldPersistLogin(%+v) = %v, want %v", tc.opts, got, tc.want)
			}
		})
	}
}

func TestPersistLoginSessionSaveOnly(t *testing.T) {
	m := newTestManager(t)
	opts := LoginOpts{Save: true, Host: "example.com", Port: 22, User: "u"}
	persistLoginForTest(t, m, opts)

	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("sessions = %d, want 1", len(list))
	}
	if list[0].Alias != "" {
		t.Fatalf("alias = %q, want empty", list[0].Alias)
	}
	if _, err := m.Store.GetSession(list[0].ID); err != nil {
		t.Fatalf("GetSession by id: %v", err)
	}

	firstID := list[0].ID
	firstExpiry := list[0].AuthExpiresAt
	time.Sleep(1100 * time.Millisecond)

	persistLoginForTest(t, m, opts)

	list, err = m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("after second save sessions = %d, want 1", len(list))
	}
	if list[0].ID != firstID {
		t.Fatalf("id=%q want %q", list[0].ID, firstID)
	}
	if list[0].AuthExpiresAt <= firstExpiry {
		t.Fatalf("auth_expires_at=%d want > %d", list[0].AuthExpiresAt, firstExpiry)
	}
}

func TestPersistLoginSessionAliasOnly(t *testing.T) {
	m := newTestManager(t)
	opts := LoginOpts{Alias: "prod", Host: "example.com", Port: 22, User: "u"}
	persistLoginForTest(t, m, opts)

	sess, err := m.Store.GetSession("prod")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Alias != "prod" {
		t.Fatalf("alias = %q", sess.Alias)
	}
}

func TestPersistLoginSessionSaveAndAlias(t *testing.T) {
	m := newTestManager(t)
	opts := LoginOpts{Save: true, Alias: "prod", Host: "h", Port: 22, User: "u"}
	persistLoginForTest(t, m, opts)

	sess, err := m.Store.GetSession("prod")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Host != "h" {
		t.Fatalf("host = %q", sess.Host)
	}
}

func TestPersistLoginSessionSetsMeta(t *testing.T) {
	m := newTestManager(t)
	opts := LoginOpts{Save: true, Host: "h", Port: 22, User: "u", Password: "secret"}
	persistLoginForTest(t, m, opts)

	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("sessions = %d, want 1", len(list))
	}

	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur != list[0].ID {
		t.Fatalf("meta current = %q, want %q", cur, list[0].ID)
	}

	p, err := creds.GetSessionUnlock(list[0].ID)
	if err != nil || p != "secret" {
		t.Fatalf("keyring password: %q err=%v", p, err)
	}
}

func TestPersistInteractiveLoginSetsMeta(t *testing.T) {
	m := newTestManager(t)
	sess := &store.Session{
		ID: "sess-1", Host: "h", Port: 22, Username: "u",
		AuthExpiresAt: time.Now().Add(time.Hour).Unix(),
		Status:        store.StatusSuspended,
	}
	if err := m.persistInteractiveLogin(sess, LoginOpts{Password: "secret"}); err != nil {
		t.Fatal(err)
	}

	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur != "sess-1" {
		t.Fatalf("meta current = %q", cur)
	}

	p, err := creds.GetSessionUnlock("sess-1")
	if err != nil || p != "secret" {
		t.Fatalf("keyring password: %q err=%v", p, err)
	}
}

func TestResolveLoginSessionIDReusesAlias(t *testing.T) {
	m := newTestManager(t)
	existing := &store.Session{
		ID: "fixed-id", Alias: "prod", Host: "old", Port: 22, Username: "u",
		AuthExpiresAt: time.Now().Add(time.Hour).Unix(), Status: store.StatusSuspended,
	}
	if err := m.Store.UpsertSession(existing); err != nil {
		t.Fatal(err)
	}

	id := m.resolveLoginSessionID(LoginOpts{Alias: "prod", Host: "old", Port: 22, User: "u"}, "new-id")
	if id != "fixed-id" {
		t.Fatalf("id = %q, want fixed-id", id)
	}
}

func TestResolveLoginSessionIDClearsScrollbackOnHostChange(t *testing.T) {
	m := newTestManager(t)
	existing := &store.Session{
		ID: "fixed-id", Alias: "prod", Host: "old", Port: 22, Username: "u",
		AuthExpiresAt: time.Now().Add(time.Hour).Unix(), Status: store.StatusSuspended,
	}
	if err := m.Store.UpsertSession(existing); err != nil {
		t.Fatal(err)
	}
	st := &store.SessionState{SessionID: "fixed-id", ScrollbackPath: "fixed-id.log"}
	if err := m.Store.SaveSessionState(st); err != nil {
		t.Fatal(err)
	}

	_ = m.resolveLoginSessionID(LoginOpts{Alias: "prod", Host: "new", Port: 22, User: "u"}, "new-id")
	st, err := m.Store.GetSessionState("fixed-id")
	if err != nil {
		t.Fatal(err)
	}
	if st != nil {
		t.Fatal("expected session state cleared after host change")
	}
}

func TestEphemeralSuspendErrorText(t *testing.T) {
	if !strings.Contains(errEphemeralSuspendRequiresPersist, "--save") ||
		!strings.Contains(errEphemeralSuspendRequiresPersist, "--alias") {
		t.Fatalf("message = %q", errEphemeralSuspendRequiresPersist)
	}
}

func TestClearSessionOnFailureSkipsEphemeral(t *testing.T) {
	m := newTestManager(t)
	m.clearSessionOnFailure("no-such-id", false)
	// must not create a row
	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("sessions = %d, want 0", len(list))
	}
}

func TestResolvePersistSessionID_ReusesConnection(t *testing.T) {
	m := newTestManager(t)
	opts := LoginOpts{Save: true, Host: "h", Port: 22, User: "u"}
	persistLoginForTest(t, m, opts)

	id1, err := m.resolvePersistSessionID(opts, "new-id")
	if err != nil {
		t.Fatal(err)
	}
	list, _ := m.Store.ListSessions()
	if len(list) != 1 || id1 != list[0].ID {
		t.Fatalf("id1=%q list=%v", id1, list)
	}

	id2, err := m.resolvePersistSessionID(opts, "another-new")
	if err != nil {
		t.Fatal(err)
	}
	if id2 != id1 {
		t.Fatalf("id2=%q want %q", id2, id1)
	}
}

func TestResolvePersistSessionID_AliasOnExistingConnection(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Save: true, Host: "h", Port: 22, User: "u"})
	list, _ := m.Store.ListSessions()
	existingID := list[0].ID

	id, err := m.resolvePersistSessionID(LoginOpts{Alias: "prod", Host: "h", Port: 22, User: "u"}, "new-id")
	if err != nil {
		t.Fatal(err)
	}
	if id == existingID {
		t.Fatalf("alias login should create new id, got reused %q", id)
	}
}

func TestResolvePersistSessionID_SaveReusesAliasedRow(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Alias: "prod", Host: "h", Port: 22, User: "u"})

	id, err := m.resolvePersistSessionID(LoginOpts{Save: true, Host: "h", Port: 22, User: "u"}, "new-id")
	if err != nil {
		t.Fatal(err)
	}
	prod, _ := m.Store.GetSession("prod")
	if id == prod.ID {
		t.Fatalf("save without alias must not reuse aliased row %q", prod.ID)
	}
}

func TestResolvePersistSessionID_SeparateAliasesSameHost(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Alias: "prod", Host: "h", Port: 22, User: "u"})
	persistLoginForTest(t, m, LoginOpts{Alias: "staging", Host: "h", Port: 22, User: "u"})

	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("sessions=%d want 2", len(list))
	}
}

func TestResolvePersistSessionID_SaveAndAliasSeparateRows(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Save: true, Host: "h", Port: 22, User: "u"})
	persistLoginForTest(t, m, LoginOpts{Alias: "prod", Host: "h", Port: 22, User: "u"})

	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("sessions=%d want 2", len(list))
	}
}

func TestPlanLoginPersistence_NoFlagSkipsAliasedOnly(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Alias: "prod", Host: "h", Port: 22, User: "u"})

	plan, err := m.planLoginPersistence(LoginOpts{Host: "h", Port: 22, User: "u"}, "new-id")
	if err != nil {
		t.Fatal(err)
	}
	if plan.persist {
		t.Fatal("expected ephemeral plan when only aliased session exists")
	}
}

func TestResolvePersistSessionID_ReusesEmptyAliasRow(t *testing.T) {
	m := newTestManager(t)
	old := &store.Session{
		ID: "id-old", Host: "h", Port: 22, Username: "u",
		AuthExpiresAt: time.Now().Add(time.Hour).Unix(),
		Status:        store.StatusSuspended, UpdatedAt: time.Now().Add(-2 * time.Hour).Unix(),
		CreatedAt: time.Now().Add(-3 * time.Hour).Unix(),
	}
	if err := m.Store.UpsertSession(old); err != nil {
		t.Fatal(err)
	}

	id, err := m.resolvePersistSessionID(LoginOpts{Save: true, Host: "h", Port: 22, User: "u"}, "brand-new")
	if err != nil {
		t.Fatal(err)
	}
	if id != "id-old" {
		t.Fatalf("id=%q want id-old", id)
	}
}

func TestPersistLoginSession_ExtendsAuthExpiry(t *testing.T) {
	m := newTestManager(t)
	opts := LoginOpts{Save: true, Host: "example.com", Port: 22, User: "u"}
	persistLoginForTest(t, m, opts)
	s1, _ := m.Store.ListSessions()
	exp1 := s1[0].AuthExpiresAt

	time.Sleep(1100 * time.Millisecond)
	persistLoginForTest(t, m, opts)
	s2, _ := m.Store.ListSessions()
	if s2[0].AuthExpiresAt <= exp1 {
		t.Fatalf("expiry %d want > %d", s2[0].AuthExpiresAt, exp1)
	}
}

func TestResolvePersistSessionID_StoreError(t *testing.T) {
	m := newTestManager(t)
	if err := m.Store.Close(); err != nil {
		t.Fatal(err)
	}

	_, err := m.resolvePersistSessionID(LoginOpts{Save: true, Host: "h", Port: 22, User: "u"}, "new-id")
	if err == nil {
		t.Fatal("expected error from closed store")
	}

	_, err = m.planLoginPersistence(LoginOpts{Save: true, Host: "h2", Port: 22, User: "u"}, "new-id")
	if err == nil {
		t.Fatal("expected planLoginPersistence error from closed store")
	}
}

func TestLoginRunOncePersistGating(t *testing.T) {
	if shouldPersistLogin(LoginOpts{Host: "h", RemoteCommand: []string{"uptime"}}) {
		t.Fatal("one-shot without flags should not persist")
	}
	if !shouldPersistLogin(LoginOpts{Host: "h", Save: true, RemoteCommand: []string{"uptime"}}) {
		t.Fatal("one-shot with --save should persist")
	}
	if !shouldPersistLogin(LoginOpts{Host: "h", Alias: "a", RemoteCommand: []string{"uptime"}}) {
		t.Fatal("one-shot with --alias should persist")
	}

	m := newTestManager(t)
	opts := LoginOpts{Host: "h", Save: true, RemoteCommand: []string{"uptime"}, User: "u", Port: 22}
	persistLoginForTest(t, m, opts)
	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur == "" {
		t.Fatal("one-shot --save should set current session meta")
	}
}

func TestLoginUpdateExistingWithoutSave(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	firstID := list[0].ID
	firstExpiry := list[0].AuthExpiresAt

	time.Sleep(1100 * time.Millisecond)
	persistLoginForTest(t, m, LoginOpts{Host: "h", Port: 22, User: "u"})

	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("sessions=%d want 1", len(list))
	}
	if list[0].ID != firstID {
		t.Fatalf("id=%q want %q", list[0].ID, firstID)
	}
	if list[0].AuthExpiresAt <= firstExpiry {
		t.Fatalf("expiry %d want > %d", list[0].AuthExpiresAt, firstExpiry)
	}
}

func TestLoginUpdateExistingPreservesAlias(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Alias: "prod", Host: "h", Port: 22, User: "u"})
	prod, _ := m.Store.GetSession("prod")
	firstExpiry := prod.AuthExpiresAt

	persistLoginForTest(t, m, LoginOpts{Host: "h", Port: 22, User: "u"})

	sess, err := m.Store.GetSession("prod")
	if err != nil {
		t.Fatal(err)
	}
	if sess.Alias != "prod" {
		t.Fatalf("alias=%q want prod", sess.Alias)
	}
	if sess.AuthExpiresAt != firstExpiry {
		t.Fatalf("aliased row expiry changed without matching login")
	}

	list, _ := m.Store.ListSessions()
	if len(list) != 1 {
		t.Fatalf("sessions=%d want 1", len(list))
	}
}

// TestLoginWithoutTmuxClearsTmuxEnabled (TM10a) clears tmux_enabled when re-login omits --tmux.
func TestLoginWithoutTmuxClearsTmuxEnabled(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Tmux: true, Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	if !list[0].TmuxEnabled {
		t.Fatal("expected tmux enabled after initial save with --tmux")
	}

	persistLoginForTest(t, m, LoginOpts{Host: "h", Port: 22, User: "u"})

	sess, err := m.Store.GetSession(list[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if sess.TmuxEnabled {
		t.Fatal("re-login without --tmux must clear tmux_enabled for bookmark mode")
	}
}

// TestLoginWithTmuxSetsTmuxEnabled (TM10b) sets tmux_enabled when login uses --tmux.
func TestLoginWithTmuxSetsTmuxEnabled(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	if list[0].TmuxEnabled {
		t.Fatal("expected tmux disabled after save without --tmux")
	}

	persistLoginForTest(t, m, LoginOpts{Tmux: true, Host: "h", Port: 22, User: "u"})

	sess, err := m.Store.GetSession(list[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !sess.TmuxEnabled {
		t.Fatal("re-login with --tmux must set tmux_enabled")
	}
}

func TestLoginNoSaveNoExistingStillEphemeral(t *testing.T) {
	m := newTestManager(t)
	plan, err := m.planLoginPersistence(LoginOpts{Host: "newhost", Port: 22, User: "u"}, "new-id")
	if err != nil {
		t.Fatal(err)
	}
	if plan.persist {
		t.Fatal("expected ephemeral plan for unknown connection")
	}

	persistLoginForTest(t, m, LoginOpts{Host: "newhost", Port: 22, User: "u"})
	list, _ := m.Store.ListSessions()
	if len(list) != 0 {
		t.Fatalf("sessions=%d want 0", len(list))
	}
}

func TestLoginUpdateExistingSetsMeta(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Save: true, Host: "h", Port: 22, User: "u"})

	list, _ := m.Store.ListSessions()
	persistLoginForTest(t, m, LoginOpts{Host: "h", Port: 22, User: "u"})

	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur != list[0].ID {
		t.Fatalf("meta=%q want %q", cur, list[0].ID)
	}
}

func TestPlanLoginPersistenceExistingWithoutSave(t *testing.T) {
	m := newTestManager(t)
	persistLoginForTest(t, m, LoginOpts{Save: true, Host: "h", Port: 22, User: "u"})

	plan, err := m.planLoginPersistence(LoginOpts{Host: "h", Port: 22, User: "u"}, "new-id")
	if err != nil {
		t.Fatal(err)
	}
	if !plan.persist {
		t.Fatal("expected persist when connection exists")
	}
	if shouldPersistLogin(LoginOpts{Host: "h", Port: 22, User: "u"}) {
		t.Fatal("shouldPersistLogin should remain false without flags")
	}
}
