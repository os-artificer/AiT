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

package resolve

import (
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client"
	"github.com/os-artificer/AiT/internal/pkg/ssh/openssh"
)

// TestResolve_ProxyJumpHostAlias (PJ2) expands Host aliases via ssh -G into user@host:port.
func TestResolve_ProxyJumpHostAlias(t *testing.T) {
	prev := sshGOutput
	t.Cleanup(func() { sshGOutput = prev })

	sshGOutput = func(sshPath string, args []string) ([]byte, error) {
		dest := args[len(args)-1]
		switch dest {
		case "devhost":
			return []byte("hostname jump.example.com\nport 2222\nuser root\n"), nil
		case "final.example":
			return []byte("hostname final.example\nport 22\nuser root\n"), nil
		default:
			return []byte("hostname " + dest + "\nport 22\nuser root\n"), nil
		}
	}

	spec := openssh.Spec{
		Destination: "final.example",
		ProxyJump:   []string{"devhost"},
	}
	r, err := Resolve(spec)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.ProxyJump) != 1 {
		t.Fatalf("ProxyJump = %v", r.ProxyJump)
	}
	want := "root@jump.example.com:2222"
	if r.ProxyJump[0] != want {
		t.Fatalf("ProxyJump[0] = %q, want %q", r.ProxyJump[0], want)
	}
}

// TestResolveJumpDialOptions_UsesSSHG (PJ2) resolves hop Options for DialChain.
func TestResolveJumpDialOptions_UsesSSHG(t *testing.T) {
	prev := sshGOutput
	t.Cleanup(func() { sshGOutput = prev })

	sshGOutput = func(sshPath string, args []string) ([]byte, error) {
		return []byte("hostname jump.example.com\nport 2200\nuser jumpuser\n"), nil
	}

	opts, err := ResolveJumpDialOptions("devhost", client.Options{
		Host: "final.example",
		Port: 22,
		User: "root",
	})
	if err != nil {
		t.Fatal(err)
	}
	if opts.Host != "jump.example.com" || opts.Port != 2200 || opts.User != "jumpuser" {
		t.Fatalf("opts = %+v", opts)
	}
}

// TestFormatJumpDestination_UserHostPort covers jump string formatting.
func TestFormatJumpDestination_UserHostPort(t *testing.T) {
	got := formatJumpDestination("root", "dev.example", 2222)
	if got != "root@dev.example:2222" {
		t.Fatalf("got %q", got)
	}
	got = formatJumpDestination("", "bastion", 22)
	if got != "bastion:22" {
		t.Fatalf("got %q", got)
	}
}

// TestExpandOneProxyJump_CLIOverridesPort keeps explicit :port on the jump string.
func TestExpandOneProxyJump_CLIOverridesPort(t *testing.T) {
	prev := sshGOutput
	t.Cleanup(func() { sshGOutput = prev })

	sshGOutput = func(sshPath string, args []string) ([]byte, error) {
		joined := strings.Join(args, " ")
		if !strings.Contains(joined, "-p 2222") {
			t.Fatalf("expected -p 2222 in ssh -G args, got %v", args)
		}
		return []byte("hostname jump.example.com\nport 22\nuser root\n"), nil
	}

	got, err := expandOneProxyJump("ssh", openssh.Spec{}, "root@devhost:2222")
	if err != nil {
		t.Fatal(err)
	}
	if got != "root@jump.example.com:2222" {
		t.Fatalf("got %q", got)
	}
}
