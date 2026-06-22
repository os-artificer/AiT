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

package cliroute

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/store"
)

type aliasLookupStub map[string]struct{}

func (s aliasLookupStub) GetSession(idOrAlias string) (*store.Session, error) {
	if _, ok := s[idOrAlias]; ok {
		return &store.Session{Alias: idOrAlias}, nil
	}
	return nil, store.ErrSessionNotFound
}

func TestHostConfigured_MatchesHostBlock(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	content := "Host devhost prod\n  HostName example.com\n\nHost *.internal\n  User deploy\n"
	if err := os.WriteFile(cfg, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if !hostConfiguredInFile("devhost", cfg) {
		t.Fatal("devhost should match")
	}
	if !hostConfiguredInFile("prod", cfg) {
		t.Fatal("prod should match")
	}
	if !hostConfiguredInFile("app.internal", cfg) {
		t.Fatal("wildcard should match")
	}
	if hostConfiguredInFile("nosuchcmd", cfg) {
		t.Fatal("nosuchcmd should not match")
	}
}

func TestHostConfigured_HostStarIgnored(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config")
	content := "Host devhost\n  HostName example.com\n\nHost *\n  ServerAliveInterval 60\n"
	if err := os.WriteFile(cfg, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if !hostConfiguredInFile("devhost", cfg) {
		t.Fatal("devhost should match")
	}
	if hostConfiguredInFile("nosuchcmd", cfg) {
		t.Fatal("Host * must not make every bare word a config host")
	}
}

func TestShouldREPLSSHConnect_Narrow(t *testing.T) {
	SetHostConfiguredHook(func(name string) bool {
		return name == "devhost"
	})
	t.Cleanup(func() { SetHostConfiguredHook(nil) })

	if !ShouldREPLSSHConnect([]string{"devhost"}, nil) {
		t.Fatal("config host should connect")
	}
	if ShouldREPLSSHConnect([]string{"nosuchcmd"}, nil) {
		t.Fatal("unknown bare word should not connect")
	}
	if !ShouldREPLSSHConnect([]string{"-p", "22", "newhost"}, nil) {
		t.Fatal("flag-led should connect")
	}
	if !ShouldREPLSSHConnect([]string{"user@host"}, nil) {
		t.Fatal("user@host should connect")
	}
	if !ShouldREPLSSHConnect([]string{"192.168.1.1:2222"}, nil) {
		t.Fatal("host:port should connect via explicit destination parse")
	}
}

func TestIsMetaChainConnectArgs_UnknownNotChain(t *testing.T) {
	SetHostConfiguredHook(func(string) bool { return false })
	t.Cleanup(func() { SetHostConfiguredHook(nil) })
	if IsMetaChainConnectArgs([]string{"nosuch"}, nil) {
		t.Fatal("unknown meta argv must not chain login")
	}
	if !IsMetaChainConnectArgs([]string{"user@hostB"}, nil) {
		t.Fatal("user@host should chain login")
	}
}

func TestIsMetaChainConnectArgs_Alias(t *testing.T) {
	SetHostConfiguredHook(func(string) bool { return false })
	t.Cleanup(func() { SetHostConfiguredHook(nil) })

	lookup := aliasLookupStub{"prod": {}}
	if !IsMetaChainConnectArgs([]string{"prod"}, lookup) {
		t.Fatal("saved alias should chain login")
	}
	if IsMetaChainConnectArgs([]string{"prod"}, nil) {
		t.Fatal("alias without lookup must not chain when not in ssh config")
	}
}
