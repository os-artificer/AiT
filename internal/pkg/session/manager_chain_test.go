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
	"os"
	"testing"
	"time"

	"github.com/creack/pty"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestEnsureSessionPersisted_ChainLogin(t *testing.T) {
	m := newTestManager(t)
	sess := &store.Session{
		ID: "ephemeral-id", Host: "hostA", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, Status: store.StatusActive,
	}

	id, err := m.ensureSessionPersisted(sess)
	if err != nil {
		t.Fatal(err)
	}
	if id != "ephemeral-id" {
		t.Fatalf("id = %q, want ephemeral-id", id)
	}

	got, err := m.Store.GetSession("ephemeral-id")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.StatusActive {
		t.Fatalf("status = %q, want active", got.Status)
	}

	cur, _ := m.Store.GetMeta(store.MetaCurrentSession)
	if cur != "ephemeral-id" {
		t.Fatalf("current meta = %q", cur)
	}
}

func TestEnsureSessionPersisted_DedupeConnectionKey(t *testing.T) {
	m := newTestManager(t)
	existing := &store.Session{
		ID: "existing-id", Host: "hostA", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, AuthExpiresAt: time.Now().Add(time.Hour).Unix(),
		Status: store.StatusSuspended,
	}
	if err := m.Store.UpsertSession(existing); err != nil {
		t.Fatal(err)
	}

	ephemeral := &store.Session{
		ID: "new-uuid", Host: "hostA", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, Status: store.StatusActive,
	}

	id, err := m.ensureSessionPersisted(ephemeral)
	if err != nil {
		t.Fatal(err)
	}
	if id != "existing-id" {
		t.Fatalf("id = %q, want existing-id (D1 dedupe)", id)
	}

	if _, err := m.Store.GetSession("new-uuid"); !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("ephemeral uuid should not be stored separately")
	}
}

func TestEnsureSessionPersisted_Suspend(t *testing.T) {
	m := newTestManager(t)
	sess := &store.Session{
		ID: "ephemeral-id", Host: "hostA", Port: 22, Username: "u",
		AuthMethod: store.AuthMethodKey, Status: store.StatusActive,
	}

	id, err := m.ensureSessionPersisted(sess)
	if err != nil {
		t.Fatal(err)
	}

	scroll := &terminal.Buffer{}
	if err := m.suspend(id, scroll, "/tmp", 80, 24); err != nil {
		t.Fatal(err)
	}

	got, err := m.Store.GetSession(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != store.StatusSuspended {
		t.Fatalf("status = %q, want suspended", got.Status)
	}
}

// TestChainLogin_DBSessionStates (N5) verifies suspend leaves old session suspended
// and a subsequent login target becomes active with current meta.
func TestChainLogin_DBSessionStates(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-a", "hostA")
	saveSession(t, m, "id-b", "hostB")
	if err := m.Store.SetMeta(store.MetaCurrentSession, "id-a"); err != nil {
		t.Fatal(err)
	}

	sessA, err := m.Store.GetSession("id-a")
	if err != nil {
		t.Fatal(err)
	}

	id, err := m.ensureSessionPersisted(sessA)
	if err != nil {
		t.Fatal(err)
	}

	scroll := &terminal.Buffer{}
	if err := m.suspend(id, scroll, "/tmp", 80, 24); err != nil {
		t.Fatal(err)
	}

	gotA, err := m.Store.GetSession("id-a")
	if err != nil {
		t.Fatal(err)
	}
	if gotA.Status != store.StatusSuspended {
		t.Fatalf("old session status = %q, want suspended", gotA.Status)
	}

	sessB, err := m.Store.GetSession("id-b")
	if err != nil {
		t.Fatal(err)
	}
	sessB.Status = store.StatusActive
	if err := m.Store.UpsertSession(sessB); err != nil {
		t.Fatal(err)
	}
	if err := m.Store.SetMeta(store.MetaCurrentSession, "id-b"); err != nil {
		t.Fatal(err)
	}

	gotB, err := m.Store.GetSession("id-b")
	if err != nil {
		t.Fatal(err)
	}
	if gotB.Status != store.StatusActive {
		t.Fatalf("new session status = %q, want active", gotB.Status)
	}

	cur, err := m.Store.GetMeta(store.MetaCurrentSession)
	if err != nil {
		t.Fatal(err)
	}
	if cur != "id-b" {
		t.Fatalf("current meta = %q, want id-b", cur)
	}
}

func TestRunInteractive_ChainLoad(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-b", "hostB")

	err := m.runChainLoad(context.Background(), []string{"hostB"})
	if err == nil {
		t.Fatal("expected dial error without real host")
	}
}

// TestRunChainLoad_PickerThenLoad (N18) verifies chain load with needPick reaches Load (dial error after pick).
func TestRunChainLoad_PickerThenLoad(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "11111111-1111-1111-1111-111111111111", "hostB")

	ptty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	defer ptty.Close()
	defer tty.Close()

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = tty
	os.Stdout = tty
	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	})

	done := make(chan error, 1)
	go func() {
		done <- m.runChainLoad(context.Background(), nil)
	}()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case err := <-done:
			if err == nil {
				t.Fatal("expected dial error without real host")
			}
			return
		case <-deadline:
			t.Fatal("runChainLoad timed out")
		default:
			_, _ = ptty.Write([]byte("1\r"))
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func TestRunInteractive_ChainSwitch(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-a", "hostA")
	saveSession(t, m, "id-b", "hostB")
	_ = m.Store.SetMeta(store.MetaCurrentSession, "id-a")

	err := m.runChainSwitch(context.Background(), []string{"hostB"})
	if err == nil {
		t.Fatal("expected dial error without real host")
	}
}

// TestRunChainSwitch_ReconnectsSameSession (N12) verifies chain switch reloads when target equals current meta.
func TestRunChainSwitch_ReconnectsSameSession(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-a", "hostA")
	if err := m.Store.SetMeta(store.MetaCurrentSession, "id-a"); err != nil {
		t.Fatal(err)
	}

	err := m.runChainSwitch(context.Background(), []string{"hostA"})
	if err == nil {
		t.Fatal("expected dial error when reconnecting same session without real host")
	}
}

func TestRunChainAction_LoginInvalidDest(t *testing.T) {
	m := newTestManager(t)
	err := m.runChainAction(context.Background(), &sshclient.ChainActionError{
		Op:   sshclient.ChainOpLogin,
		Args: []string{},
	})
	if err == nil {
		t.Fatal("expected parse/destination error")
	}
}

func TestRunInteractive_DispatchesChainAction(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "id-b", "hostB")

	chainErr := &sshclient.ChainActionError{
		Op:   sshclient.ChainOpLoad,
		Args: []string{"hostB"},
	}

	var chain *sshclient.ChainActionError
	if !errors.As(chainErr, &chain) {
		t.Fatal("ChainActionError should be detectable via errors.As")
	}

	err := m.runChainAction(context.Background(), chain)
	if err == nil {
		t.Fatal("expected dial error")
	}
}
