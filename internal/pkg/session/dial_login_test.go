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
	"errors"
	"testing"

	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
)

func TestShouldPromptSSHPassword(t *testing.T) {
	if shouldPromptSSHPassword(sshclient.Options{IdentityExplicit: true}, "") {
		t.Fatal("should not prompt when -i was explicit")
	}
	if !shouldPromptSSHPassword(sshclient.Options{IdentityPath: "/home/u/.ssh/id_ed25519"}, "") {
		t.Fatal("should prompt when identity from config only")
	}
	if !shouldPromptSSHPassword(sshclient.Options{}, "") {
		t.Fatal("should prompt when no identity and no CLI password")
	}
	if shouldPromptSSHPassword(sshclient.Options{}, "secret") {
		t.Fatal("should not prompt when CLI password provided")
	}
	if shouldPromptSSHPassword(sshclient.Options{BatchMode: true}, "") {
		t.Fatal("should not prompt in batch mode")
	}
}

func TestIsSSHAuthFailure(t *testing.T) {
	if !isSSHAuthFailure(errors.New("ssh: handshake failed: ssh: unable to authenticate")) {
		t.Fatal("want auth failure")
	}
	if isSSHAuthFailure(errors.New("knownhosts: key is unknown")) {
		t.Fatal("host key unknown should not be auth failure")
	}
	if isSSHAuthFailure(errors.New("host key for x has changed")) {
		t.Fatal("host key changed should not be auth failure")
	}
	if isSSHAuthFailure(errors.New("host key verification failed")) {
		t.Fatal("user declined host key should not trigger password prompt")
	}
}
