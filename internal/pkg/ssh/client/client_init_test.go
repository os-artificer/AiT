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
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
)

func TestInjectInitialCommandsSingleLine(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	done := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond)
		c.touchRemoteOut()
		close(done)
	}()

	cmds := RemotePromptInitialCommands("u", "h")
	if err := c.injectInitialCommands(stdin, cmds); err != nil {
		t.Fatal(err)
	}
	<-done
	_, _ = c.scroll.Write([]byte("AiT u@h:/home> "))
	c.finishRemoteSetup(stdin, "u", "h")

	body := stdin.String()
	if !strings.HasPrefix(body, cmds[0]) {
		t.Fatalf("stdin missing setup line: %q", body)
	}
	if strings.Contains(body, "\r") {
		t.Fatalf("stdin should not refresh when marker already buffered: %q", body)
	}
	if c.pauseRemoteOut.Load() {
		t.Fatal("finishRemoteSetup should unpause")
	}
	if c.remoteSetupSuppress.Load() {
		t.Fatal("finishRemoteSetup should clear setup suppress")
	}
}

func TestWaitRemoteOutputQuietAfterIgnoresPreInject(t *testing.T) {
	c := &Client{}
	injectAt := time.Now()
	c.lastRemoteOut.Store(injectAt.Add(-time.Second).UnixNano())

	start := time.Now()
	c.waitRemoteOutputQuietAfter(injectAt, 50*time.Millisecond, 120*time.Millisecond)
	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond || elapsed > 200*time.Millisecond {
		t.Fatalf("expected max wait ~120ms, got %v", elapsed)
	}
}

func TestWaitRemoteOutputQuietAfterPostInject(t *testing.T) {
	c := &Client{}
	injectAt := time.Now().Add(-50 * time.Millisecond)
	c.touchRemoteOut()

	start := time.Now()
	c.waitRemoteOutputQuietAfter(injectAt, 80*time.Millisecond, time.Second)
	if time.Since(start) < 80*time.Millisecond {
		t.Fatalf("returned before quiet period: %v", time.Since(start))
	}
}

// TestFinishRemoteSetup_RefreshOnly (W6) does not print a static prompt; it refreshes remote PS1.
func TestFinishRemoteSetup_RefreshOnly(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	c.finishRemoteSetup(stdin, "u", "h")

	_ = w.Close()
	os.Stdout = oldStdout
	got := string(<-done)

	if strings.Contains(got, RemotePrompt("u", "h")) {
		t.Fatalf("stdout should not contain static placeholder prompt: %q", got)
	}
	if !c.remotePromptShown.Load() {
		t.Fatal("expected remotePromptShown after paused setup")
	}
	if stdin.String() != "\r" {
		t.Fatalf("stdin want refresh \\r, got %q", stdin.String())
	}
}

func TestFinishRemoteSetupShowsPromptOnceWhenPaused(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	stdin := &bytes.Buffer{}
	c.pauseRemoteOut.Store(true)
	c.remoteSetupSuppress.Store(true)

	_, _ = c.scroll.Write([]byte("AiT u@h:/work> "))

	c.finishRemoteSetup(stdin, "u", "h")

	if stdin.Len() != 0 {
		t.Fatalf("stdin should be empty when marker already buffered, got %q", stdin.String())
	}
	if !c.remotePromptShown.Load() {
		t.Fatal("expected remotePromptShown")
	}
}

func TestCopyOutDropsDuplicatePromptAfterShown(t *testing.T) {
	c := &Client{}
	c.remotePromptShown.Store(true)
	user, host := "u", "h"
	marker := remoteSetupPromptMarker(user, host)

	if got := c.filterRemoteOutputChunk([]byte("AiT u@h:/tmp:main> "), user, host, marker); got != nil {
		t.Fatalf("duplicate prompt should be dropped, got %q", got)
	}
	if got := c.filterRemoteOutputChunk([]byte("ls\r\n"), user, host, marker); got == nil {
		t.Fatal("normal command output should pass through")
	}
}

func TestIsDuplicatePromptOutput(t *testing.T) {
	line := "AiT u@h:/data:main> "
	if !isDuplicatePromptOutput([]byte(line), "u", "h") {
		t.Fatal("exact prompt chunk should be duplicate")
	}
	if !isDuplicatePromptOutput([]byte("\r\n"+line+"\n"), "u", "h") {
		t.Fatal("prompt with surrounding whitespace should be duplicate")
	}
	if isDuplicatePromptOutput([]byte(line+"ls\n"), "u", "h") {
		t.Fatal("prompt plus command should not be duplicate")
	}
}

func TestWaitForRemoteAITPrompt(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	_, _ = c.scroll.Write([]byte("noise\nAiT u@h:/home> "))

	start := time.Now()
	c.waitForRemoteAITPrompt("u", "h", time.Second)
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("should return quickly when marker present: %v", time.Since(start))
	}
}
