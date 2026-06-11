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
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

// TestPickSessionFrom_CRLFWriter (N11) verifies picker list lines use CRLF when out is a CRLF writer.
func TestPickSessionFrom_CRLFWriter(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "11111111-1111-1111-1111-111111111111", "alpha")

	var buf bytes.Buffer
	out := terminal.NewCRLFWriter(&buf)
	in := strings.NewReader("1\n")

	id, err := m.PickSessionFrom(in, out)
	if err != nil {
		t.Fatalf("PickSessionFrom: %v", err)
	}
	if id != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("id = %q", id)
	}

	got := buf.String()
	if !strings.Contains(got, "alpha") {
		t.Fatalf("picker output missing session row: %q", got)
	}
	if !strings.Contains(got, "\r\n") {
		t.Fatalf("picker output should use CRLF, got %q", got)
	}
	if strings.Contains(got, "\r\r\n") {
		t.Fatalf("picker output must not double-CRLF: %q", got)
	}
}

// TestPickSessionInteractiveTTY_CRLF (N11) verifies the TTY wrapper picks on a raw PTY with CRLF output.
func TestPickSessionInteractiveTTY_CRLF(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "11111111-1111-1111-1111-111111111111", "alpha")

	ptty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	defer ptty.Close()
	defer tty.Close()

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = tty
	os.Stdout = tty
	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	})

	type pickResult struct {
		id  string
		err error
	}
	done := make(chan pickResult, 1)
	go func() {
		id, pickErr := m.PickSessionInteractiveTTY()
		done <- pickResult{id: id, err: pickErr}
	}()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case res := <-done:
			if res.err != nil {
				t.Fatalf("PickSessionInteractiveTTY: %v", res.err)
			}
			if res.id != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("id = %q", res.id)
			}
			return
		case <-deadline:
			t.Fatal("PickSessionInteractiveTTY timed out")
		default:
			_, _ = ptty.Write([]byte("1\r"))
			time.Sleep(20 * time.Millisecond)
		}
	}
}

// TestPickSessionInteractiveTTY_CtrlCCancels (N13) verifies Ctrl+C returns ErrPickCancelled with Try hint.
func TestPickSessionInteractiveTTY_CtrlCCancels(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "11111111-1111-1111-1111-111111111111", "alpha")

	ptty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	defer ptty.Close()
	defer tty.Close()

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = tty
	os.Stdout = tty
	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	})

	done := make(chan error, 1)
	go func() {
		_, pickErr := m.PickSessionInteractiveTTY()
		done <- pickErr
	}()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case pickErr := <-done:
			if !errors.Is(pickErr, ErrPickCancelled) {
				t.Fatalf("expected ErrPickCancelled, got %v", pickErr)
			}
			if !strings.Contains(pickErr.Error(), "Try:") {
				t.Fatalf("expected Try block, got %v", pickErr)
			}
			return
		case <-deadline:
			t.Fatal("PickSessionInteractiveTTY Ctrl+C timed out")
		default:
			_, _ = ptty.Write([]byte{3})
			time.Sleep(20 * time.Millisecond)
		}
	}
}

// TestPickSessionForChain_UsesRawTTY (N14) verifies pickSessionForChain delegates to the TTY wrapper.
func TestPickSessionForChain_UsesRawTTY(t *testing.T) {
	m := newTestManager(t)
	saveSession(t, m, "22222222-2222-2222-2222-222222222222", "beta")

	ptty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	defer ptty.Close()
	defer tty.Close()

	oldStdin := os.Stdin
	oldStdout := os.Stdout
	os.Stdin = tty
	os.Stdout = tty
	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	})

	done := make(chan struct {
		id  string
		err error
	}, 1)
	go func() {
		id, pickErr := m.pickSessionForChain()
		done <- struct {
			id  string
			err error
		}{id, pickErr}
	}()

	deadline := time.After(5 * time.Second)
	for {
		select {
		case res := <-done:
			if res.err != nil {
				t.Fatalf("pickSessionForChain: %v", res.err)
			}
			if res.id != "22222222-2222-2222-2222-222222222222" {
				t.Fatalf("id = %q", res.id)
			}
			return
		case <-deadline:
			t.Fatal("pickSessionForChain timed out")
		default:
			_, _ = ptty.Write([]byte("1\r"))
			time.Sleep(20 * time.Millisecond)
		}
	}
}
