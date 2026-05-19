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
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	// Pin the keyring to a per-test file backend so Delete/DeleteAll cannot
	// touch the host keychain or block on DBus prompts in CI.
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", t.TempDir())
	t.Setenv("AIT_KEYRING_FILE_PASSWORD", "ait-test")
	t.Setenv("AIT_SSH_INSECURE", "1")
	st, err := store.OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	cfg := &config.Config{}
	cfg.Session.AuthTTL = "168h"
	return &Manager{Store: st, Config: cfg}
}

func saveSession(t *testing.T, m *Manager, id, alias string) {
	t.Helper()
	s := &store.Session{
		ID: id, Alias: alias, Host: "h", Port: 22, Username: "u",
		AuthMethod:    store.AuthMethodKey,
		AuthExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		Status:        store.StatusActive,
	}
	if err := m.Store.UpsertSession(s); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteWithForce(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-1", "dev")
	if err := m.Delete("dev", true); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := m.Store.GetSession("dev"); !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestDeleteAllWithForce(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-1", "a")
	saveSession(t, m, "id-2", "b")
	if err := m.DeleteAll(true); err != nil {
		t.Fatalf("DeleteAll: %v", err)
	}
	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("sessions remain: %d", len(list))
	}
}

func TestListEmpty(t *testing.T) {
	m := newTestManager(t)
	if err := m.List(); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestSwitch_AlreadyOnTarget(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-1", "dev")
	if err := m.Store.SetMeta(store.MetaCurrentSession, "id-1"); err != nil {
		t.Fatal(err)
	}
	if err := m.Switch(context.Background(), "dev", false, "", false); err != nil {
		t.Fatalf("Switch should be no-op when already on target, got %v", err)
	}
}

func TestSwitch_RefusesActiveCurrent(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-1", "current")
	saveSession(t, m, "id-2", "target")
	if err := m.Store.SetMeta(store.MetaCurrentSession, "id-1"); err != nil {
		t.Fatal(err)
	}
	err := m.Switch(context.Background(), "target", false, "", false)
	if err == nil {
		t.Fatal("expected error: current session still active")
	}
}

func TestSwitch_UnknownTarget(t *testing.T) {
	m := newTestManager(t)
	err := m.Switch(context.Background(), "ghost", false, "", false)
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSwitch_StaleMetaCleared(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-target", "target")
	// Meta points at a session that was already deleted elsewhere.
	if err := m.Store.SetMeta(store.MetaCurrentSession, "id-missing"); err != nil {
		t.Fatal(err)
	}
	// Switch will fail at the eventual Dial step (no real SSH), but stale
	// meta must be cleared before that.
	_ = m.Switch(context.Background(), "target", false, "", false)
	got, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if got == "id-missing" {
		t.Fatalf("stale meta not cleared, still %q", got)
	}
}

func TestQuotePath(t *testing.T) {
	cases := map[string]string{
		"/tmp":                "'/tmp'",
		"/home/me/with space": "'/home/me/with space'",
		"/var/it's":           `'/var/it'\''s'`,
	}
	for in, want := range cases {
		if got := quotePath(in); got != want {
			t.Errorf("quotePath(%q)=%q want %q", in, got, want)
		}
	}
}

func TestNormalizePort(t *testing.T) {
	if got := normalizePort(0); got != 22 {
		t.Fatalf("normalizePort(0)=%d want 22", got)
	}
	if got := normalizePort(2222); got != 2222 {
		t.Fatalf("normalizePort(2222)=%d want 2222", got)
	}
}

func TestSuspendCurrent_PrintsInfoWithoutError(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-1", "dev")
	if err := m.Store.SetMeta(store.MetaCurrentSession, "id-1"); err != nil {
		t.Fatal(err)
	}
	if err := m.SuspendCurrent(); err != nil {
		t.Fatalf("SuspendCurrent with current session should succeed: %v", err)
	}
}

func TestSuspendCurrent_NoSessionIsError(t *testing.T) {
	m := newTestManager(t)
	if err := m.SuspendCurrent(); err == nil {
		t.Fatal("expected error when no current session")
	}
}
