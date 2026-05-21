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

package client

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestBuildAuthSkipsMissingConfigIdentity(t *testing.T) {
	dir := t.TempDir()
	sshDir := filepath.Join(dir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatal(err)
	}
	rsaPath := filepath.Join(sshDir, "id_rsa")
	if err := os.WriteFile(rsaPath, mustRSAPEM(t), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)

	missing := filepath.Join(sshDir, "id_dsa")
	methods, _, err := buildAuth(Options{
		IdentityPath:     missing,
		IdentityExplicit: false,
		UseAgent:         false,
	})
	if err != nil {
		t.Fatalf("buildAuth: %v", err)
	}
	if len(methods) == 0 {
		t.Fatal("expected auth method from default id_rsa")
	}
}

func TestBuildAuthExplicitMissingIdentity(t *testing.T) {
	_, _, err := buildAuth(Options{
		IdentityPath:     "/nonexistent/id_ed25519",
		IdentityExplicit: true,
		UseAgent:         false,
	})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("want not found error, got %v", err)
	}
}

func TestKnownHostsAddresses(t *testing.T) {
	if got := knownHostsAddresses("h.com", 22); len(got) != 1 || got[0] != "h.com" {
		t.Fatalf("port 22: %v", got)
	}
	if got := knownHostsAddresses("h.com", 60528); len(got) != 1 || got[0] != "[h.com]:60528" {
		t.Fatalf("port 60528: %v", got)
	}
}

func TestHostKeyUnknownErrorPort(t *testing.T) {
	err := hostKeyUnknownError("ylgeeker.com", 60528, "/tmp/kh")
	if err == nil || !strings.Contains(err.Error(), "-p 60528") {
		t.Fatalf("got %v", err)
	}
}

func TestAppendKnownHost(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := appendKnownHost(path, "example.com", 60528, signer.PublicKey()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "[example.com]:60528") {
		t.Fatalf("known_hosts content: %q", data)
	}
	if data[len(data)-1] != '\n' {
		t.Fatalf("known_hosts line must end with newline: %q", data)
	}
}

func mustRSAPEM(t *testing.T) []byte {
	t.Helper()
	path := writeTestRSAPrivateKey(t)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
