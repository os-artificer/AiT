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

	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/store"
)

// TestBuildSessionForLogin_TmuxFollowsCLI (TM10d) uses opts.Tmux as the session tmux flag source of truth.
func TestBuildSessionForLogin_TmuxFollowsCLI(t *testing.T) {
	existing := &store.Session{
		ID: "id-1", TmuxEnabled: true, TmuxSession: "ait-id-1",
		Host: "h", Port: 22, Username: "u",
	}
	plan := loginPersistPlan{persist: true, id: "id-1", existing: existing}

	plain := buildSessionForLogin(
		LoginOpts{Host: "h", Port: 22, User: "u", Tmux: false},
		sshclient.Options{},
		plan,
		0,
	)
	if plain.TmuxEnabled {
		t.Fatal("login without --tmux must not reuse existing tmux_enabled")
	}
	if plain.TmuxSession != "ait-id-1" {
		t.Fatalf("tmux session name = %q want ait-id-1", plain.TmuxSession)
	}

	tmuxSess := buildSessionForLogin(
		LoginOpts{Host: "h", Port: 22, User: "u", Tmux: true},
		sshclient.Options{},
		plan,
		0,
	)
	if !tmuxSess.TmuxEnabled {
		t.Fatal("login with --tmux must set tmux_enabled")
	}
}
