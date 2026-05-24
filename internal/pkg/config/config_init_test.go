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

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigEnsureFile_Creates(t *testing.T) {
	root := isolateConfigHome(t)

	res, err := EnsureFile(false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created {
		t.Fatal("want created")
	}
	if _, err := os.Stat(filepath.Join(root, "config.yaml")); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Session.AuthTTL == "" {
		t.Fatal("missing session.auth_ttl")
	}
	if _, ok := cfg.AI.Providers[ProviderDeepseek]; !ok {
		t.Fatal("missing deepseek provider")
	}
	if _, ok := cfg.AI.Providers[ProviderMinimax]; !ok {
		t.Fatal("missing minimax provider")
	}
}

func TestConfigEnsureFile_MergesPartial(t *testing.T) {
	root := isolateConfigHome(t)
	path := filepath.Join(root, "config.yaml")
	if err := os.WriteFile(path, []byte("session:\n  auth_ttl: 24h\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := EnsureFile(false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Created {
		t.Fatal("file already existed")
	}
	if !res.Updated {
		t.Fatal("want updated")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Session.AuthTTL != "24h" {
		t.Fatalf("auth_ttl=%q want 24h", cfg.Session.AuthTTL)
	}
	if _, ok := cfg.AI.Providers[ProviderMinimax]; !ok {
		t.Fatal("minimax should be added")
	}
}

func TestConfigEnsureFile_NoOp(t *testing.T) {
	isolateConfigHome(t)
	if _, err := EnsureFile(false); err != nil {
		t.Fatal(err)
	}

	res, err := EnsureFile(false)
	if err != nil {
		t.Fatal(err)
	}
	if res.Created || res.Updated {
		t.Fatalf("want noop, got created=%v updated=%v keys=%v", res.Created, res.Updated, res.AddedKeys)
	}
}

func TestConfigEnsureFile_DryRun(t *testing.T) {
	root := isolateConfigHome(t)
	path := filepath.Join(root, "config.yaml")

	res, err := EnsureFile(true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created {
		t.Fatal("want would-create")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("dry-run should not create config.yaml")
	}

	if err := os.WriteFile(path, []byte("session:\n  auth_ttl: 24h\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	res, err = EnsureFile(true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Updated {
		t.Fatal("want would-update")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("dry-run should not modify existing config.yaml")
	}
}

func TestConfigEnsureFile_NoRootDeepseek(t *testing.T) {
	root := isolateConfigHome(t)
	if _, err := EnsureFile(false); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "\ndeepseek:\n") || strings.HasPrefix(string(b), "deepseek:") {
		t.Fatalf("root deepseek block present: %s", string(b))
	}
}

func isolateConfigHome(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)
	t.Setenv("HOME", root)
	return root
}
