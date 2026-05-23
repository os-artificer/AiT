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

package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/store"
)

func sessionTestEnv(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("AIT_KEYRING_BACKEND", "file")
	t.Setenv("AIT_KEYRING_DIR", dir)
	t.Setenv("AIT_KEYRING_FILE_PASSWORD", "ait-test")
}

func TestTopLevelListLoadSuspendRemoved(t *testing.T) {
	for _, args := range [][]string{
		{"list"},
		{"load", "x"},
		{"suspend"},
	} {
		err := ExecuteWriters(args, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected %v to fail", args)
		}
	}
}

func TestSessionSubcommandsRemoved(t *testing.T) {
	for _, args := range [][]string{
		{"session", "switch"},
		{"session", "delete", "x"},
	} {
		err := ExecuteWriters(args, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatalf("expected %v to fail", args)
		}
	}
}

func TestSessionCmd_NoActionFails(t *testing.T) {
	err := ExecuteWriters([]string{"session"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), errSessionActionRequired) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_MutuallyExclusiveFlags(t *testing.T) {
	err := ExecuteWriters([]string{"session", "--list", "--load", "x"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), errSessionActionExclusive) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_ListEmpty(t *testing.T) {
	sessionTestEnv(t)
	var out bytes.Buffer
	err := ExecuteWriters([]string{"session", "--list"}, &out, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "No saved sessions.") {
		t.Fatalf("out=%q", out.String())
	}
}

func TestSessionCmd_LoadNotFound(t *testing.T) {
	err := ExecuteWriters([]string{"session", "--load", "ghost"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_SuspendNoCurrent(t *testing.T) {
	sessionTestEnv(t)
	err := ExecuteWriters([]string{"session", "--suspend"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no current session") {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_SwitchNotFound(t *testing.T) {
	err := ExecuteWriters([]string{"session", "--switch", "ghost"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_SwitchPickEmpty(t *testing.T) {
	sessionTestEnv(t)
	err := ExecuteWriters([]string{"session", "--switch"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no saved sessions") {
		t.Fatalf("err=%v", err)
	}
}

func TestSessionCmd_DeleteAllForceEmptyStore(t *testing.T) {
	sessionTestEnv(t)
	err := ExecuteWriters([]string{"session", "--delete", "all", "--force"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("delete all on empty store: %v", err)
	}
}
