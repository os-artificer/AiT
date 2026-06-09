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
	"database/sql"
	"errors"
	"path/filepath"
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

func TestFindSessionByKey_NotFound(t *testing.T) {
	st, err := OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	_, err = st.FindSessionByKey("u", "host.example", 22, "")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("got %v want ErrSessionNotFound", err)
	}
}

func TestFindSessionByKey_DifferentUserOrPort(t *testing.T) {
	st, err := OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	base := newSession("id-a", "")
	base.Username = "alice"
	base.Host = "h.example"
	base.Port = 22
	if err := st.UpsertSession(base); err != nil {
		t.Fatal(err)
	}

	otherUser := newSession("id-b", "")
	otherUser.Username = "bob"
	otherUser.Host = "h.example"
	otherUser.Port = 22
	if err := st.UpsertSession(otherUser); err != nil {
		t.Fatal(err)
	}

	got, err := st.FindSessionByKey("alice", "h.example", 22, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "id-a" {
		t.Fatalf("id=%q want id-a", got.ID)
	}

	otherPort := newSession("id-c", "")
	otherPort.Username = "alice"
	otherPort.Host = "h.example"
	otherPort.Port = 2222
	if err := st.UpsertSession(otherPort); err != nil {
		t.Fatal(err)
	}

	got, err = st.FindSessionByKey("alice", "h.example", 2222, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "id-c" {
		t.Fatalf("id=%q want id-c", got.ID)
	}
}

func TestFindSessionByKey_PortZeroEquals22(t *testing.T) {
	st, err := OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	s := newSession("id-22", "")
	s.Username = "u"
	s.Host = "h"
	s.Port = 22
	if err := st.UpsertSession(s); err != nil {
		t.Fatal(err)
	}

	got, err := st.FindSessionByKey("u", "h", 22, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "id-22" {
		t.Fatalf("id=%q", got.ID)
	}
}

func TestFindSessionByKey_DifferentAliasSameConnection(t *testing.T) {
	st, err := OpenPath(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	prod := newSession("id-prod", "prod")
	prod.Username = "u"
	prod.Host = "h"
	prod.Port = 22
	staging := newSession("id-staging", "staging")
	staging.Username = "u"
	staging.Host = "h"
	staging.Port = 22
	if err := st.UpsertSession(prod); err != nil {
		t.Fatal(err)
	}
	if err := st.UpsertSession(staging); err != nil {
		t.Fatal(err)
	}

	gotProd, err := st.FindSessionByKey("u", "h", 22, "prod")
	if err != nil {
		t.Fatal(err)
	}
	if gotProd.ID != "id-prod" {
		t.Fatalf("prod id=%q", gotProd.ID)
	}

	gotStaging, err := st.FindSessionByKey("u", "h", 22, "staging")
	if err != nil {
		t.Fatal(err)
	}
	if gotStaging.ID != "id-staging" {
		t.Fatalf("staging id=%q", gotStaging.ID)
	}

	_, err = st.FindSessionByKey("u", "h", 22, "")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("empty alias lookup: got %v want ErrSessionNotFound", err)
	}
}

func TestMigrate_DedupeConnectionKeyKeepsNewest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ait.db")
	db, err := sql.Open("sqlite", buildDSN(path))
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}

	stmts := []string{
		`CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			alias TEXT UNIQUE,
			host TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 22,
			username TEXT NOT NULL,
			identity_path TEXT,
			auth_method TEXT NOT NULL,
			auth_expires_at INTEGER NOT NULL,
			tmux_enabled INTEGER NOT NULL DEFAULT 0,
			tmux_session TEXT,
			login_destination TEXT,
			proxy_jump TEXT,
			proxy_command TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			status TEXT NOT NULL
		)`,
		`CREATE TABLE session_state (
			session_id TEXT PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
			cols INTEGER, rows INTEGER, remote_cwd TEXT,
			scrollback_path TEXT, suspended_at INTEGER
		)`,
		`CREATE TABLE app_meta (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}

	now := time.Now().Unix()
	oldUpdated := now - 7200
	newUpdated := now - 60
	_, err = db.Exec(`INSERT INTO sessions (
		id, alias, host, port, username, identity_path, auth_method,
		auth_expires_at, tmux_enabled, tmux_session, created_at, updated_at, status
	) VALUES (?, NULL, ?, ?, ?, NULL, ?, ?, 0, NULL, ?, ?, ?),
	(?, NULL, ?, ?, ?, NULL, ?, ?, 0, NULL, ?, ?, ?)`,
		"id-old", "h", 22, "u", AuthMethodKey, now, oldUpdated, oldUpdated, StatusSuspended,
		"id-new", "h", 22, "u", AuthMethodKey, now, newUpdated, newUpdated, StatusSuspended,
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.dedupeConnectionKeys(); err != nil {
		t.Fatal(err)
	}
	if err := s.ensureSessionsDialColumns(); err != nil {
		t.Fatal(err)
	}
	if err := s.ensureConnectionKeyIndex(); err != nil {
		t.Fatal(err)
	}

	got, err := s.FindSessionByKey("u", "h", 22, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "id-new" {
		t.Fatalf("id=%q want id-new", got.ID)
	}

	list, err := s.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("sessions=%d want 1 after dedupe", len(list))
	}
}
