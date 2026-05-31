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
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
	pkgtransfer "github.com/os-artificer/AiT/internal/pkg/transfer"
	"github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"
)

func TestZModemSession_ExclusiveIO(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	var stdin bytes.Buffer

	_, cleanup, err := c.BeginZModemSession(true, &stdin, t.TempDir()+"/out", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if !c.zmodemActive.Load() {
		t.Fatal("expected zmodem active")
	}

	c.feedZModemOutput([]byte("test payload"))
	if strings.Contains(c.scroll.String(), "test payload") {
		t.Fatal("zmodem bytes should not enter scrollback")
	}
}

func TestZModemSession_StripProtocolNoise(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	_, _ = c.scroll.Write([]byte("noise**\x18B0more"))

	var stdin bytes.Buffer
	_, cleanup, err := c.BeginZModemSession(true, &stdin, t.TempDir()+"/out", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanup()

	if strings.Contains(c.scroll.String(), "**\x18B0") {
		t.Fatal("expected zmodem noise stripped from scrollback")
	}
}

func TestHandleTransferPTY_UsesZModem(t *testing.T) {
	c := &Client{scroll: &terminal.Buffer{}}
	c.SetChainMode(true)

	origProbe := injectCaptureForRetry
	defer func() { injectCaptureForRetry = origProbe }()
	injectCaptureForRetry = func(client *Client, stdin io.Writer, cmd string) (string, error) {
		if strings.Contains(cmd, "command -v sz") {
			return "ok", nil
		}
		return "", ErrInjectCaptureNoMarkers
	}

	dir := t.TempDir()
	local := dir + "/local"
	if err := os.WriteFile(local, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}

	origWait := zmodemTransferWaitHook
	defer func() { zmodemTransferWaitHook = origWait }()
	zmodemTransferWaitHook = func(ctx context.Context, sess *zmodem.Session) error {
		return fmt.Errorf("zmodem: no protocol response from remote peer")
	}

	err := c.handleTransferPTY(t.Context(), &bytes.Buffer{}, pkgtransfer.DirPut, []string{local, "./remote"}, NewTransferProgress(io.Discard))
	if err == nil {
		t.Fatal("expected zmodem put to fail without live PTY peer")
	}
	if !strings.Contains(err.Error(), "zmodem") {
		t.Fatalf("expected zmodem error, got %v", err)
	}
}

func TestZModemTransfer_ProgressCRLF(t *testing.T) {
	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.Phase("handshake…")
	p.reportPercent(50)
	p.Done("done: get a → b (1 bytes)")

	got := buf.String()
	if !strings.Contains(got, termEOL) {
		t.Fatalf("missing CRLF progress: %q", got)
	}
	if strings.Contains(got, "transferring") {
		t.Fatalf("expected copying label: %q", got)
	}
	if !strings.Contains(got, "copying") {
		t.Fatalf("missing copying progress: %q", got)
	}
}

func TestZModemTransfer_ProgressUnifiedCopying(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	zt := &ZModemTransfer{progress: p}

	getFn := zt.zmodemProgressFn()
	putFn := zt.zmodemProgressFn()
	if getFn == nil || putFn == nil {
		t.Fatal("expected progress callbacks")
	}

	getFn(25, 100)
	putFn(50, 100)

	got := buf.String()
	if strings.Contains(got, "transferring") {
		t.Fatalf("expected copying label: %q", got)
	}
	if strings.Count(got, "copying… 25%") != 1 {
		t.Fatalf("missing get progress: %q", got)
	}
	if strings.Count(got, "copying… 50%") != 1 {
		t.Fatalf("missing put progress: %q", got)
	}
}

func TestZModemTransfer_HandshakeInlineOverwritten(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	p.StartPut("./local", "./remote", 100)
	p.PhaseInline("handshake…")

	zt := &ZModemTransfer{progress: p}
	fn := zt.zmodemProgressFn()
	for i := int64(1); i <= 5; i++ {
		fn(i*20, 100)
	}
	p.Done("done: put ./local → ./remote (100 bytes)")

	got := buf.String()
	if strings.Contains(got, "handshake…"+termEOL) {
		t.Fatalf("handshake should not be a finalized line: %q", got)
	}
	if !strings.Contains(got, "\rAiT: copying") {
		t.Fatalf("expected inline copying with \\r: %q", got)
	}
	if strings.Count(got, "copying…") != 5 {
		t.Fatalf("expected 5 copying updates, got %q", got)
	}
	if strings.Count(got, termEOL) != 2 {
		t.Fatalf("expected start+done CRLF only, got %d in %q", strings.Count(got, termEOL), got)
	}
}

func TestPTYZModemErrHint_IncludesLrzsz(t *testing.T) {
	err := wrapPTYZModemErr(ErrMissingLrzsz)
	s := err.Error()
	if !strings.Contains(s, "lrzsz") {
		t.Fatalf("missing lrzsz hint: %q", s)
	}
}
