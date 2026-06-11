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
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/ssh/client/prompt"
)

// TestRunInteractive_TmuxAttachBurstSuppressed (TM9a) verifies attach burst is not written while paused.
func TestRunInteractive_TmuxAttachBurstSuppressed(t *testing.T) {
	c := testClientWithPrompt("root", "devhost")
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)
	c.setupFilter = setupOutputFilter{}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	burst := []byte("Now using node v25.8.1\r\nroot@devhost:~# \r\n")
	c.handleRemoteOutput(burst)

	_ = w.Close()
	os.Stdout = oldStdout
	got := <-done

	if got != "" {
		t.Fatalf("attach burst should be suppressed while paused, got %q", got)
	}
	if !strings.Contains(c.scroll.String(), "node v25.8.1") {
		t.Fatal("burst should still land in scrollback")
	}
}

// TestHandleRemoteOutput_TmuxLateSetupLeakFiltered (TM9b) drops late setup echo after unpause grace.
func TestHandleRemoteOutput_TmuxLateSetupLeakFiltered(t *testing.T) {
	user, host := "root", "dev.os-artificer.com"
	c := testClientWithPrompt(user, host)
	c.setupLeakGraceUntil.Store(time.Now().Add(setupLeakGraceMax).UnixNano())

	leak := []byte("stty -echo 2>/dev/null; set +o history 2>/dev/null\r\n")

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	c.handleRemoteOutput(leak)

	_ = w.Close()
	os.Stdout = oldStdout
	got := <-done

	if got != "" {
		t.Fatalf("late setup leak should be filtered from stdout, got %q", got)
	}
}

func TestPartitionInitialCommands(t *testing.T) {
	user, host := "u", "h"
	setup := prompt.RemoteInitialCommands(user, host, false)
	if len(setup) != 1 {
		t.Fatalf("setup lines: %d", len(setup))
	}

	gotSetup, gotAfter := partitionInitialCommands(
		append(setup, " cd '/data/work'"),
		user,
		host,
	)
	if len(gotSetup) != 1 || gotSetup[0] != setup[0] {
		t.Fatalf("setup = %#v", gotSetup)
	}
	if len(gotAfter) != 1 || gotAfter[0] != " cd '/data/work'" {
		t.Fatalf("after = %#v", gotAfter)
	}
}
