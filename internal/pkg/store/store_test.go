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

package store

import (
	"errors"
	"testing"
	"time"
)

func newSession(id, alias string) *Session {
	return &Session{
		ID: id, Alias: alias, Host: "example.com", Port: 22,
		Username: "root", AuthMethod: AuthMethodKey,
		AuthExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		Status:        StatusActive,
	}
}

func TestStoreCRUD(t *testing.T) {
	st, err := OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	if err := st.UpsertSession(newSession("test-id", "dev")); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetSession("dev")
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != "example.com" {
		t.Fatalf("host: %s", got.Host)
	}
	if err := st.SetMeta(MetaCurrentSession, "test-id"); err != nil {
		t.Fatal(err)
	}
	v, err := st.GetMeta(MetaCurrentSession)
	if err != nil || v != "test-id" {
		t.Fatalf("meta: %v %s", err, v)
	}
	if err := st.DeleteSession("test-id"); err != nil {
		t.Fatal(err)
	}
}

func TestGetSessionNotFoundIsSentinel(t *testing.T) {
	st, _ := OpenPath(":memory:")
	defer st.Close()
	_, err := st.GetSession("missing")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("got %v, want ErrSessionNotFound", err)
	}
}

func TestDeleteAllSessionsTransactional(t *testing.T) {
	st, _ := OpenPath(":memory:")
	defer st.Close()

	for i, alias := range []string{"a", "b", "c"} {
		s := newSession("id-"+alias, alias)
		_ = i
		if err := st.UpsertSession(s); err != nil {
			t.Fatal(err)
		}
		if err := st.SaveSessionState(&SessionState{
			SessionID: s.ID, Cols: 80, Rows: 24, RemoteCwd: "/tmp",
			SuspendedAt: time.Now().Unix(),
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.DeleteAllSessions(); err != nil {
		t.Fatal(err)
	}
	list, err := st.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("sessions remain: %d", len(list))
	}
	// session_state should also be empty (cascade or explicit DELETE).
	if got, _ := st.GetSessionState("id-a"); got != nil {
		t.Fatal("state row leaked")
	}
}

func TestDeleteSession_RemovesState(t *testing.T) {
	st, _ := OpenPath(":memory:")
	defer st.Close()
	s := newSession("id-1", "dev")
	if err := st.UpsertSession(s); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveSessionState(&SessionState{
		SessionID: s.ID, Cols: 80, Rows: 24,
		RemoteCwd: "/tmp", ScrollbackPath: "/tmp/x.log",
		SuspendedAt: time.Now().Unix(),
	}); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.GetSessionState("id-1"); got == nil {
		t.Fatal("state row not saved")
	}
	if err := st.DeleteSession("dev"); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.GetSessionState("id-1"); got != nil {
		t.Fatal("state row should be removed when session is deleted")
	}
}

func TestForeignKeysCascade(t *testing.T) {
	st, err := OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	s := newSession("fk-id", "fk")
	if err := st.UpsertSession(s); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveSessionState(&SessionState{
		SessionID: s.ID, Cols: 80, Rows: 24, SuspendedAt: time.Now().Unix(),
	}); err != nil {
		t.Fatal(err)
	}
	// Bypass DeleteSession's explicit state delete; rely on ON DELETE CASCADE.
	if _, err := st.db.Exec(`DELETE FROM sessions WHERE id = ?`, s.ID); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.GetSessionState(s.ID); got != nil {
		t.Fatal("session_state should cascade when session deleted via raw SQL")
	}
}

func TestDeleteSessionStateStandalone(t *testing.T) {
	st, _ := OpenPath(":memory:")
	defer st.Close()
	s := newSession("id-2", "")
	if err := st.UpsertSession(s); err != nil {
		t.Fatal(err)
	}
	if err := st.SaveSessionState(&SessionState{
		SessionID: s.ID, Cols: 1, Rows: 1, SuspendedAt: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteSessionState("id-2"); err != nil {
		t.Fatal(err)
	}
	if got, _ := st.GetSessionState("id-2"); got != nil {
		t.Fatal("state row should be gone")
	}
	// session itself must survive a state-only deletion.
	if _, err := st.GetSession("id-2"); err != nil {
		t.Fatalf("session row should still exist: %v", err)
	}
}

func TestGetSessionAmbiguous(t *testing.T) {
	st, _ := OpenPath(":memory:")
	defer st.Close()
	a := newSession("same", "other")
	b := newSession("other", "same")
	if err := st.UpsertSession(a); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertSession(b); err != nil {
		t.Fatal(err)
	}
	if _, err := st.GetSession("same"); !errors.Is(err, ErrSessionAmbiguous) {
		t.Fatalf("got %v want ErrSessionAmbiguous", err)
	}
}

func TestRefreshAuthExpiryAndUpdateStatus(t *testing.T) {
	st, _ := OpenPath(":memory:")
	defer st.Close()
	s := newSession("id-1", "x")
	if err := st.UpsertSession(s); err != nil {
		t.Fatal(err)
	}
	want := time.Now().Add(48 * time.Hour).Unix()
	if err := st.RefreshAuthExpiry("id-1", want); err != nil {
		t.Fatal(err)
	}
	if err := st.UpdateStatus("id-1", StatusSuspended); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetSession("id-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.AuthExpiresAt != want {
		t.Fatalf("auth_expires_at=%d want %d", got.AuthExpiresAt, want)
	}
	if got.Status != StatusSuspended {
		t.Fatalf("status=%s want %s", got.Status, StatusSuspended)
	}
}
