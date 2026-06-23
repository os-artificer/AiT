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
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

func sessionTestEnv(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", dir)
	t.Setenv("AIT_KEYRING_FILE_PASSWORD", "ait-test")
}

func seedSessionStore(t *testing.T) {
	t.Helper()
	sessionTestEnv(t)
	err := ExecuteWriters([]string{"config", "set", "session.auth_ttl", "168h"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("config set: %v", err)
	}
	now := time.Now().Unix()
	s := &store.Session{
		ID:   "621b71ac-0000-4000-8000-000000000099",
		Host: "h", Port: 22, Username: "u",
		AuthMethod:    store.AuthMethodKey,
		AuthExpiresAt: now + 86400,
		Status:        store.StatusActive,
	}
	st, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.UpsertSession(s); err != nil {
		t.Fatal(err)
	}
}

func TestTopLevelListLoadSuspendRemoved(t *testing.T) {
	for _, args := range [][]string{
		{"list"},
		{"load", "x"},
		{"suspend"},
	} {
		err := ExecuteWriters(args, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected %v to fail", args)
		}
	}
}

func TestSessionSubcommandsRemoved(t *testing.T) {
	for _, args := range [][]string{
		{"session", "switch"},
		{"session", "delete", "x"},
	} {
		err := ExecuteWriters(args, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected %v to fail", args)
		}
	}
}

func TestSessionCmd_NoActionFails(t *testing.T) {
	err := ExecuteWriters([]string{"session"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "specify one session action") {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("expected Try block: %v", err)
	}
}

func TestSessionCmd_MutuallyExclusiveFlags(t *testing.T) {
	err := ExecuteWriters([]string{"session", "--list", "--load", "x"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "only one session action flag may be set") {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("expected Try block: %v", err)
	}
}

func TestSessionCmd_ListEmpty(t *testing.T) {
	sessionTestEnv(t)
	var out bytes.Buffer
	err := ExecuteWriters([]string{"session", "--list"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No saved sessions.") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestSessionCmd_ListNumberColumn(t *testing.T) {
	seedSessionStore(t)
	var out bytes.Buffer
	err := ExecuteWriters([]string{"session", "--list"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "#") || !strings.Contains(s, "ID") {
		t.Fatalf("missing # or ID header:\n%s", s)
	}
	if !strings.Contains(s, "621b71ac") || !strings.Contains(s, "1  621b71ac") {
		t.Fatalf("missing row number and display id:\n%s", s)
	}
}

func TestSessionCmd_ListVerboseFullID(t *testing.T) {
	seedSessionStore(t)
	var out bytes.Buffer
	err := ExecuteWriters([]string{"session", "--list", "-v"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	s := out.String()
	if !strings.Contains(s, "FULL_ID") {
		t.Fatalf("missing FULL_ID:\n%s", s)
	}
	if !strings.Contains(s, "621b71ac-0000-4000-8000-000000000099") {
		t.Fatalf("missing full uuid:\n%s", s)
	}
}

func TestSessionCmd_LoadNotFound(t *testing.T) {
	err := ExecuteWriters([]string{"session", "--load", "ghost"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_LoadDisplayIDNotFound(t *testing.T) {
	seedSessionStore(t)
	err := ExecuteWriters([]string{"session", "--load", "621b71ac"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), "display only") {
		t.Fatalf("expected display hint, got %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("expected Try block, got %v", err)
	}
}

func TestSessionCmd_LoadPickNonTTY(t *testing.T) {
	seedSessionStore(t)
	err := ExecuteWriters([]string{"session", "--load"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, session.ErrNonInteractivePick) {
		t.Fatalf("err=%v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("expected Try block, got %v", err)
	}
}

func TestSessionCmd_LoadAliasFound(t *testing.T) {
	sessionTestEnv(t)
	now := time.Now().Unix()
	st, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.UpsertSession(&store.Session{
		ID: "id-1", Alias: "prod", Host: "h", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, AuthExpiresAt: now + 86400, Status: store.StatusActive,
	}); err != nil {
		t.Fatal(err)
	}
	err = ExecuteWriters([]string{"session", "--load", "prod"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected dial error, not success")
	}
	if errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("alias prod should resolve before dial: %v", err)
	}
}

func TestSessionCmd_SuspendNoCurrent(t *testing.T) {
	sessionTestEnv(t)
	err := ExecuteWriters([]string{"session", "--suspend"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no current session") {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_SwitchNotFound(t *testing.T) {
	err := ExecuteWriters([]string{"session", "--switch", "ghost"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_SwitchPickNonTTY(t *testing.T) {
	sessionTestEnv(t)
	err := ExecuteWriters([]string{"session", "--switch"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, session.ErrNonInteractivePick) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_DeleteAllForceEmptyStore(t *testing.T) {
	sessionTestEnv(t)
	err := ExecuteWriters([]string{"session", "--delete", "all", "--force"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("delete all on empty store: %v", err)
	}
}
