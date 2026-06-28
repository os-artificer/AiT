//go:build linux

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

package terminal

import (
	"os"
	"testing"
	"time"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
)

func stdinPendingBytes(fd int) (int, error) {
	return unix.IoctlGetInt(fd, unix.TIOCINQ)
}

// TestDrainStdin_DiscardsPending (N15) verifies DrainStdin clears bytes queued on a TTY.
func TestDrainStdin_DiscardsPending(t *testing.T) {
	ptty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	defer ptty.Close()
	defer tty.Close()

	oldStdin := os.Stdin
	os.Stdin = tty
	t.Cleanup(func() { os.Stdin = oldStdin })

	if _, err := ptty.Write([]byte("xy\r")); err != nil {
		t.Fatalf("write: %v", err)
	}
	time.Sleep(20 * time.Millisecond)

	pending, err := stdinPendingBytes(int(tty.Fd()))
	if err != nil {
		t.Fatalf("TIOCINQ before drain: %v", err)
	}
	if pending == 0 {
		t.Fatal("expected pending stdin bytes before drain")
	}

	DrainStdin()

	pending, err = stdinPendingBytes(int(tty.Fd()))
	if err != nil {
		t.Fatalf("TIOCINQ after drain: %v", err)
	}
	if pending != 0 {
		t.Fatalf("expected drained stdin, %d bytes still pending", pending)
	}
}
