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

package creds

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	"github.com/99designs/keyring"
)

func pinFileKeyringEnv(t *testing.T) string {
	t.Helper()
	t.Setenv(envBackend, "file")
	t.Setenv(envFilePassword, "ait-test")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("AIT_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	os.Unsetenv(envFileDir)

	dir, err := platform.KeyringDir()
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestSetGetDeepseekAPIKey_FileBackendDefaultDir(t *testing.T) {
	dir := pinFileKeyringEnv(t)

	const key = "sk-test-keyring-roundtrip"
	if err := SetDeepseekAPIKey(key); err != nil {
		t.Fatal(err)
	}

	got, err := GetDeepseekAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if got != key {
		t.Fatalf("got %q want %q", got, key)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected encrypted key files under keyring dir")
	}
}

func TestSessionUnlock_FileBackendDefaultDir(t *testing.T) {
	pinFileKeyringEnv(t)

	const id, secret = "sess-1", "s3cret-pass"
	if err := SetSessionUnlock(id, secret); err != nil {
		t.Fatal(err)
	}

	got, err := GetSessionUnlock(id)
	if err != nil {
		t.Fatal(err)
	}
	if got != secret {
		t.Fatalf("got %q want %q", got, secret)
	}

	if err := DeleteSessionUnlock(id); err != nil {
		t.Fatal(err)
	}
	if _, err := GetSessionUnlock(id); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestResolveFileKeyringDir_UsesPlatformKeyringDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("AIT_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	os.Unsetenv(envFileDir)

	got, err := resolveFileKeyringDir()
	if err != nil {
		t.Fatal(err)
	}

	want, err := platform.KeyringDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("resolveFileKeyringDir=%q KeyringDir=%q", got, want)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("want absolute path, got %q", got)
	}
}

func TestKeyring_MigratesLegacyService(t *testing.T) {
	dir := pinFileKeyringEnv(t)

	const id, secret = "legacy-sess", "legacy-secret"
	legacy, err := openLegacyRing()
	if err != nil {
		t.Fatal(err)
	}
	if err := legacy.Set(keyring.Item{
		Key:  sessionKey(id),
		Data: []byte(secret),
	}); err != nil {
		t.Fatal(err)
	}

	got, err := GetSessionUnlock(id)
	if err != nil {
		t.Fatal(err)
	}
	if got != secret {
		t.Fatalf("got %q want %q", got, secret)
	}

	current, err := openRing()
	if err != nil {
		t.Fatal(err)
	}
	item, err := current.Get(sessionKey(id))
	if err != nil {
		t.Fatal(err)
	}
	if string(item.Data) != secret {
		t.Fatalf("new service data=%q", item.Data)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected encrypted key files under keyring dir")
	}
}

func TestResolveFileKeyringDir_AIT_KEYRING_DIROverride(t *testing.T) {
	custom := t.TempDir()
	t.Setenv(envFileDir, custom)
	os.Unsetenv("HOME")

	got, err := resolveFileKeyringDir()
	if err != nil {
		t.Fatal(err)
	}
	if got != custom {
		t.Fatalf("got %q want %q", got, custom)
	}
}
