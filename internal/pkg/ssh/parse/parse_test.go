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

package parse

import (
	"reflect"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestParseSSHStyle(t *testing.T) {
	spec, err := Parse([]string{"-p", "22", "-i", "~/.ssh/id_ed25519", "deploy@10.0.0.5"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Port != 22 || spec.Identity != "~/.ssh/id_ed25519" || spec.Destination != "deploy@10.0.0.5" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseFlagsAfterDestination(t *testing.T) {
	spec, err := Parse([]string{"-p", "22", "root@remote.example.com", "-i", "~/.ssh/id_ed25519"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Port != 22 || spec.Identity != "~/.ssh/id_ed25519" || spec.Destination != "root@remote.example.com" {
		t.Fatalf("spec=%+v", spec)
	}
	if len(spec.RemoteCommand) != 0 {
		t.Fatalf("RemoteCommand=%v want empty", spec.RemoteCommand)
	}
}

func TestParseFlagsAnyOrder(t *testing.T) {
	cases := [][]string{
		{"-p", "22", "root@h", "-i", "~/.ssh/key"},
		{"-i", "~/.ssh/key", "root@h", "-p", "22"},
		{"root@h", "-p", "22", "-i", "~/.ssh/key"},
		{"-p22", "root@h", "-i~/.ssh/key"},
	}
	for _, argv := range cases {
		spec, err := Parse(argv)
		if err != nil {
			t.Fatalf("%v: %v", argv, err)
		}
		if spec.Port != 22 || spec.Identity != "~/.ssh/key" || spec.Destination != "root@h" {
			t.Fatalf("%v: spec=%+v", argv, spec)
		}
		if len(spec.RemoteCommand) != 0 {
			t.Fatalf("%v: RemoteCommand=%v", argv, spec.RemoteCommand)
		}
	}
}

func TestParseGluedPort(t *testing.T) {
	spec, err := Parse([]string{"-p22", "user@host"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Port != 22 || spec.Destination != "user@host" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseGluedIdentity(t *testing.T) {
	spec, err := Parse([]string{"user@host", "-i~/.ssh/id_ed25519"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Identity != "~/.ssh/id_ed25519" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseLegacyHost(t *testing.T) {
	spec, err := Parse([]string{"--host", "192.168.1.1", "--user", "root"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Destination != "192.168.1.1" || spec.LegacyUser != "root" || spec.User != "root" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseRemoteCommand(t *testing.T) {
	spec, err := Parse([]string{"user@host", "uptime", "-p"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"uptime", "-p"}
	if !reflect.DeepEqual(spec.RemoteCommand, want) {
		t.Fatalf("RemoteCommand=%v want %v", spec.RemoteCommand, want)
	}
	if spec.Port != 0 {
		t.Fatalf("Port=%d want 0 (flag after command is not login -p)", spec.Port)
	}
}

func TestParseRemoteCommandWithPriorFlag(t *testing.T) {
	spec, err := Parse([]string{"-p", "22", "user@host", "uptime"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Port != 22 || spec.Destination != "user@host" {
		t.Fatalf("spec=%+v", spec)
	}
	if !reflect.DeepEqual(spec.RemoteCommand, []string{"uptime"}) {
		t.Fatalf("RemoteCommand=%v", spec.RemoteCommand)
	}
}

func TestParseRemoteCommandMultiArg(t *testing.T) {
	spec, err := Parse([]string{"user@host", "cmd1", "cmd2"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(spec.RemoteCommand, []string{"cmd1", "cmd2"}) {
		t.Fatalf("RemoteCommand=%v", spec.RemoteCommand)
	}
}

func TestParseUnsupportedFlag(t *testing.T) {
	_, err := Parse([]string{"-L", "8080:localhost:80", "host"})
	if err == nil {
		t.Fatal("expected error for -L")
	}
}

func TestParseUnsupportedFlag_NamesAisshNotLogin(t *testing.T) {
	_, err := Parse([]string{"-L", "8080:localhost:80", "host"})
	if err == nil {
		t.Fatal("expected error for -L")
	}
	msg := err.Error()
	if strings.Contains(msg, "login") {
		t.Fatalf("unsupported-flag hint must not mention login: %q", msg)
	}
	if !strings.Contains(msg, "not supported by "+toolname.AISsh) {
		t.Fatalf("want aissh in unsupported hint, got %q", msg)
	}
}

func TestParseUnknownFlagHint_ConnectHelp(t *testing.T) {
	_, err := Parse([]string{"-Z", "user@host"})
	if err == nil {
		t.Fatal("expected error for -Z")
	}
	msg := err.Error()
	if strings.Contains(msg, "login") {
		t.Fatalf("unknown-flag hint must not mention login: %q", msg)
	}
	if !strings.Contains(msg, clihints.TryConnectHelp) {
		t.Fatalf("missing connect help try line: %q", msg)
	}
	if !strings.Contains(msg, "Try:") {
		t.Fatalf("missing Try block: %q", msg)
	}
}

func TestParseVerboseCluster(t *testing.T) {
	spec, err := Parse([]string{"-vvv", "user@host"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Destination != "user@host" {
		t.Fatalf("spec=%+v", spec)
	}

	spec, err = Parse([]string{"-vq", "user@host"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Destination != "user@host" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseDestination(t *testing.T) {
	tests := []struct {
		dest       string
		user, host string
		port       int
	}{
		{"user@example.com", "user", "example.com", 0},
		{"example.com:22", "", "example.com", 22},
		{"ssh://bob@host:44/path", "bob", "host", 44},
		{"user@[2001:db8::1]:22", "user", "2001:db8::1", 22},
	}
	for _, tc := range tests {
		u, h, p, err := ParseDestination(tc.dest)
		if err != nil {
			t.Fatalf("%q: %v", tc.dest, err)
		}
		if u != tc.user || h != tc.host || p != tc.port {
			t.Fatalf("%q: got user=%q host=%q port=%d", tc.dest, u, h, p)
		}
	}
}

func TestParseAliasTmux(t *testing.T) {
	spec, err := Parse([]string{"--alias", "prod", "--tmux", "myhost"})
	if err != nil {
		t.Fatal(err)
	}
	if spec.Alias != "prod" || !spec.Tmux || spec.Destination != "myhost" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseSave(t *testing.T) {
	spec, err := Parse([]string{"--save", "-p", "2222", "deploy@host"})
	if err != nil {
		t.Fatal(err)
	}
	if !spec.Save || spec.Port != 2222 || spec.Destination != "deploy@host" {
		t.Fatalf("spec=%+v", spec)
	}
}

func TestParseSaveWithAlias(t *testing.T) {
	spec, err := Parse([]string{"--save", "--alias", "prod", "myhost"})
	if err != nil {
		t.Fatal(err)
	}
	if !spec.Save || spec.Alias != "prod" {
		t.Fatalf("spec=%+v", spec)
	}
}
