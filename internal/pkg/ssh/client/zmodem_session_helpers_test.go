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

func TestStripZModemControlRunsString_DropsRuns(t *testing.T) {
	got := stripZModemControlRunsString("ok*\x18\x01\x02#")
	if got != "ok*#" {
		t.Fatalf("got %q", got)
	}
}

func TestStripZModemControlRunsString_KeepsSingleControl(t *testing.T) {
	got := stripZModemControlRunsString("a\x01b")
	if got != "a\x01b" {
		t.Fatalf("got %q", got)
	}
}

func TestDrainZModemTail_SuppressesStdout(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
		w.Close()
		_, _ = io.Copy(io.Discard, r)
		r.Close()
	}()

	origDrain := zmodemDrainWaitHook
	zmodemDrainWaitHook = func(c *Client, drainAt time.Time) {
		c.touchRemoteOut()
		c.handleRemoteOutput([]byte("**\x18B0"))
	}
	defer func() { zmodemDrainWaitHook = origDrain }()

	c := &Client{scroll: &terminal.Buffer{}}
	c.pauseRemoteOut.Store(true)
	c.drainZModemTail()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	_, _ = io.Copy(&stdout, r)
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q want empty while paused", stdout.String())
	}
	if strings.Contains(c.scroll.String(), "**\x18B0") {
		t.Fatalf("scrollback still contains protocol noise: %q", c.scroll.String())
	}
}

func TestHandleRemoteOutput_AfterZModemDrain(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	defer func() {
		os.Stdout = oldStdout
		w.Close()
		_, _ = io.Copy(io.Discard, r)
		r.Close()
	}()

	origDrain := zmodemDrainWaitHook
	zmodemDrainWaitHook = func(*Client, time.Time) {}
	defer func() { zmodemDrainWaitHook = origDrain }()

	c := &Client{scroll: &terminal.Buffer{}}
	c.pauseRemoteOut.Store(true)
	c.handleRemoteOutput([]byte("tail**\x18B01#"))
	c.drainZModemTail()
	c.pauseRemoteOut.Store(false)
	c.handleRemoteOutput([]byte("prompt> "))

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(r)
	if strings.Contains(string(got), "**\x18B01") {
		t.Fatalf("stdout leaked protocol noise: %q", got)
	}
	if !strings.Contains(string(got), "prompt> ") {
		t.Fatalf("stdout=%q want clean prompt", got)
	}
}

func TestDrainZModemTail_FastWhenQuiet(t *testing.T) {
	origDrain := zmodemDrainWaitHook
	called := false
	zmodemDrainWaitHook = func(*Client, time.Time) { called = true }
	defer func() { zmodemDrainWaitHook = origDrain }()

	c := &Client{scroll: &terminal.Buffer{}}
	start := time.Now()
	c.drainZModemTail()
	if time.Since(start) > 400*time.Millisecond {
		t.Fatal("drainZModemTail took too long on idle PTY")
	}
	if !called {
		t.Fatal("expected drain wait hook")
	}
}
