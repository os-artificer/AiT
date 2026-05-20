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
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
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
	if err := m.List(io.Discard); err != nil {
		t.Fatalf("List: %v", err)
	}
}

func TestListTableColumns(t *testing.T) {
	m := newTestManager(t)
	now := time.Now().Unix()
	s1 := &store.Session{
		ID: "id-active", Alias: "active-one", Host: "h1", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, AuthExpiresAt: now + 100*3600,
		Status: store.StatusActive,
	}
	s2 := &store.Session{
		ID: "id-susp", Alias: "susp-one", Host: "h2", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, AuthExpiresAt: now + 200*3600,
		Status: store.StatusSuspended,
	}
	if err := m.Store.UpsertSession(s1); err != nil {
		t.Fatal(err)
	}
	if err := m.Store.UpsertSession(s2); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := m.List(&buf); err != nil {
		t.Fatalf("List: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "auth:") {
		t.Fatal("list output must not contain auth: prefix")
	}
	if !strings.Contains(out, "REMAINING") {
		t.Fatal("list output missing REMAINING header")
	}
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected header + 2 rows, got %d lines:\n%s", len(lines), out)
	}
	const rem1, rem2 = "100h0m0s", "200h0m0s"
	var col1, col2 int
	for _, line := range lines[1:] {
		switch {
		case strings.Contains(line, rem1):
			col1 = strings.Index(line, rem1)
		case strings.Contains(line, rem2):
			col2 = strings.Index(line, rem2)
		}
	}
	if col1 <= 0 || col2 <= 0 {
		t.Fatalf("missing remaining values in output:\n%s", out)
	}
	if col1 != col2 {
		t.Fatalf("REMAINING column misaligned: %d vs %d\n%s", col1, col2, out)
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
