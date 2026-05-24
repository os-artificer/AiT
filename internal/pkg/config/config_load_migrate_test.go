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
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigLoad_MigratesLegacyConfig(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("legacy config path is Linux-only")
	}

	home := t.TempDir()
	dataRoot := filepath.Join(home, ".local", "share", "ait")
	legacyDir := filepath.Join(home, ".config", "ait")
	t.Setenv("HOME", home)
	t.Setenv("AIT_HOME", "")
	t.Setenv("AIT_DATA_DIR", "")
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))

	if err := os.MkdirAll(legacyDir, 0o700); err != nil {
		t.Fatal(err)
	}
	legacyPath := filepath.Join(legacyDir, "config.yaml")
	if err := os.WriteFile(legacyPath, []byte("session:\n  auth_ttl: 48h\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Session.AuthTTL != "48h" {
		t.Fatalf("auth_ttl=%q", cfg.Session.AuthTTL)
	}

	newPath := filepath.Join(dataRoot, "config.yaml")
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("migrated file missing: %v", err)
	}
}

func TestConfigSave_AfterMigration(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AIT_HOME", root)

	cfg := defaults()
	cfg.Session.AuthTTL = "72h"
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "config.yaml")); err != nil {
		t.Fatal(err)
	}
}
