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

func TestSSHOptionsFromSession(t *testing.T) {
	sess := &store.Session{
		Host: "example.com", Port: 0, Username: "alice", IdentityPath: "/home/alice/.ssh/id_ed25519",
	}
	opts := sshOptionsFromSession(sess, "secret")
	if opts.Host != "example.com" || opts.Port != 22 || opts.User != "alice" {
		t.Fatalf("opts=%+v", opts)
	}
	if opts.IdentityPath != "/home/alice/.ssh/id_ed25519" || opts.Password != "secret" || !opts.UseAgent {
		t.Fatalf("auth opts=%+v", opts)
	}
}

func TestSessionHostLabel(t *testing.T) {
	if got := sessionHostLabel(&store.Session{Username: "u", Host: "h"}); got != "u@h" {
		t.Fatalf("got %q", got)
	}
	if got := sessionHostLabel(&store.Session{Host: "h"}); got != "h" {
		t.Fatalf("got %q", got)
	}
}
