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

package hostkeys

import (
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestParseKnownHostsFiles(t *testing.T) {
	files := parseKnownHostsFiles("/root/.ssh/known_hosts /root/.ssh/known_hosts2")
	if len(files) != 2 || files[0] != "/root/.ssh/known_hosts" || files[1] != "/root/.ssh/known_hosts2" {
		t.Fatalf("got %v", files)
	}
	if primaryKnownHostsFile(files) != "/root/.ssh/known_hosts" {
		t.Fatal(primaryKnownHostsFile(files))
	}
}

func TestAppendKnownHostWritesNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "known_hosts")
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := appendKnownHost(path, "example.com", 2222, signer.PublicKey()); err != nil {
		t.Fatal(err)
	}
	if err := appendKnownHost(path, "other.com", 22, signer.PublicKey()); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %q", len(lines), data)
	}
	if !strings.HasPrefix(lines[0], "[example.com]:2222") {
		t.Fatalf("line0: %q", lines[0])
	}
}

func TestOpenKnownHostsDBSkipsMissing(t *testing.T) {
	dir := t.TempDir()
	primary := filepath.Join(dir, "known_hosts")
	if err := os.WriteFile(primary, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "known_hosts2")
	_, err := openKnownHostsDB([]string{primary, missing})
	if err != nil {
		t.Fatalf("openKnownHostsDB: %v", err)
	}
}
