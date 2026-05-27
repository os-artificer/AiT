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
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func mustRSAPEM(t *testing.T) []byte {
	t.Helper()
	path := writeTestRSAPrivateKey(t)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
