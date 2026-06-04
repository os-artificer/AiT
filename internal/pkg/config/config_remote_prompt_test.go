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
	"strings"
	"testing"
)

// TestConfigRemotePromptPath_Default (W7) verifies defaults() uses short mode.
func TestConfigRemotePromptPath_Default(t *testing.T) {
	cfg := defaults()
	if cfg.Session.RemotePrompt.Path != RemotePromptPathShort {
		t.Fatalf("default path=%q want %q", cfg.Session.RemotePrompt.Path, RemotePromptPathShort)
	}
	if cfg.RemotePromptPathAbsolute() {
		t.Fatal("default should not be absolute")
	}
}

// TestSetGetKey_RemotePromptPath (W9) verifies set/get for session.remote_prompt.path.
func TestSetGetKey_RemotePromptPath(t *testing.T) {
	cfg := defaults()

	if err := cfg.SetKey("session.remote_prompt.path", RemotePromptPathAbsolute); err != nil {
		t.Fatal(err)
	}
	v, err := cfg.GetKey("session.remote_prompt.path")
	if err != nil || v != RemotePromptPathAbsolute {
		t.Fatalf("GetKey absolute: v=%q err=%v", v, err)
	}
	if !cfg.RemotePromptPathAbsolute() {
		t.Fatal("RemotePromptPathAbsolute should be true")
	}

	if err := cfg.SetKey("session.remote_prompt.path", RemotePromptPathShort); err != nil {
		t.Fatal(err)
	}
	v, err = cfg.GetKey("session.remote_prompt.path")
	if err != nil || v != RemotePromptPathShort {
		t.Fatalf("GetKey short: v=%q err=%v", v, err)
	}

	if err := cfg.SetKey("session.remote_prompt.path", "full"); err == nil {
		t.Fatal("SetKey should reject invalid path mode")
	}
}

// TestConfigEnsureFile_RemotePromptPath (W7) verifies init merges session.remote_prompt.path.
func TestConfigEnsureFile_RemotePromptPath(t *testing.T) {
	root := isolateConfigHome(t)
	if _, err := EnsureFile(false); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Session.RemotePrompt.Path != RemotePromptPathShort {
		t.Fatalf("loaded path=%q want short", cfg.Session.RemotePrompt.Path)
	}

	b, err := os.ReadFile(filepath.Join(root, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "remote_prompt") {
		t.Fatalf("config.yaml missing remote_prompt: %s", string(b))
	}
}
