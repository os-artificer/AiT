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
	"time"

	"github.com/os-artificer/AiT/internal/pkg/terminal"
	pkgtransfer "github.com/os-artificer/AiT/internal/pkg/transfer"
	"github.com/os-artificer/AiT/internal/pkg/transfer/zmodem"
	xzmodem "github.com/os-artificer/AiT/pkg/third-party/xiwh/zmodem/zmodem"
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
	_, _ = c.scroll.Write([]byte("tail**\x18B01#"))

	var stdin bytes.Buffer
	_, cleanup, err := c.BeginZModemSession(true, &stdin, t.TempDir()+"/out", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	cleanup()
	c.drainZModemTail()

	s := c.scroll.String()
	if strings.Contains(s, "**\x18B0") || strings.Contains(s, "**\x18B01") {
		t.Fatalf("expected zmodem noise stripped from scrollback, got %q", s)
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

// TestZModemTransfer_FinishingPhase (V5) shows finishing… before wait completes.
func TestZModemTransfer_FinishingPhase(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	origWait := zmodemTransferWaitHook
	defer func() { zmodemTransferWaitHook = origWait }()
	zmodemTransferWaitHook = func(context.Context, *zmodem.Session) error {
		return nil
	}

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	c := &Client{}
	zt := &ZModemTransfer{client: c, progress: p}

	sess, err := zmodem.NewReceiveSession(io.Discard, t.TempDir()+"/out", nil)
	if err != nil {
		t.Fatal(err)
	}
	c.zmodemSession = sess

	zt.waitTransfer(t.Context())

	got := buf.String()
	if !strings.Contains(got, "finishing…") {
		t.Fatalf("missing finishing phase: %q", got)
	}
}

// TestZModemTransfer_PutWaitAfterProgress100 (V14) completes waitTransfer after send upload reaches 100%.
func TestZModemTransfer_PutWaitAfterProgress100(t *testing.T) {
	enableTransferProgressInlineTTY(t)

	path := t.TempDir() + "/payload"
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	p := NewTransferProgress(&buf)
	c := &Client{}
	zt := &ZModemTransfer{client: c, progress: p}

	progressFn := zt.zmodemProgressFn()
	sess, err := zmodem.NewSendSession(io.Discard, path, "remote.txt", progressFn)
	if err != nil {
		t.Fatal(err)
	}
	c.zmodemSession = sess
	sess.EnterSendStateForTest()

	size := int64(4)
	zfin, err := xzmodem.HexFrameBytes(xzmodem.ZFIN, xzmodem.DEFAULT_HEADER_DATA)
	if err != nil {
		t.Fatal(err)
	}

	origWait := zmodemTransferWaitHook
	defer func() { zmodemTransferWaitHook = origWait }()
	zmodemTransferWaitHook = func(ctx context.Context, s *zmodem.Session) error {
		go func() {
			time.Sleep(30 * time.Millisecond)
			s.ReportSendUploadProgressForTest(progressFn, size, size)
			_, _ = s.Feed(zfin)
		}()
		return s.Wait(ctx)
	}

	p.StartPut("./payload", "./remote.txt", size)
	if err := zt.waitTransfer(t.Context()); err != nil {
		t.Fatalf("waitTransfer: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "copying… 100%") {
		t.Fatalf("missing 100%% progress: %q", got)
	}
	if !strings.Contains(got, "finishing…") {
		t.Fatalf("missing finishing phase: %q", got)
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

func TestPTYZModemErrHint_TeardownTimeout(t *testing.T) {
	err := wrapPTYZModemErr(fmt.Errorf("zmodem get: %w", zmodem.ErrTeardownTimeout))
	s := err.Error()
	if !strings.Contains(s, "retry the transfer") {
		t.Fatalf("missing teardown retry hint: %q", s)
	}
	if !strings.Contains(s, "lrzsz") {
		t.Fatalf("missing lrzsz hint: %q", s)
	}
}
