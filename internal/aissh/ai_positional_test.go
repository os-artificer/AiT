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

package aissh

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
)

func TestMain(m *testing.M) {
	_ = os.Setenv("AIT_KEYRING_BACKEND", "file")
	os.Exit(m.Run())
}

func setupAITestKeyring(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)
	t.Setenv("AIT_KEYRING_DIR", dir)
	if err := creds.SetProviderAPIKey(config.ProviderDeepseek, "test-key"); err != nil {
		t.Fatal(err)
	}
}

func TestAICmd_PositionalMetaNoTTY(t *testing.T) {
	setupAITestKeyring(t)
	var out, errOut bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- ExecuteWriters([]string{"ai", "model", "list"}, &out, &errOut)
	}()

	select {
	case err := <-done:
		if err != nil && !strings.Contains(err.Error(), "AI not configured") &&
			!strings.Contains(err.Error(), "API key not set") {
			t.Fatalf("unexpected error: %v stderr=%q", err, errOut.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("aitssh ai model list blocked (expected non-interactive exit)")
	}

	if out.Len() > 0 {
		s := out.String()
		if strings.Contains(s, "PROVIDER") || strings.Contains(s, "deepseek") {
			return
		}
	}
	if errOut.Len() > 0 && strings.Contains(errOut.String(), "AI not configured") {
		return
	}
	if out.Len() == 0 && errOut.Len() == 0 {
		t.Log("no output (config may be empty); command returned without blocking")
	}
}

func TestAICmd_PositionalClearNoTTY(t *testing.T) {
	setupAITestKeyring(t)
	done := make(chan error, 1)
	go func() {
		done <- ExecuteWriters([]string{"ai", "clear"}, io.Discard, io.Discard)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("aitssh ai clear blocked (expected non-interactive meta exit)")
	}
}

func isAllowedChatCLIErr(err error) bool {
	if err == nil {
		return true
	}
	s := err.Error()
	return strings.Contains(s, "AI not configured") ||
		strings.Contains(s, "API key") ||
		strings.Contains(s, "providers failed") ||
		strings.Contains(s, "Forbidden") ||
		strings.Contains(s, "connection refused") ||
		strings.Contains(s, "ask requires a message")
}

func withStdin(t *testing.T, data string, fn func()) {
	t.Helper()
	old := os.Stdin
	t.Cleanup(func() { os.Stdin = old })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	go func() {
		if data != "" {
			_, _ = io.WriteString(w, data)
		}
		_ = w.Close()
	}()
	fn()
}

func TestAICmd_ChatNoTTY(t *testing.T) {
	setupAITestKeyring(t)
	done := make(chan error, 1)
	go func() {
		done <- ExecuteWriters([]string{"ai", "ask", "ping"}, &bytes.Buffer{}, &bytes.Buffer{})
	}()

	select {
	case err := <-done:
		if !isAllowedChatCLIErr(err) {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("aitssh ai ask blocked (expected non-interactive exit)")
	}
}

func TestAICmd_ChatStdinNoTTY(t *testing.T) {
	setupAITestKeyring(t)
	withStdin(t, "ping\n", func() {
		done := make(chan error, 1)
		go func() {
			done <- ExecuteWriters([]string{"ai", "ask"}, &bytes.Buffer{}, &bytes.Buffer{})
		}()

		select {
		case err := <-done:
			if !isAllowedChatCLIErr(err) {
				t.Fatalf("unexpected error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("aitssh ai ask (stdin) blocked (expected non-interactive exit)")
		}
	})
}

func TestAICmd_ChatEmptyNoTTY(t *testing.T) {
	setupAITestKeyring(t)
	withStdin(t, "", func() {
		done := make(chan error, 1)
		go func() {
			done <- ExecuteWriters([]string{"ai", "ask"}, &bytes.Buffer{}, &bytes.Buffer{})
		}()

		select {
		case err := <-done:
			if err == nil {
				t.Fatal("expected error for empty chat message")
			}
			if !strings.Contains(err.Error(), "ask requires a message") &&
				!isAllowedChatCLIErr(err) {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(err.Error(), "Try:") {
				t.Fatalf("expected Try block: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("aitssh ai ask (empty) blocked (expected quick error)")
		}
	})
}
