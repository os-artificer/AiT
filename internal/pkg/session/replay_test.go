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

package session

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/store"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestReplayScrollback_StripsSetupLeak (N19) sanitizes polluted scrollback files on replay.
func TestReplayScrollback_StripsSetupLeak(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())

	dir, err := platform.ScrollbackDir()
	if err != nil {
		t.Fatalf("ScrollbackDir: %v", err)
	}

	const id = "abc-123"
	name := terminal.ScrollbackFileName(id)
	polluted := strings.Join([]string{
		"Welcome to TencentOS Server 3 x86_64",
		"stty -echo 2>/dev/null; set +o history 2>/dev/null",
		`AITPWDEND__.';  printf '';  stty echo 2>/dev/null`,
		"hello world",
	}, "\r\n")
	if err := os.WriteFile(filepath.Join(dir, name), []byte(polluted), 0o600); err != nil {
		t.Fatalf("write scrollback: %v", err)
	}

	sess := &store.Session{ID: id, Username: "u", Host: "h"}
	st := &store.SessionState{SessionID: id, ScrollbackPath: name}

	m := &Manager{}
	data := m.scrollbackDataForReplay(sess, st)
	if data == "" {
		t.Fatal("scrollbackDataForReplay should return sanitized data")
	}

	for _, leak := range []string{"stty -echo", "AITPWDEND", "set +o history"} {
		if strings.Contains(data, leak) {
			t.Errorf("sanitized scrollback should not contain %q, got %q", leak, data)
		}
	}
	for _, keep := range []string{"Welcome to TencentOS", "hello world"} {
		if !strings.Contains(data, keep) {
			t.Errorf("sanitized scrollback should keep %q, got %q", keep, data)
		}
	}
}
