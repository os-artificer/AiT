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

package syncpkg

import (
	"strings"
	"testing"

	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
)

func TestShellQuote(t *testing.T) {
	cases := map[string]string{
		"plain":       "'plain'",
		"with space":  "'with space'",
		"it's":        `'it'\''s'`,
		"`backticks`": "'`backticks`'",
	}
	for in, want := range cases {
		if got := shellQuote(in); got != want {
			t.Errorf("shellQuote(%q)=%q want %q", in, got, want)
		}
	}
}

func TestRshArgQuotesIdentityPath(t *testing.T) {
	got := rshArg(sshclient.Options{
		Port:         2222,
		IdentityPath: "/home/me/keys/has space/id_ed25519",
	})
	if !strings.HasPrefix(got, "ssh ") {
		t.Fatalf("rshArg missing ssh prefix: %q", got)
	}
	if !strings.Contains(got, "'/home/me/keys/has space/id_ed25519'") {
		t.Fatalf("rshArg did not shell-quote identity path: %q", got)
	}
	if !strings.Contains(got, "'-p'") || !strings.Contains(got, "'2222'") {
		t.Fatalf("rshArg port not present/quoted: %q", got)
	}
}

func TestParseRemoteDest(t *testing.T) {
	user, host, path, err := parseRemoteDest("user@host.example:/var/log/app.log")
	if err != nil || user != "user" || host != "host.example" || path != "/var/log/app.log" {
		t.Fatalf("got user=%q host=%q path=%q err=%v", user, host, path, err)
	}
	user, host, path, err = parseRemoteDest("user@[2001:db8::1]:/tmp/x")
	if err != nil || user != "user" || host != "2001:db8::1" || path != "/tmp/x" {
		t.Fatalf("ipv6: user=%q host=%q path=%q err=%v", user, host, path, err)
	}
}
