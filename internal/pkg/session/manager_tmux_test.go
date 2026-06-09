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

	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"

	"github.com/google/uuid"
)

// TestLogin_TmuxImpliesSave (TM1) verifies --tmux alone persists a new connection with tmux_enabled.
func TestLogin_TmuxImpliesSave(t *testing.T) {
	m := newTestManager(t)
	opts := LoginOpts{Tmux: true, Host: "new-tmux-host", Port: 22, User: "u", LoginDestination: "new-tmux-host"}
	plan, err := m.planLoginPersistence(opts, uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}
	if !plan.persist {
		t.Fatal("expected persist plan for login --tmux on new host")
	}

	sess := buildSessionForLogin(opts, clientOptsFromLogin(opts), plan, loginAuthExpiry(m))
	if err := m.Store.UpsertSession(sess); err != nil {
		t.Fatal(err)
	}

	list, err := m.Store.ListSessions()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("sessions=%d want 1", len(list))
	}
	if !list[0].TmuxEnabled {
		t.Fatal("expected tmux_enabled after login --tmux")
	}
}

// TestLoad_TmuxSSHConfigHostAlias (TM8) verifies login_destination drives ssh -G destination on load.
func TestLoad_TmuxSSHConfigHostAlias(t *testing.T) {
	sess := storeSessionFixture()
	spec := loginSpecFromSession(sess, "")
	if spec.Destination != "devhost" {
		t.Fatalf("Destination = %q, want devhost", spec.Destination)
	}
	if spec.User != "deploy" {
		t.Fatalf("User = %q, want deploy", spec.User)
	}
}

// TestLoginSpecFromSession_ProxyJump (TM5) verifies stored proxy_jump is passed to re-resolve spec.
func TestLoginSpecFromSession_ProxyJump(t *testing.T) {
	s := storeSessionFixture()
	s.ProxyJump = "bastion,final"
	spec := loginSpecFromSession(s, "")
	if len(spec.ProxyJump) != 2 || spec.ProxyJump[0] != "bastion" || spec.ProxyJump[1] != "final" {
		t.Fatalf("ProxyJump = %v", spec.ProxyJump)
	}
}

func TestTmuxEnsureErr_Format(t *testing.T) {
	err := tmuxEnsureErr(storeSessionFixture(), errTestCause("tmux not found"))
	s := err.Error()
	if !strings.Contains(s, "tmux setup failed") {
		t.Fatalf("missing sentinel: %q", s)
	}
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
	if strings.Contains(s, "which tmux") {
		return
	}
	t.Fatalf("missing which tmux try: %q", s)
}

func TestTmuxRequiresTTYErr_Format(t *testing.T) {
	err := TmuxRequiresTTYErr()
	s := err.Error()
	if !strings.Contains(s, errTmuxRequiresTTY) {
		t.Fatalf("missing sentinel: %q", s)
	}
	if !strings.Contains(s, "Try:") {
		t.Fatalf("missing Try block: %q", s)
	}
}

type testCause string

func (e testCause) Error() string { return string(e) }

func errTestCause(msg string) error { return testCause(msg) }

func clientOptsFromLogin(opts LoginOpts) sshclient.Options {
	return sshclient.Options{
		Host: opts.Host, Port: opts.Port, User: opts.User,
		LoginDestination: opts.LoginDestination,
		IdentityPath:     opts.IdentityPath,
		Password:         opts.Password,
	}
}

func storeSessionFixture() *store.Session {
	return &store.Session{
		ID: "id-1", Host: "dev.example.com", Port: 22, Username: "deploy",
		LoginDestination: "devhost", TmuxEnabled: true, TmuxSession: "ait-id-1",
	}
}
